package wallet

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/paymentintent"
)

var (
	ErrInvalidAmount   = errors.New("invalid amount: must be greater than 0")
	ErrDepositNotFound = errors.New("deposit not found")
	ErrPaymentFailed   = errors.New("payment failed")
)

// StripeConfig holds Stripe configuration for deposits.
type StripeConfig struct {
	SecretKey string
}

// Service handles wallet operations.
type Service struct {
	repo         *Repository
	stripeConfig StripeConfig
}

// NewService creates a new wallet service.
func NewService(repo *Repository, stripeConfig StripeConfig) *Service {
	stripe.Key = stripeConfig.SecretKey
	return &Service{
		repo:         repo,
		stripeConfig: stripeConfig,
	}
}

// CreateDeposit creates a new deposit with a Stripe payment intent.
func (s *Service) CreateDeposit(ctx context.Context, userID uuid.UUID, req *CreateDepositRequest) (*CreateDepositResponse, error) {
	if req.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	// Normalize currency
	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	// Convert to cents for Stripe
	amountCents := int64(req.Amount * 100)

	// Create Stripe payment intent
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amountCents),
		Currency: stripe.String(normalizeCurrency(currency)),
		Metadata: map[string]string{
			"type":    "wallet_deposit",
			"user_id": userID.String(),
		},
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}

	if req.ReturnURL != "" {
		params.ReturnURL = stripe.String(req.ReturnURL)
	}

	intent, err := paymentintent.New(params)
	if err != nil {
		return nil, err
	}

	// Create deposit record
	deposit := &Deposit{
		ID:                    uuid.New(),
		UserID:                userID,
		Amount:                req.Amount,
		Currency:              currency,
		StripePaymentIntentID: intent.ID,
		StripeClientSecret:    intent.ClientSecret,
		Status:                DepositStatusPending,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	if err := s.repo.CreateDeposit(ctx, deposit); err != nil {
		return nil, err
	}

	// Create Stripe Checkout session for easy payment
	checkoutURL := ""
	if req.ReturnURL != "" {
		checkoutParams := &stripe.CheckoutSessionParams{
			Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
			PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
				Metadata: map[string]string{
					"deposit_id": deposit.ID.String(),
					"type":       "wallet_deposit",
					"user_id":    userID.String(),
				},
			},
			LineItems: []*stripe.CheckoutSessionLineItemParams{
				{
					PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
						Currency: stripe.String(normalizeCurrency(currency)),
						ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
							Name:        stripe.String("Wallet Deposit"),
							Description: stripe.String(fmt.Sprintf("Add $%.2f to your SwarmMarket wallet", req.Amount)),
						},
						UnitAmount: stripe.Int64(amountCents),
					},
					Quantity: stripe.Int64(1),
				},
			},
			SuccessURL: stripe.String(req.ReturnURL + "?deposit=success&deposit_id=" + deposit.ID.String()),
			CancelURL:  stripe.String(req.ReturnURL + "?deposit=cancelled"),
		}
		checkoutSession, err := session.New(checkoutParams)
		if err == nil {
			checkoutURL = checkoutSession.URL
		}
	}

	return &CreateDepositResponse{
		DepositID:    deposit.ID,
		ClientSecret: intent.ClientSecret,
		CheckoutURL:  checkoutURL,
		Amount:       req.Amount,
		Currency:     currency,
		Instructions: "To complete this deposit, either: (1) Open the checkout_url in a browser to pay via Stripe Checkout, or (2) Use the client_secret with Stripe.js/Elements to build a custom payment form. The deposit will be credited to your wallet once payment is confirmed.",
	}, nil
}

// GetDeposit retrieves a deposit by ID.
func (s *Service) GetDeposit(ctx context.Context, id uuid.UUID) (*Deposit, error) {
	deposit, err := s.repo.GetDeposit(ctx, id)
	if err != nil {
		return nil, err
	}
	if deposit == nil {
		return nil, ErrDepositNotFound
	}
	return deposit, nil
}

// GetUserDeposits retrieves all deposits for a user.
func (s *Service) GetUserDeposits(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Deposit, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.GetUserDeposits(ctx, userID, limit, offset)
}

