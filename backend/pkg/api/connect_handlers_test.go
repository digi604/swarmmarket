package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/digi604/swarmmarket/backend/internal/payment"
	"github.com/digi604/swarmmarket/backend/internal/user"
	"github.com/digi604/swarmmarket/backend/pkg/middleware"
	"github.com/google/uuid"
)

func withUserContext(r *http.Request, usr *user.User) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.UserContextKey, usr)
	return r.WithContext(ctx)
}

func TestConnectHandler_Onboard_NoAuth(t *testing.T) {
	handler := NewConnectHandler(payment.NewConnectService(), nil)
	req := httptest.NewRequest("POST", "/api/v1/dashboard/connect/onboard", nil)
	rr := httptest.NewRecorder()

	handler.Onboard(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestConnectHandler_GetStatus_NoAuth(t *testing.T) {
	handler := NewConnectHandler(payment.NewConnectService(), nil)
	req := httptest.NewRequest("GET", "/api/v1/dashboard/connect/status", nil)
	rr := httptest.NewRecorder()

	handler.GetStatus(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestConnectHandler_GetStatus_NoAccount(t *testing.T) {
	handler := NewConnectHandler(payment.NewConnectService(), nil)
	usr := &user.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}

	req := httptest.NewRequest("GET", "/api/v1/dashboard/connect/status", nil)
	req = withUserContext(req, usr)
	rr := httptest.NewRecorder()

	handler.GetStatus(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["account_id"] != nil {
		t.Errorf("expected nil account_id, got %v", resp["account_id"])
	}
	if resp["charges_enabled"] != false {
		t.Errorf("expected charges_enabled=false, got %v", resp["charges_enabled"])
	}
	if resp["payouts_enabled"] != false {
		t.Errorf("expected payouts_enabled=false, got %v", resp["payouts_enabled"])
	}
	if resp["details_submitted"] != false {
		t.Errorf("expected details_submitted=false, got %v", resp["details_submitted"])
	}
}

func TestConnectHandler_CreateLoginLink_NoAuth(t *testing.T) {
	handler := NewConnectHandler(payment.NewConnectService(), nil)
	req := httptest.NewRequest("POST", "/api/v1/dashboard/connect/login-link", nil)
	rr := httptest.NewRecorder()

	handler.CreateLoginLink(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestConnectHandler_CreateLoginLink_NoAccount(t *testing.T) {
	handler := NewConnectHandler(payment.NewConnectService(), nil)
	usr := &user.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}

	req := httptest.NewRequest("POST", "/api/v1/dashboard/connect/login-link", nil)
	req = withUserContext(req, usr)
	rr := httptest.NewRecorder()

	handler.CreateLoginLink(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestConnectHandler_CreateLoginLink_HasAccount(t *testing.T) {
	// With a fake account ID, GetAccountStatus will fail at Stripe API → 500
	handler := NewConnectHandler(payment.NewConnectService(), nil)
	usr := &user.User{
		ID:                     uuid.New(),
		Email:                  "test@example.com",
		StripeConnectAccountID: "acct_123",
	}

	req := httptest.NewRequest("POST", "/api/v1/dashboard/connect/login-link", nil)
	req = withUserContext(req, usr)
	rr := httptest.NewRecorder()

	handler.CreateLoginLink(rr, req)

	// Stripe API call fails with test key → 500
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 (Stripe API failure), got %d", rr.Code)
	}
}
