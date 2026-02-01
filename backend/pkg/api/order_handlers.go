package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/digi604/swarmmarket/backend/internal/common"
	"github.com/digi604/swarmmarket/backend/internal/transaction"
	"github.com/digi604/swarmmarket/backend/pkg/middleware"
)

// TransactionService defines the interface for transaction operations.
type TransactionService interface {
	GetTransaction(ctx context.Context, id uuid.UUID) (*transaction.Transaction, error)
	ListTransactions(ctx context.Context, params transaction.ListTransactionsParams) (*transaction.TransactionListResult, error)
	FundEscrow(ctx context.Context, transactionID, buyerID uuid.UUID) (*transaction.EscrowFundingResult, error)
	MarkDelivered(ctx context.Context, transactionID, sellerID uuid.UUID, deliveryProof, message string) (*transaction.Transaction, error)
	ConfirmDelivery(ctx context.Context, transactionID, buyerID uuid.UUID) (*transaction.Transaction, error)
	SubmitRating(ctx context.Context, transactionID, raterID uuid.UUID, req *transaction.SubmitRatingRequest) (*transaction.Rating, error)
	GetTransactionRatings(ctx context.Context, transactionID uuid.UUID) ([]*transaction.Rating, error)
	DisputeTransaction(ctx context.Context, transactionID, agentID uuid.UUID, req *transaction.DisputeRequest) (*transaction.Transaction, error)
}

// OrderHandler handles order/transaction HTTP requests.
type OrderHandler struct {
	service TransactionService
}

// NewOrderHandler creates a new order handler.
func NewOrderHandler(service TransactionService) *OrderHandler {
	return &OrderHandler{service: service}
}

// ListOrders handles GET /orders - list transactions for authenticated agent.
func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	params := transaction.ListTransactionsParams{
		AgentID: &agent.ID,
		Role:    r.URL.Query().Get("role"), // "buyer", "seller", or ""
		Limit:   parseIntParam(r, "limit", 20),
		Offset:  parseIntParam(r, "offset", 0),
	}

	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		status := transaction.TransactionStatus(statusStr)
		params.Status = &status
	}

	result, err := h.service.ListTransactions(r.Context(), params)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to list orders"))
		return
	}

	common.WriteJSON(w, http.StatusOK, result)
}

// GetOrder handles GET /orders/{id} - get transaction details.
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid order id"))
		return
	}

	tx, err := h.service.GetTransaction(r.Context(), id)
	if err != nil {
		if err == transaction.ErrTransactionNotFound {
			common.WriteError(w, http.StatusNotFound, common.ErrNotFound("order not found"))
			return
		}
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to get order"))
		return
	}

	// Check authorization - must be buyer or seller
	if tx.BuyerID != agent.ID && tx.SellerID != agent.ID {
		common.WriteError(w, http.StatusForbidden, common.ErrForbidden("not authorized to view this order"))
		return
	}

	common.WriteJSON(w, http.StatusOK, tx)
}

// FundEscrow handles POST /orders/{id}/fund - buyer initiates escrow payment.
func (h *OrderHandler) FundEscrow(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid order id"))
		return
	}

	result, err := h.service.FundEscrow(r.Context(), id, agent.ID)
	if err != nil {
		switch err {
		case transaction.ErrTransactionNotFound:
			common.WriteError(w, http.StatusNotFound, common.ErrNotFound("order not found"))
		case transaction.ErrNotAuthorized:
			common.WriteError(w, http.StatusForbidden, common.ErrForbidden("only the buyer can fund escrow"))
		case transaction.ErrInvalidStatus:
			common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("order is not in a valid state for funding"))
		default:
			common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to initiate payment: "+err.Error()))
		}
		return
	}

	common.WriteJSON(w, http.StatusOK, result)
}

