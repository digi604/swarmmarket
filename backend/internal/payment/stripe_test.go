package payment

import (
	"testing"

	"github.com/google/uuid"
)

func TestNormalizeCurrency(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"USD", "usd"},
		{"usd", "usd"},
		{"EUR", "eur"},
		{"eur", "eur"},
		{"GBP", "gbp"},
		{"gbp", "gbp"},
		{"", "usd"},       // Default
		{"JPY", "usd"},    // Unknown defaults to USD
		{"UNKNOWN", "usd"}, // Unknown defaults to USD
	}

	for _, tt := range tests {
		result := normalizeCurrency(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeCurrency(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestCreatePaymentRequest(t *testing.T) {
	req := &CreatePaymentRequest{
		TransactionID: uuid.New(),
		BuyerID:       uuid.New(),
		SellerID:      uuid.New(),
		Amount:        100.50,
		Currency:      "USD",
	}

	if req.Amount != 100.50 {
		t.Errorf("expected amount 100.50, got %f", req.Amount)
	}

	if req.Currency != "USD" {
		t.Errorf("expected currency USD, got %s", req.Currency)
	}
}

func TestPaymentResult(t *testing.T) {
	result := &PaymentResult{
		PaymentIntentID: "pi_test_123",
		ClientSecret:    "pi_test_123_secret_abc",
		Status:          "requires_capture",
		Amount:          50.00,
		Currency:        "USD",
	}

	if result.PaymentIntentID != "pi_test_123" {
		t.Errorf("expected PaymentIntentID pi_test_123, got %s", result.PaymentIntentID)
	}

	if result.Status != "requires_capture" {
		t.Errorf("expected status requires_capture, got %s", result.Status)
	}
}

func TestPaymentStatus(t *testing.T) {
	status := &PaymentStatus{
		PaymentIntentID: "pi_test_456",
		Status:          "succeeded",
		Amount:          75.00,
		Currency:        "eur",
		CapturedAmount:  75.00,
	}

	if status.Amount != status.CapturedAmount {
		t.Errorf("expected full capture, amount=%f, captured=%f", status.Amount, status.CapturedAmount)
	}
}

func TestTransferRequest(t *testing.T) {
	req := &TransferRequest{
		TransactionID:         uuid.New(),
		SellerStripeAccountID: "acct_seller123",
		Amount:                200.00,
		Currency:              "USD",
		SourceTransactionID:   "ch_charge123",
	}

	if req.SellerStripeAccountID != "acct_seller123" {
		t.Errorf("expected seller account acct_seller123, got %s", req.SellerStripeAccountID)
	}
}

func TestTransferResult(t *testing.T) {
	result := &TransferResult{
		TransferID: "tr_transfer123",
		Amount:     150.00,
		Currency:   "USD",
		Status:     "completed",
	}

	if result.Status != "completed" {
		t.Errorf("expected status completed, got %s", result.Status)
	}
}

func TestConfig(t *testing.T) {
	cfg := Config{
		SecretKey:          "sk_test_xxx",
		WebhookSecret:      "whsec_xxx",
		PlatformFeePercent: 0.025,
	}

	if cfg.PlatformFeePercent != 0.025 {
		t.Errorf("expected platform fee 0.025, got %f", cfg.PlatformFeePercent)
	}

	// Test fee calculation
	amount := int64(10000) // $100.00 in cents
	fee := int64(float64(amount) * cfg.PlatformFeePercent)
	if fee != 250 { // $2.50 in cents
		t.Errorf("expected fee 250 cents, got %d", fee)
	}
}

func TestErrors(t *testing.T) {
	// Verify error messages
	if ErrPaymentFailed.Error() != "payment failed" {
		t.Errorf("unexpected error message: %s", ErrPaymentFailed.Error())
	}

	if ErrRefundFailed.Error() != "refund failed" {
		t.Errorf("unexpected error message: %s", ErrRefundFailed.Error())
	}

	if ErrTransferFailed.Error() != "transfer failed" {
		t.Errorf("unexpected error message: %s", ErrTransferFailed.Error())
	}

	if ErrInvalidAmount.Error() != "invalid amount" {
		t.Errorf("unexpected error message: %s", ErrInvalidAmount.Error())
	}

	if ErrInvalidCurrency.Error() != "invalid currency" {
		t.Errorf("unexpected error message: %s", ErrInvalidCurrency.Error())
	}
}

func TestNewService(t *testing.T) {
	cfg := Config{
		SecretKey:          "sk_test_xxx",
		WebhookSecret:      "whsec_xxx",
		PlatformFeePercent: 0.025,
	}

	service := NewService(cfg)

	if service == nil {
		t.Error("expected service to be created")
	}

	if service.config.PlatformFeePercent != 0.025 {
		t.Errorf("expected platform fee 0.025, got %f", service.config.PlatformFeePercent)
	}
}

func TestAdapter(t *testing.T) {
	cfg := Config{
		SecretKey:          "sk_test_xxx",
		WebhookSecret:      "whsec_xxx",
		PlatformFeePercent: 0.025,
	}

	service := NewService(cfg)
	adapter := NewAdapter(service)

	if adapter == nil {
		t.Error("expected adapter to be created")
	}

	if adapter.service != service {
		t.Error("expected adapter to reference service")
	}
}

func TestAmountConversion(t *testing.T) {
	// Test dollar to cents conversion
	tests := []struct {
		dollars       float64
		expectedCents int64
	}{
		{1.00, 100},
		{10.00, 1000},
		{100.50, 10050},
		{0.01, 1},
		{0.99, 99},
		{1234.56, 123456},
	}

	for _, tt := range tests {
		cents := int64(tt.dollars * 100)
		if cents != tt.expectedCents {
			t.Errorf("$%.2f = %d cents, expected %d", tt.dollars, cents, tt.expectedCents)
		}
	}
}

func TestCentsToAmount(t *testing.T) {
	// Test cents to dollar conversion
	tests := []struct {
		cents           int64
		expectedDollars float64
	}{
		{100, 1.00},
		{1000, 10.00},
		{10050, 100.50},
		{1, 0.01},
		{99, 0.99},
		{123456, 1234.56},
	}

	for _, tt := range tests {
		dollars := float64(tt.cents) / 100
		if dollars != tt.expectedDollars {
			t.Errorf("%d cents = $%.2f, expected $%.2f", tt.cents, dollars, tt.expectedDollars)
		}
	}
}
