package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPaymentHandler_CreatePaymentIntent_NoService(t *testing.T) {
	handler := &PaymentHandler{paymentService: nil}
	req := httptest.NewRequest("POST", "/api/v1/payments/intent", nil)
	rr := httptest.NewRecorder()

	handler.CreatePaymentIntent(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rr.Code)
	}
}

func TestPaymentHandler_GetPaymentStatus_NoService(t *testing.T) {
	handler := &PaymentHandler{paymentService: nil}
	req := httptest.NewRequest("GET", "/api/v1/payments/pi_test", nil)
	rr := httptest.NewRecorder()

	handler.GetPaymentStatus(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rr.Code)
	}
}

func TestPaymentHandler_HandleWebhook_NoService(t *testing.T) {
	handler := &PaymentHandler{paymentService: nil}
	req := httptest.NewRequest("POST", "/stripe/webhook", nil)
	rr := httptest.NewRecorder()

	handler.HandleWebhook(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rr.Code)
	}
}

func TestPaymentHandler_SetUserRepo(t *testing.T) {
	handler := &PaymentHandler{}
	if handler.userRepo != nil {
		t.Error("expected nil userRepo initially")
	}

	// SetUserRepo accepts *user.Repository; we can't easily create one without a pool,
	// but we can verify the method exists and doesn't panic with nil.
	handler.SetUserRepo(nil)
	if handler.userRepo != nil {
		t.Error("expected userRepo to remain nil when set to nil")
	}
}

func TestPaymentHandler_HandleAccountUpdated_NoUserRepo(t *testing.T) {
	handler := &PaymentHandler{}
	// Should not panic when userRepo is nil
	handler.handleAccountUpdated(nil, []byte(`{"id":"acct_123","charges_enabled":true}`))
}
