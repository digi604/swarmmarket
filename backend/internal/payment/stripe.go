package payment

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/paymentmethod"
	"github.com/stripe/stripe-go/v76/refund"
	"github.com/stripe/stripe-go/v76/setupintent"
	"github.com/stripe/stripe-go/v76/transfer"
)

var (
	ErrPaymentFailed    = errors.New("payment failed")
	ErrRefundFailed     = errors.New("refund failed")
	ErrTransferFailed   = errors.New("transfer failed")
	ErrInvalidAmount    = errors.New("invalid amount")
	ErrInvalidCurrency  = errors.New("invalid currency")
	ErrSellerNotPayable = errors.New("seller is not set up to receive payments")
	ErrNoPaymentMethod  = errors.New("no saved payment method; owner must add one in the dashboard")
)

// ConnectAccountResolver resolves a seller agent's Connect account ID.
type ConnectAccountResolver interface {
	GetConnectAccountIDForAgent(ctx context.Context, agentID uuid.UUID) (string, error)
}

// PaymentMethodResolver resolves an agent's owner's saved payment method.
type PaymentMethodResolver interface {
	GetPaymentMethodForAgent(ctx context.Context, agentID uuid.UUID) (customerID, pmID string, ownerUserID uuid.UUID, err error)
}

// SpendingChecker checks spending limits for an agent.
type SpendingChecker interface {
	CheckSpendingLimit(ctx context.Context, agentID uuid.UUID, amount float64) error
}

// Config holds Stripe configuration.
type Config struct {
	SecretKey          string
	WebhookSecret      string
	PlatformFeePercent float64
	DefaultReturnURL   string
}

// Service handles Stripe payments for escrow.
type Service struct {
	config Config
}

// NewService creates a new payment service.
func NewService(cfg Config) *Service {
	stripe.Key = cfg.SecretKey
	return &Service{config: cfg}
}

// CreateEscrowPayment creates a payment intent for escrow (original, for direct Stripe calls).
func (s *Service) CreateEscrowPayment(ctx context.Context, req *CreatePaymentRequest) (*PaymentResult, error) {
	if req.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	amountCents := int64(req.Amount * 100)
	platformFee := int64(float64(amountCents) * s.config.PlatformFeePercent)

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amountCents),
		Currency: stripe.String(normalizeCurrency(req.Currency)),
		Metadata: map[string]string{
			"transaction_id": req.TransactionID.String(),
			"buyer_id":       req.BuyerID.String(),
			"seller_id":      req.SellerID.String(),
		},
		CaptureMethod: stripe.String("manual"),
	}

	// Off-session with saved payment method
	if req.CustomerID != "" && req.PaymentMethodID != "" {
		params.Customer = stripe.String(req.CustomerID)
		params.PaymentMethod = stripe.String(req.PaymentMethodID)
		params.OffSession = stripe.Bool(true)
		params.Confirm = stripe.Bool(true)
	}

	if req.SellerStripeAccountID != "" {
		params.TransferData = &stripe.PaymentIntentTransferDataParams{
			Destination: stripe.String(req.SellerStripeAccountID),
		}
		params.ApplicationFeeAmount = stripe.Int64(platformFee)
	}

	intent, err := paymentintent.New(params)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPaymentFailed, err)
	}

	return &PaymentResult{
		PaymentIntentID: intent.ID,
		Status:          string(intent.Status),
		Amount:          req.Amount,
		Currency:        req.Currency,
	}, nil
}

// CapturePayment captures a held payment (releases from escrow to seller).
func (s *Service) CapturePayment(ctx context.Context, paymentIntentID string) error {
	_, err := paymentintent.Capture(paymentIntentID, nil)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrPaymentFailed, err)
	}
	return nil
}

// RefundPayment refunds a payment.
func (s *Service) RefundPayment(ctx context.Context, paymentIntentID string, amount *float64) error {
	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(paymentIntentID),
	}
	if amount != nil {
		params.Amount = stripe.Int64(int64(*amount * 100))
	}
	_, err := refund.New(params)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrRefundFailed, err)
	}
	return nil
}

