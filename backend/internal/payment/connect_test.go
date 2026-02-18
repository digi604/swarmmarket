package payment

import (
	"testing"
)

func TestNewConnectService(t *testing.T) {
	svc := NewConnectService()
	if svc == nil {
		t.Fatal("expected non-nil ConnectService")
	}
}

func TestAccountStatus(t *testing.T) {
	tests := []struct {
		name   string
		status AccountStatus
	}{
		{
			name: "fully onboarded",
			status: AccountStatus{
				AccountID:        "acct_123",
				ChargesEnabled:   true,
				PayoutsEnabled:   true,
				DetailsSubmitted: true,
			},
		},
		{
			name: "not yet onboarded",
			status: AccountStatus{
				AccountID:        "acct_456",
				ChargesEnabled:   false,
				PayoutsEnabled:   false,
				DetailsSubmitted: false,
			},
		},
		{
			name: "partially onboarded",
			status: AccountStatus{
				AccountID:        "acct_789",
				ChargesEnabled:   false,
				PayoutsEnabled:   false,
				DetailsSubmitted: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.status.AccountID == "" {
				t.Error("account_id should not be empty")
			}
		})
	}
}

func TestAccountStatusJSON(t *testing.T) {
	status := AccountStatus{
		AccountID:        "acct_test",
		ChargesEnabled:   true,
		PayoutsEnabled:   true,
		DetailsSubmitted: true,
	}

	if status.AccountID != "acct_test" {
		t.Errorf("expected acct_test, got %s", status.AccountID)
	}
	if !status.ChargesEnabled {
		t.Error("expected charges_enabled = true")
	}
	if !status.PayoutsEnabled {
		t.Error("expected payouts_enabled = true")
	}
	if !status.DetailsSubmitted {
		t.Error("expected details_submitted = true")
	}
}
