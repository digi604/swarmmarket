package payment

import (
	"context"
	"fmt"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/account"
	"github.com/stripe/stripe-go/v76/accountlink"
	"github.com/stripe/stripe-go/v76/loginlink"
)

// AccountStatus is the current state of a Connect account (queried from Stripe, not cached).
type AccountStatus struct {
	AccountID        string `json:"account_id"`
	ChargesEnabled   bool   `json:"charges_enabled"`
	PayoutsEnabled   bool   `json:"payouts_enabled"`
	DetailsSubmitted bool   `json:"details_submitted"`
}

// ConnectService manages Stripe Connect Express accounts.
type ConnectService struct{}

// NewConnectService creates a new Connect service.
// Stripe key must already be set via payment.NewService or stripe.Key.
func NewConnectService() *ConnectService {
	return &ConnectService{}
}

// CreateExpressAccount creates a new Express Connect account.
func (s *ConnectService) CreateExpressAccount(ctx context.Context, email string) (string, error) {
	params := &stripe.AccountParams{
		Type:  stripe.String(string(stripe.AccountTypeExpress)),
		Email: stripe.String(email),
		Capabilities: &stripe.AccountCapabilitiesParams{
			CardPayments: &stripe.AccountCapabilitiesCardPaymentsParams{
				Requested: stripe.Bool(true),
			},
			Transfers: &stripe.AccountCapabilitiesTransfersParams{
				Requested: stripe.Bool(true),
			},
		},
	}

	acct, err := account.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create express account: %w", err)
	}
	return acct.ID, nil
}

// CreateAccountLink generates an onboarding or re-onboarding URL.
func (s *ConnectService) CreateAccountLink(ctx context.Context, accountID, refreshURL, returnURL string) (string, error) {
	params := &stripe.AccountLinkParams{
		Account:    stripe.String(accountID),
		RefreshURL: stripe.String(refreshURL),
		ReturnURL:  stripe.String(returnURL),
		Type:       stripe.String("account_onboarding"),
	}

	link, err := accountlink.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create account link: %w", err)
	}
	return link.URL, nil
}

// CreateLoginLink generates a link to the Express dashboard.
func (s *ConnectService) CreateLoginLink(ctx context.Context, accountID string) (string, error) {
	params := &stripe.LoginLinkParams{
		Account: stripe.String(accountID),
	}

	link, err := loginlink.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create login link: %w", err)
	}
	return link.URL, nil
}

// GetAccountStatus queries Stripe for current account state.
func (s *ConnectService) GetAccountStatus(ctx context.Context, accountID string) (*AccountStatus, error) {
	acct, err := account.GetByID(accountID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	return &AccountStatus{
		AccountID:        acct.ID,
		ChargesEnabled:   acct.ChargesEnabled,
		PayoutsEnabled:   acct.PayoutsEnabled,
		DetailsSubmitted: acct.DetailsSubmitted,
	}, nil
}
