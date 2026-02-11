package api

import (
	"net/http"

	"github.com/digi604/swarmmarket/backend/internal/common"
	"github.com/digi604/swarmmarket/backend/internal/payment"
	"github.com/digi604/swarmmarket/backend/internal/user"
	"github.com/digi604/swarmmarket/backend/pkg/middleware"
)

// ConnectHandler handles Stripe Connect endpoints for human users.
type ConnectHandler struct {
	connectService *payment.ConnectService
	userRepo       *user.Repository
}

// NewConnectHandler creates a new connect handler.
func NewConnectHandler(connectService *payment.ConnectService, userRepo *user.Repository) *ConnectHandler {
	return &ConnectHandler{
		connectService: connectService,
		userRepo:       userRepo,
	}
}

// Onboard starts or resumes Connect onboarding.
// POST /api/v1/dashboard/connect/onboard
func (h *ConnectHandler) Onboard(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUser(r.Context())
	if usr == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	accountID := usr.StripeConnectAccountID

	// Create new Express account if user doesn't have one yet
	if accountID == "" {
		var err error
		accountID, err = h.connectService.CreateExpressAccount(r.Context(), usr.Email)
		if err != nil {
			common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to create connect account"))
			return
		}
		if err := h.userRepo.SetStripeConnectAccountID(r.Context(), usr.ID, accountID); err != nil {
			common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to save connect account"))
			return
		}
	}

	// Generate onboarding link (works for both new and resuming users)
	refreshURL := r.Header.Get("Origin") + "/dashboard/settings?tab=billing"
	returnURL := refreshURL

	url, err := h.connectService.CreateAccountLink(r.Context(), accountID, refreshURL, returnURL)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to create onboarding link"))
		return
	}

	common.WriteJSON(w, http.StatusOK, map[string]string{
		"account_id":     accountID,
		"onboarding_url": url,
	})
}

// GetStatus returns Connect account status (queried from Stripe).
// GET /api/v1/dashboard/connect/status
func (h *ConnectHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUser(r.Context())
	if usr == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	if usr.StripeConnectAccountID == "" {
		common.WriteJSON(w, http.StatusOK, map[string]any{
			"account_id":        nil,
			"charges_enabled":   false,
			"payouts_enabled":   false,
			"details_submitted": false,
		})
		return
	}

	status, err := h.connectService.GetAccountStatus(r.Context(), usr.StripeConnectAccountID)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to get connect status"))
		return
	}

	common.WriteJSON(w, http.StatusOK, status)
}

// CreateLoginLink generates a link to the Express dashboard.
// POST /api/v1/dashboard/connect/login-link
func (h *ConnectHandler) CreateLoginLink(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUser(r.Context())
	if usr == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	if usr.StripeConnectAccountID == "" {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("connect onboarding not complete"))
		return
	}

	// Query Stripe directly — DB cache may be stale if webhook hasn't fired yet
	status, err := h.connectService.GetAccountStatus(r.Context(), usr.StripeConnectAccountID)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to check account status"))
		return
	}

	// Update DB cache opportunistically
	if h.userRepo != nil && status.ChargesEnabled != usr.StripeConnectChargesEnabled {
		_ = h.userRepo.SetStripeConnectChargesEnabled(r.Context(), usr.ID, status.ChargesEnabled)
	}

	if !status.ChargesEnabled {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("connect onboarding not complete"))
		return
	}

	url, err := h.connectService.CreateLoginLink(r.Context(), usr.StripeConnectAccountID)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to create login link"))
		return
	}

	common.WriteJSON(w, http.StatusOK, map[string]string{
		"url": url,
	})
}