// MarkDelivered handles POST /orders/{id}/deliver - seller marks order as delivered.
func (h *OrderHandler) MarkDelivered(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid order id"))
		return
	}

	var req struct {
		DeliveryProof string `json:"delivery_proof"`
		Message       string `json:"message"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	tx, err := h.service.MarkDelivered(r.Context(), id, agent.ID, req.DeliveryProof, req.Message)
	if err != nil {
		switch err {
		case transaction.ErrTransactionNotFound:
			common.WriteError(w, http.StatusNotFound, common.ErrNotFound("order not found"))
		case transaction.ErrNotAuthorized:
			common.WriteError(w, http.StatusForbidden, common.ErrForbidden("only the seller can mark as delivered"))
		case transaction.ErrInvalidStatus:
			common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("order is not in a valid state for delivery"))
		default:
			common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to mark as delivered"))
		}
		return
	}

	common.WriteJSON(w, http.StatusOK, tx)
}

// ConfirmDelivery handles POST /orders/{id}/confirm - confirm goods/services delivered.
func (h *OrderHandler) ConfirmDelivery(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid order id"))
		return
	}

	tx, err := h.service.ConfirmDelivery(r.Context(), id, agent.ID)
	if err != nil {
		switch err {
		case transaction.ErrTransactionNotFound:
			common.WriteError(w, http.StatusNotFound, common.ErrNotFound("order not found"))
		case transaction.ErrNotAuthorized:
			common.WriteError(w, http.StatusForbidden, common.ErrForbidden("only the buyer can confirm delivery"))
		case transaction.ErrInvalidStatus:
			common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("order is not in a valid state for delivery confirmation"))
		default:
			common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to confirm delivery"))
		}
		return
	}

	common.WriteJSON(w, http.StatusOK, tx)
}

// SubmitRating handles POST /orders/{id}/rating - submit a rating.
func (h *OrderHandler) SubmitRating(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid order id"))
		return
	}

	var req transaction.SubmitRatingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid request body"))
		return
	}

	rating, err := h.service.SubmitRating(r.Context(), id, agent.ID, &req)
	if err != nil {
		switch err {
		case transaction.ErrTransactionNotFound:
			common.WriteError(w, http.StatusNotFound, common.ErrNotFound("order not found"))
		case transaction.ErrNotAuthorized:
			common.WriteError(w, http.StatusForbidden, common.ErrForbidden("not authorized to rate this order"))
		case transaction.ErrInvalidRating:
			common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("rating score must be between 1 and 5"))
		case transaction.ErrRatingAlreadyExists:
			common.WriteError(w, http.StatusConflict, common.ErrConflict("you have already rated this order"))
		case transaction.ErrTransactionNotReady:
			common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("order must be delivered or completed before rating"))
		default:
			common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to submit rating"))
		}
		return
	}

	common.WriteJSON(w, http.StatusCreated, rating)
}

// GetRatings handles GET /orders/{id}/ratings - get ratings for an order.
func (h *OrderHandler) GetRatings(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid order id"))
		return
	}

	ratings, err := h.service.GetTransactionRatings(r.Context(), id)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to get ratings"))
		return
	}

	common.WriteJSON(w, http.StatusOK, map[string]any{"ratings": ratings})
}

// DisputeOrder handles POST /orders/{id}/dispute - open a dispute.
func (h *OrderHandler) DisputeOrder(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid order id"))
		return
	}

	var req transaction.DisputeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid request body"))
		return
	}

	if req.Reason == "" {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("reason is required"))
		return
	}

	tx, err := h.service.DisputeTransaction(r.Context(), id, agent.ID, &req)
	if err != nil {
		switch err {
		case transaction.ErrTransactionNotFound:
			common.WriteError(w, http.StatusNotFound, common.ErrNotFound("order not found"))
		case transaction.ErrNotAuthorized:
			common.WriteError(w, http.StatusForbidden, common.ErrForbidden("not authorized to dispute this order"))
		case transaction.ErrInvalidStatus:
			common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("order is not in a valid state for disputes"))
		default:
			common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to open dispute"))
		}
		return
	}

	common.WriteJSON(w, http.StatusOK, tx)
}
