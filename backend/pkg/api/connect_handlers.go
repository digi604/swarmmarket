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

	// Re-fetch user from DB to get latest charges_enabled (webhook may have updated it after auth)
	if h.userRepo == nil {
		common.WriteError(w, http.StatusServiceUnavailable, common.ErrInternalServer("service unavailable"))
		return
	}
	freshUser, err := h.userRepo.GetUserByID(r.Context(), usr.ID)
	if err != nil || freshUser == nil {
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to load user"))
		return
	}

	if freshUser.StripeConnectAccountID == "" || !freshUser.StripeConnectChargesEnabled {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("connect onboarding not complete"))
		return
	}

	url, err := h.connectService.CreateLoginLink(r.Context(), freshUser.StripeConnectAccountID)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to create login link"))
		return
	}

	common.WriteJSON(w, http.StatusOK, map[string]string{
		"url": url,
	})
}
