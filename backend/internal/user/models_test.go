package user

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestUser_ConnectFields(t *testing.T) {
	u := User{
		ID:                          uuid.New(),
		ClerkUserID:                 "clerk_123",
		Email:                       "test@example.com",
		StripeConnectAccountID:      "acct_abc",
		StripeConnectChargesEnabled: true,
		CreatedAt:                   time.Now(),
		UpdatedAt:                   time.Now(),
	}

	if u.StripeConnectAccountID != "acct_abc" {
		t.Errorf("expected acct_abc, got %s", u.StripeConnectAccountID)
	}
	if !u.StripeConnectChargesEnabled {
		t.Error("expected charges_enabled = true")
	}
}

func TestUser_ConnectFieldsDefault(t *testing.T) {
	u := User{
		ID:          uuid.New(),
		ClerkUserID: "clerk_456",
		Email:       "test2@example.com",
	}

	if u.StripeConnectAccountID != "" {
		t.Errorf("expected empty account id, got %s", u.StripeConnectAccountID)
	}
	if u.StripeConnectChargesEnabled {
		t.Error("expected charges_enabled = false by default")
	}
}

func TestOwnershipToken_IsExpired(t *testing.T) {
	token := OwnershipToken{
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	if !token.IsExpired() {
		t.Error("expected token to be expired")
	}

	token.ExpiresAt = time.Now().Add(1 * time.Hour)
	if token.IsExpired() {
		t.Error("expected token to not be expired")
	}
}

func TestOwnershipToken_IsUsed(t *testing.T) {
	token := OwnershipToken{}
	if token.IsUsed() {
		t.Error("expected token to not be used")
	}

	now := time.Now()
	token.UsedAt = &now
	if !token.IsUsed() {
		t.Error("expected token to be used")
	}
}