// TransferToSeller transfers funds directly to seller's connected account.
func (s *Service) TransferToSeller(ctx context.Context, req *TransferRequest) (*TransferResult, error) {
	if req.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	amountCents := int64(req.Amount * 100)
	params := &stripe.TransferParams{
		Amount:      stripe.Int64(amountCents),
		Currency:    stripe.String(normalizeCurrency(req.Currency)),
		Destination: stripe.String(req.SellerStripeAccountID),
		Metadata: map[string]string{
			"transaction_id": req.TransactionID.String(),
		},
	}
	if req.SourceTransactionID != "" {
		params.SourceTransaction = stripe.String(req.SourceTransactionID)
	}

	xfer, err := transfer.New(params)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTransferFailed, err)
	}

	return &TransferResult{
		TransferID: xfer.ID,
		Amount:     req.Amount,
		Currency:   req.Currency,
		Status:     "completed",
	}, nil
}

// GetPaymentIntent retrieves a payment intent.
func (s *Service) GetPaymentIntent(ctx context.Context, paymentIntentID string) (*PaymentStatus, error) {
	intent, err := paymentintent.Get(paymentIntentID, nil)
	if err != nil {
		return nil, err
	}
	return &PaymentStatus{
		PaymentIntentID: intent.ID,
		Status:          string(intent.Status),
		Amount:          float64(intent.Amount) / 100,
		Currency:        string(intent.Currency),
		CapturedAmount:  float64(intent.AmountReceived) / 100,
	}, nil
}

// CreateCustomer creates a Stripe Customer.
func (s *Service) CreateCustomer(ctx context.Context, email, name string) (string, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(email),
		Name:  stripe.String(name),
	}
	c, err := customer.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create stripe customer: %w", err)
	}
	return c.ID, nil
}

// CreateSetupIntent creates a SetupIntent for saving a payment method.
func (s *Service) CreateSetupIntent(ctx context.Context, customerID string) (string, error) {
	params := &stripe.SetupIntentParams{
		Customer: stripe.String(customerID),
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
	}
	si, err := setupintent.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create setup intent: %w", err)
	}
	return si.ClientSecret, nil
}

// PaymentMethodInfo describes a saved payment method.
type PaymentMethodInfo struct {
	ID        string `json:"id"`
	Brand     string `json:"brand"`
	Last4     string `json:"last4"`
	ExpMonth  int64  `json:"exp_month"`
	ExpYear   int64  `json:"exp_year"`
	IsDefault bool   `json:"is_default"`
}

// ListPaymentMethods lists saved payment methods for a customer.
func (s *Service) ListPaymentMethods(ctx context.Context, customerID, defaultPMID string) ([]*PaymentMethodInfo, error) {
	params := &stripe.PaymentMethodListParams{
		Customer: stripe.String(customerID),
		Type:     stripe.String("card"),
	}
	iter := paymentmethod.List(params)
	var methods []*PaymentMethodInfo
	for iter.Next() {
		pm := iter.PaymentMethod()
		info := &PaymentMethodInfo{
			ID:        pm.ID,
			IsDefault: pm.ID == defaultPMID,
		}
		if pm.Card != nil {
			info.Brand = string(pm.Card.Brand)
			info.Last4 = pm.Card.Last4
			info.ExpMonth = pm.Card.ExpMonth
			info.ExpYear = pm.Card.ExpYear
		}
		methods = append(methods, info)
	}
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to list payment methods: %w", err)
	}
	return methods, nil
}

// DetachPaymentMethod removes a payment method from the customer.
func (s *Service) DetachPaymentMethod(ctx context.Context, pmID string) error {
	_, err := paymentmethod.Detach(pmID, nil)
	if err != nil {
		return fmt.Errorf("failed to detach payment method: %w", err)
	}
	return nil
}

// --- Types ---

