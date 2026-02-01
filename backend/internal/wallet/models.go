package wallet

import (
	"time"

	"github.com/google/uuid"
)

// DepositStatus represents the status of a wallet deposit.
type DepositStatus string

const (
	DepositStatusPending    DepositStatus = "pending"
	DepositStatusProcessing DepositStatus = "processing"
	DepositStatusCompleted  DepositStatus = "completed"
	DepositStatusFailed     DepositStatus = "failed"
	DepositStatusCancelled  DepositStatus = "cancelled"
)

// Deposit represents a wallet deposit.
type Deposit struct {
	ID                    uuid.UUID     `json:"id"`
	UserID                uuid.UUID     `json:"user_id,omitempty"`
	AgentID               uuid.UUID     `json:"agent_id,omitempty"`
	Amount                float64       `json:"amount"`
	Currency              string        `json:"currency"`
	StripePaymentIntentID string        `json:"stripe_payment_intent_id,omitempty"`
	StripeClientSecret    string        `json:"stripe_client_secret,omitempty"`
	Status                DepositStatus `json:"status"`
	FailureReason         string        `json:"failure_reason,omitempty"`
	CreatedAt             time.Time     `json:"created_at"`
	UpdatedAt             time.Time     `json:"updated_at"`
	CompletedAt           *time.Time    `json:"completed_at,omitempty"`
}

// CreateDepositRequest is the request to create a deposit.
type CreateDepositRequest struct {
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	ReturnURL string  `json:"return_url"`
}

// CreateDepositResponse is the response from creating a deposit.
type CreateDepositResponse struct {
	DepositID    uuid.UUID `json:"deposit_id"`
	ClientSecret string    `json:"client_secret"`
	CheckoutURL  string    `json:"checkout_url"`
	Amount       float64   `json:"amount"`
	Currency     string    `json:"currency"`
	Instructions string    `json:"instructions"`
}

// WalletBalance represents the user's wallet balance.
type WalletBalance struct {
	Available float64 `json:"available"`
	Pending   float64 `json:"pending"`
	Currency  string  `json:"currency"`
}