// GetWalletBalance calculates the user's wallet balance.
func (s *Service) GetWalletBalance(ctx context.Context, userID uuid.UUID) (*WalletBalance, error) {
	available, err := s.repo.GetCompletedDepositsTotal(ctx, userID)
	if err != nil {
		return nil, err
	}

	pending, err := s.repo.GetPendingDepositsTotal(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &WalletBalance{
		Available: available,
		Pending:   pending,
		Currency:  "USD",
	}, nil
}

// CreateAgentDeposit creates a new deposit for an agent with a Stripe payment intent.
func (s *Service) CreateAgentDeposit(ctx context.Context, agentID uuid.UUID, req *CreateDepositRequest) (*CreateDepositResponse, error) {
	if req.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	// Normalize currency
	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	// Convert to cents for Stripe
	amountCents := int64(req.Amount * 100)

	// Create Stripe payment intent
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amountCents),
		Currency: stripe.String(normalizeCurrency(currency)),
		Metadata: map[string]string{
			"type":     "wallet_deposit",
			"agent_id": agentID.String(),
		},
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}

	if req.ReturnURL != "" {
		params.ReturnURL = stripe.String(req.ReturnURL)
	}

	intent, err := paymentintent.New(params)
	if err != nil {
		return nil, err
	}

	// Create deposit record
	deposit := &Deposit{
		ID:                    uuid.New(),
		AgentID:               agentID,
		Amount:                req.Amount,
		Currency:              currency,
		StripePaymentIntentID: intent.ID,
		StripeClientSecret:    intent.ClientSecret,
		Status:                DepositStatusPending,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	if err := s.repo.CreateDeposit(ctx, deposit); err != nil {
		return nil, err
	}

	// Create Stripe Checkout session for easy payment
	checkoutURL := ""
	if req.ReturnURL != "" {
		checkoutParams := &stripe.CheckoutSessionParams{
			Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
			PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
				Metadata: map[string]string{
					"deposit_id": deposit.ID.String(),
					"type":       "wallet_deposit",
					"agent_id":   agentID.String(),
				},
			},
			LineItems: []*stripe.CheckoutSessionLineItemParams{
				{
					PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
						Currency: stripe.String(normalizeCurrency(currency)),
						ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
							Name:        stripe.String("Agent Wallet Deposit"),
							Description: stripe.String(fmt.Sprintf("Add $%.2f to agent wallet", req.Amount)),
						},
						UnitAmount: stripe.Int64(amountCents),
					},
					Quantity: stripe.Int64(1),
				},
			},
			SuccessURL: stripe.String(req.ReturnURL + "?deposit=success&deposit_id=" + deposit.ID.String()),
			CancelURL:  stripe.String(req.ReturnURL + "?deposit=cancelled"),
		}
		checkoutSession, err := session.New(checkoutParams)
		if err == nil {
			checkoutURL = checkoutSession.URL
		}
	}

	return &CreateDepositResponse{
		DepositID:    deposit.ID,
		ClientSecret: intent.ClientSecret,
		CheckoutURL:  checkoutURL,
		Amount:       req.Amount,
		Currency:     currency,
		Instructions: "To complete this deposit, either: (1) Open the checkout_url in a browser to pay via Stripe Checkout, or (2) Use the client_secret with Stripe.js/Elements to build a custom payment form. The deposit will be credited to your agent's wallet once payment is confirmed.",
	}, nil
}

// GetAgentDeposits retrieves all deposits for an agent.
func (s *Service) GetAgentDeposits(ctx context.Context, agentID uuid.UUID, limit, offset int) ([]*Deposit, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.GetAgentDeposits(ctx, agentID, limit, offset)
}

// GetAgentWalletBalance calculates the agent's wallet balance.
func (s *Service) GetAgentWalletBalance(ctx context.Context, agentID uuid.UUID) (*WalletBalance, error) {
	available, err := s.repo.GetAgentCompletedDepositsTotal(ctx, agentID)
	if err != nil {
		return nil, err
	}

	pending, err := s.repo.GetAgentPendingDepositsTotal(ctx, agentID)
	if err != nil {
		return nil, err
	}

	return &WalletBalance{
		Available: available,
		Pending:   pending,
		Currency:  "USD",
	}, nil
}

// HandlePaymentIntentSucceeded handles a successful payment intent webhook.
func (s *Service) HandlePaymentIntentSucceeded(ctx context.Context, paymentIntentID string) error {
	deposit, err := s.repo.GetDepositByPaymentIntent(ctx, paymentIntentID)
	if err != nil {
		return err
	}
	if deposit == nil {
		// Not a deposit payment intent, ignore
		return nil
	}

	return s.repo.UpdateDepositStatus(ctx, deposit.ID, DepositStatusCompleted, "")
}

// HandlePaymentIntentFailed handles a failed payment intent webhook.
func (s *Service) HandlePaymentIntentFailed(ctx context.Context, paymentIntentID string, reason string) error {
	deposit, err := s.repo.GetDepositByPaymentIntent(ctx, paymentIntentID)
	if err != nil {
		return err
	}
	if deposit == nil {
		// Not a deposit payment intent, ignore
		return nil
	}

	return s.repo.UpdateDepositStatus(ctx, deposit.ID, DepositStatusFailed, reason)
}

func normalizeCurrency(currency string) string {
	switch currency {
	case "USD", "usd":
		return "usd"
	case "EUR", "eur":
		return "eur"
	case "GBP", "gbp":
		return "gbp"
	default:
		return "usd"
	}
}