type CreatePaymentRequest struct {
	TransactionID         uuid.UUID
	BuyerID               uuid.UUID
	SellerID              uuid.UUID
	Amount                float64
	Currency              string
	CustomerID            string
	PaymentMethodID       string
	SellerStripeAccountID string
}

type PaymentResult struct {
	PaymentIntentID string  `json:"payment_intent_id"`
	Status          string  `json:"status"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
}

type PaymentStatus struct {
	PaymentIntentID string  `json:"payment_intent_id"`
	Status          string  `json:"status"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
	CapturedAmount  float64 `json:"captured_amount"`
}

type TransferRequest struct {
	TransactionID         uuid.UUID
	SellerStripeAccountID string
	Amount                float64
	Currency              string
	SourceTransactionID   string
}

type TransferResult struct {
	TransferID string  `json:"transfer_id"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
	Status     string  `json:"status"`
}

// --- Adapter ---

// Adapter implements transaction.PaymentService and marketplace.PaymentCreator interfaces.
type Adapter struct {
	service          *Service
	resolver         ConnectAccountResolver
	paymentResolver  PaymentMethodResolver
	spendingChecker  SpendingChecker
}

func NewAdapter(service *Service) *Adapter {
	return &Adapter{service: service}
}

func (a *Adapter) SetConnectAccountResolver(resolver ConnectAccountResolver) {
	a.resolver = resolver
}

func (a *Adapter) SetPaymentMethodResolver(resolver PaymentMethodResolver) {
	a.paymentResolver = resolver
}

func (a *Adapter) SetSpendingChecker(checker SpendingChecker) {
	a.spendingChecker = checker
}

// CreateEscrowPayment resolves payment method + Connect account, checks spending limits,
// and creates an off-session PaymentIntent.
func (a *Adapter) CreateEscrowPayment(ctx context.Context, transactionID, buyerID, sellerID string, amount float64, currency string) (string, error) {
	txID, _ := uuid.Parse(transactionID)
	bID, _ := uuid.Parse(buyerID)
	sID, _ := uuid.Parse(sellerID)

	// Check spending limits
	if a.spendingChecker != nil {
		if err := a.spendingChecker.CheckSpendingLimit(ctx, bID, amount); err != nil {
			return "", err
		}
	}

	// Resolve buyer's saved payment method
	var customerID, pmID string
	if a.paymentResolver != nil {
		var err error
		customerID, pmID, _, err = a.paymentResolver.GetPaymentMethodForAgent(ctx, bID)
		if err != nil {
			return "", fmt.Errorf("failed to resolve payment method: %w", err)
		}
		if customerID == "" || pmID == "" {
			return "", ErrNoPaymentMethod
		}
	}

	// Resolve seller Connect account
	var sellerAccount string
	if a.resolver != nil {
		var err error
		sellerAccount, err = a.resolver.GetConnectAccountIDForAgent(ctx, sID)
		if err != nil {
			return "", fmt.Errorf("failed to resolve seller connect account: %w", err)
		}
		if sellerAccount == "" {
			return "", ErrSellerNotPayable
		}
	}

	result, err := a.service.CreateEscrowPayment(ctx, &CreatePaymentRequest{
		TransactionID:         txID,
		BuyerID:               bID,
		SellerID:              sID,
		Amount:                amount,
		Currency:              currency,
		CustomerID:            customerID,
		PaymentMethodID:       pmID,
		SellerStripeAccountID: sellerAccount,
	})
	if err != nil {
		return "", err
	}
	return result.PaymentIntentID, nil
}

// CapturePayment captures a held payment.
func (a *Adapter) CapturePayment(ctx context.Context, paymentIntentID string) error {
	return a.service.CapturePayment(ctx, paymentIntentID)
}

// RefundPayment refunds a payment.
func (a *Adapter) RefundPayment(ctx context.Context, paymentIntentID string) error {
	return a.service.RefundPayment(ctx, paymentIntentID, nil)
}

func normalizeCurrency(currency string) string {
	if currency == "" {
		return "usd"
	}
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
