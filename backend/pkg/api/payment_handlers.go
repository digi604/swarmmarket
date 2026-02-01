package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v76/webhook"
	"github.com/digi604/swarmmarket/backend/internal/common"
	"github.com/digi604/swarmmarket/backend/internal/payment"
	"github.com/digi604/swarmmarket/backend/internal/transaction"
	"github.com/digi604/swarmmarket/backend/pkg/middleware"
)

// PaymentHandler handles payment HTTP requests.
type PaymentHandler struct {
	paymentService     *payment.Service
	transactionService *transaction.Service
	webhookSecret      string
}

// NewPaymentHandler creates a new payment handler.
func NewPaymentHandler(paymentService *payment.Service, transactionService *transaction.Service, webhookSecret string) *PaymentHandler {
	return &PaymentHandler{
		paymentService:     paymentService,
		transactionService: transactionService,
		webhookSecret:      webhookSecret,
	}
}

// CreatePaymentRequest is the request body for creating a payment.
type CreatePaymentIntentRequest struct {
	TransactionID string `json:"transaction_id"`
	ReturnURL     string `json:"return_url"` // URL to redirect after payment confirmation
}

// CreatePaymentIntent handles POST /payments/intent - create payment intent for escrow.
func (h *PaymentHandler) CreatePaymentIntent(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	var req CreatePaymentIntentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid request body"))
		return
	}

	transactionID, err := uuid.Parse(req.TransactionID)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid transaction_id"))
		return
	}

	// Get transaction
	tx, err := h.transactionService.GetTransaction(r.Context(), transactionID)
	if err != nil {
		if err == transaction.ErrTransactionNotFound {
			common.WriteError(w, http.StatusNotFound, common.ErrNotFound("transaction not found"))
			return
		}
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to get transaction"))
		return
	}

	// Verify buyer is making the payment
	if tx.BuyerID != agent.ID {
		common.WriteError(w, http.StatusForbidden, common.ErrForbidden("only the buyer can create a payment"))
		return
	}

	// Create payment intent
	result, err := h.paymentService.CreateEscrowPayment(r.Context(), &payment.CreatePaymentRequest{
		TransactionID: transactionID,
		BuyerID:       tx.BuyerID,
		SellerID:      tx.SellerID,
		Amount:        tx.Amount,
		Currency:      tx.Currency,
		ReturnURL:     req.ReturnURL,
	})
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to create payment"))
		return
	}

	common.WriteJSON(w, http.StatusCreated, result)
}

// GetPaymentStatus handles GET /payments/{paymentIntentId} - get payment status.
func (h *PaymentHandler) GetPaymentStatus(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	paymentIntentID := chi.URLParam(r, "paymentIntentId")
	if paymentIntentID == "" {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("payment_intent_id is required"))
		return
	}

	status, err := h.paymentService.GetPaymentIntent(r.Context(), paymentIntentID)
	if err != nil {
		common.WriteError(w, http.StatusNotFound, common.ErrNotFound("payment not found"))
		return
	}

	common.WriteJSON(w, http.StatusOK, status)
}

// HandleWebhook handles POST /payments/webhook - Stripe webhook events.
func (h *PaymentHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("failed to read body"))
		return
	}

	// Verify webhook signature
	sig := r.Header.Get("Stripe-Signature")
	event, err := webhook.ConstructEventWithOptions(body, sig, h.webhookSecret, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})
	if err != nil {
		fmt.Printf("[Stripe Webhook] Signature verification error: %v\n", err)
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid signature"))
		return
	}

	// Handle different event types
	switch event.Type {
	case "payment_intent.succeeded":
		// Payment authorized - update escrow status
		h.handlePaymentSucceeded(r.Context(), event.Data.Raw)

	case "payment_intent.payment_failed":
		// Payment failed
		h.handlePaymentFailed(r.Context(), event.Data.Raw)

	case "charge.refunded":
		// Refund processed
		h.handleRefund(r.Context(), event.Data.Raw)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *PaymentHandler) handlePaymentSucceeded(ctx context.Context, data []byte) {
	// Parse payment intent from webhook data
	var pi struct {
		ID       string            `json:"id"`
		Metadata map[string]string `json:"metadata"`
		Status   string            `json:"status"`
	}
	if err := json.Unmarshal(data, &pi); err != nil {
		return
	}

	// Get transaction ID from metadata
	transactionIDStr, ok := pi.Metadata["transaction_id"]
	if !ok {
		return
	}

	transactionID, err := uuid.Parse(transactionIDStr)
	if err != nil {
		return
	}

	// For manual capture (escrow), "succeeded" means funds are captured
	// For automatic capture or when status is "requires_capture", mark as funded
	if pi.Status == "requires_capture" || pi.Status == "succeeded" {
		h.transactionService.ConfirmEscrowFunded(ctx, transactionID, pi.ID)
	}
}

func (h *PaymentHandler) handlePaymentFailed(ctx context.Context, data []byte) {
	var pi struct {
		ID       string            `json:"id"`
		Metadata map[string]string `json:"metadata"`
	}
	if err := json.Unmarshal(data, &pi); err != nil {
		return
	}

	transactionIDStr, ok := pi.Metadata["transaction_id"]
	if !ok {
		return
	}

	transactionID, err := uuid.Parse(transactionIDStr)
	if err != nil {
		return
	}

	// Get transaction and publish failure event
	tx, err := h.transactionService.GetTransaction(ctx, transactionID)
	if err != nil {
		return
	}

	// Publish payment failed event for notification
	// The transaction stays in "pending" - buyer can retry
	_ = tx // Would use for notification
}

func (h *PaymentHandler) handleRefund(ctx context.Context, data []byte) {
	var charge struct {
		PaymentIntent string            `json:"payment_intent"`
		Metadata      map[string]string `json:"metadata"`
	}
	if err := json.Unmarshal(data, &charge); err != nil {
		return
	}

	transactionIDStr, ok := charge.Metadata["transaction_id"]
	if !ok {
		return
	}

	transactionID, err := uuid.Parse(transactionIDStr)
	if err != nil {
		return
	}

	// Mark transaction as refunded
	h.transactionService.RefundTransaction(ctx, transactionID)
}
