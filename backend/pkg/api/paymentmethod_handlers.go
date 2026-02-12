package api

import (
	"encoding/json"
	"net/http"

	"github.com/digi604/swarmmarket/backend/internal/common"
	"github.com/digi604/swarmmarket/backend/internal/payment"
	"github.com/digi604/swarmmarket/backend/internal/spending"
	"github.com/digi604/swarmmarket/backend/internal/user"
	"github.com/digi604/swarmmarket/backend/pkg/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// PaymentMethodHandler handles payment method management for dashboard users.
type PaymentMethodHandler struct {
	paymentService *payment.Service
	userRepo       *user.Repository
}

// NewPaymentMethodHandler creates a new payment method handler.
func NewPaymentMethodHandler(paymentService *payment.Service, userRepo *user.Repository) *PaymentMethodHandler {
	return &PaymentMethodHandler{
		paymentService: paymentService,
		userRepo:       userRepo,
	}
}

// CreateSetupIntent creates a Stripe Customer (if needed) and SetupIntent.
// POST /api/v1/dashboard/payment-methods/setup
func (h *PaymentMethodHandler) CreateSetupIntent(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUser(r.Context())
	if usr == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	// Create Stripe Customer if not exists
	customerID := usr.StripeCustomerID
	if customerID == "" {
		var err error
		customerID, err = h.paymentService.CreateCustomer(r.Context(), usr.Email, usr.Name)
		if err != nil {
			common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to create customer"))
			return
		}
		if err := h.userRepo.SetStripeCustomerID(r.Context(), usr.ID, customerID); err != nil {
			common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to save customer"))
			return
		}
	}

	clientSecret, err := h.paymentService.CreateSetupIntent(r.Context(), customerID)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to create setup intent"))
		return
	}

	common.WriteJSON(w, http.StatusOK, map[string]string{
		"client_secret": clientSecret,
		"customer_id":   customerID,
	})
}

// ListPaymentMethods lists saved payment methods.
// GET /api/v1/dashboard/payment-methods
func (h *PaymentMethodHandler) ListPaymentMethods(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUser(r.Context())
	if usr == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	if usr.StripeCustomerID == "" {
		common.WriteJSON(w, http.StatusOK, map[string]any{"payment_methods": []any{}})
		return
	}

	methods, err := h.paymentService.ListPaymentMethods(r.Context(), usr.StripeCustomerID, usr.StripeDefaultPaymentMethodID)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to list payment methods"))
		return
	}
	if methods == nil {
		methods = []*payment.PaymentMethodInfo{}
	}

	common.WriteJSON(w, http.StatusOK, map[string]any{"payment_methods": methods})
}

// DeletePaymentMethod detaches a payment method.
// DELETE /api/v1/dashboard/payment-methods/{id}
func (h *PaymentMethodHandler) DeletePaymentMethod(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUser(r.Context())
	if usr == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	pmID := chi.URLParam(r, "id")
	if pmID == "" {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("payment method id required"))
		return
	}

	if err := h.paymentService.DetachPaymentMethod(r.Context(), pmID); err != nil {
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to delete payment method"))
		return
	}

	// If this was the default, clear it
	if usr.StripeDefaultPaymentMethodID == pmID {
		h.userRepo.SetStripeDefaultPaymentMethod(r.Context(), usr.StripeCustomerID, "")
	}

	w.WriteHeader(http.StatusNoContent)
}

// SetDefault sets a payment method as the default.
// PUT /api/v1/dashboard/payment-methods/{id}/default
func (h *PaymentMethodHandler) SetDefault(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUser(r.Context())
	if usr == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	pmID := chi.URLParam(r, "id")
	if pmID == "" {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("payment method id required"))
		return
	}

	if usr.StripeCustomerID == "" {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("no customer account"))
		return
	}

	if err := h.userRepo.SetStripeDefaultPaymentMethod(r.Context(), usr.StripeCustomerID, pmID); err != nil {
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to set default"))
		return
	}

	common.WriteJSON(w, http.StatusOK, map[string]string{"default_payment_method_id": pmID})
}

// SpendingLimitHandler handles spending limit endpoints for dashboard users.
type SpendingLimitHandler struct {
	spendingService *spending.Service
}

// NewSpendingLimitHandler creates a new spending limit handler.
func NewSpendingLimitHandler(spendingService *spending.Service) *SpendingLimitHandler {
	return &SpendingLimitHandler{spendingService: spendingService}
}

// GetSpendingLimits returns spending limits for an owned agent.
// GET /api/v1/dashboard/agents/{id}/spending-limits
func (h *SpendingLimitHandler) GetSpendingLimits(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUser(r.Context())
	if usr == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	agentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid agent id"))
		return
	}

	limits, err := h.spendingService.GetLimits(r.Context(), agentID)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to get spending limits"))
		return
	}

	if limits == nil {
		common.WriteJSON(w, http.StatusOK, map[string]any{"limits": nil})
		return
	}

	common.WriteJSON(w, http.StatusOK, map[string]any{"limits": limits})
}

// SetSpendingLimits sets spending limits for an owned agent.
// PUT /api/v1/dashboard/agents/{id}/spending-limits
func (h *SpendingLimitHandler) SetSpendingLimits(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUser(r.Context())
	if usr == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	agentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid agent id"))
		return
	}

	var req spending.SetSpendingLimitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid request body"))
		return
	}

	if err := h.spendingService.SetLimits(r.Context(), usr.ID, agentID, &req); err != nil {
		if err == spending.ErrNotAgentOwner {
			common.WriteError(w, http.StatusForbidden, common.ErrForbidden("not the owner of this agent"))
			return
		}
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to set spending limits"))
		return
	}

	common.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
