package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/digi604/swarmmarket/backend/internal/common"
	"github.com/digi604/swarmmarket/backend/internal/payment"
	"github.com/digi604/swarmmarket/backend/internal/transaction"
	"github.com/digi604/swarmmarket/backend/internal/user"
	"github.com/digi604/swarmmarket/backend/pkg/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v76/webhook"
)

// webhookTimestampTolerance is the maximum age of a webhook event we'll accept.
// This helps prevent replay attacks.
const webhookTimestampTolerance = 5 * time.Minute

// PaymentHandler handles payment HTTP requests.
type PaymentHandler struct {
	paymentService     *payment.Service
	transactionService *transaction.Service
	userRepo           *user.Repository
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

// SetUserRepo sets the user repository for Connect webhook handling.
func (h *PaymentHandler) SetUserRepo(repo *user.Repository) {
	h.userRepo = repo
}

// CreatePaymentIntentRequest is the request body for creating a payment.
type CreatePaymentIntentRequest struct {
	TransactionID string `json:"transaction_id"`
}

// CreatePaymentIntent handles POST /payments/intent - create payment intent for escrow.
func (h *PaymentHandler) CreatePaymentIntent(w http.ResponseWriter, r *http.Request) {
	if h.paymentService == nil {
		common.WriteError(w, http.StatusServiceUnavailable, common.ErrServiceUnavailable("payments are not configured"))
		return
	}

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
	})
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to create payment"))
		return
	}

	common.WriteJSON(w, http.StatusCreated, result)
}

// GetPaymentStatus handles GET /payments/{paymentIntentId} - get payment status.
func (h *PaymentHandler) GetPaymentStatus(w http.ResponseWriter, r *http.Request) {
	if h.paymentService == nil {
		common.WriteError(w, http.StatusServiceUnavailable, common.ErrServiceUnavailable("payments are not configured"))
		return
	}

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
	if h.paymentService == nil {
		common.WriteError(w, http.StatusServiceUnavailable, common.ErrServiceUnavailable("payments are not configured"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("failed to read body"))
		return
	}

	// Verify webhook signature with timestamp tolerance to prevent replay attacks
	sig := r.Header.Get("Stripe-Signature")
	event, err := webhook.ConstructEventWithOptions(body, sig, h.webhookSecret, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})
	if err != nil {
		fmt.Printf("[Stripe Webhook] Signature verification error: %v\n", err)
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid signature"))
		return
	}

	// Additional timestamp check for replay attack prevention
	// Stripe's signature includes a timestamp, but we add an extra check
	eventTime := time.Unix(event.Created, 0)
	if time.Since(eventTime) > webhookTimestampTolerance {
		fmt.Printf("[Stripe Webhook] Event too old: %v\n", eventTime)
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("event expired"))
		return
	}

	// Handle different event types
	switch event.Type {
	case "payment_intent.succeeded":
		h.handlePaymentSucceeded(r.Context(), event.Data.Raw)

	case "payment_intent.payment_failed":
		h.handlePaymentFailed(r.Context(), event.Data.Raw)

	case "charge.refunded":
		h.handleRefund(r.Context(), event.Data.Raw)

	case "account.updated":
		h.handleAccountUpdated(r.Context(), event.Data.Raw)

	case "setup_intent.succeeded":
		h.handleSetupIntentSucceeded(r.Context(), event.Data.Raw)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *PaymentHandler) handlePaymentSucceeded(ctx context.Context, data []byte) {
	var pi struct {
		ID       string            `json:"id"`
		Metadata map[string]string `json:"metadata"`
		Status   string            `json:"status"`
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

	if pi.Status == "requires_capture" || pi.Status == "succeeded" {
		h.transactionService.ConfirmEscrowFunded(ctx, transactionID, pi.ID)
	}
}

func (h *PaymentHandler) handlePaymentFailed(ctx context.Context, data []byte) {
	var pi struct {
		ID               string            `json:"id"`
		Metadata         map[string]string `json:"metadata"`
		LastPaymentError *struct {
			Message string `json:"message"`
		} `json:"last_payment_error"`
	}
	if err := json.Unmarshal(data, &pi); err != nil {
		return
	}

	failureReason := "payment failed"
	if pi.LastPaymentError != nil && pi.LastPaymentError.Message != "" {
		failureReason = pi.LastPaymentError.Message
	}

	transactionIDStr, ok := pi.Metadata["transaction_id"]
	if !ok {
		return
	}

	transactionID, err := uuid.Parse(transactionIDStr)
	if err != nil {
		return
	}

	tx, err := h.transactionService.GetTransaction(ctx, transactionID)
	if err != nil {
		return
	}

	h.transactionService.PublishPaymentFailed(ctx, tx, pi.ID, failureReason)
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

func (h *PaymentHandler) handleSetupIntentSucceeded(ctx context.Context, data []byte) {
	if h.userRepo == nil {
		return
	}

	var si struct {
		Customer      string `json:"customer"`
		PaymentMethod string `json:"payment_method"`
	}
	if err := json.Unmarshal(data, &si); err != nil {
		return
	}

	if si.Customer == "" || si.PaymentMethod == "" {
		return
	}

	if err := h.userRepo.SetStripeDefaultPaymentMethod(ctx, si.Customer, si.PaymentMethod); err != nil {
		fmt.Printf("[Stripe Webhook] setup_intent.succeeded: failed to save payment method for %s: %v\n", si.Customer, err)
	}
}

func (h *PaymentHandler) handleAccountUpdated(ctx context.Context, data []byte) {
	if h.userRepo == nil {
		return
	}

	var acct struct {
		ID             string `json:"id"`
		ChargesEnabled bool   `json:"charges_enabled"`
	}
	if err := json.Unmarshal(data, &acct); err != nil {
		return
	}

	usr, err := h.userRepo.GetUserByStripeConnectAccountID(ctx, acct.ID)
	if err != nil {
		fmt.Printf("[Stripe Webhook] account.updated: user not found for %s: %v\n", acct.ID, err)
		return
	}

	if err := h.userRepo.SetStripeConnectChargesEnabled(ctx, usr.ID, acct.ChargesEnabled); err != nil {
		fmt.Printf("[Stripe Webhook] account.updated: failed to update charges_enabled for %s: %v\n", acct.ID, err)
	}
}
