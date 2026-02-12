package transaction

import (
	"time"

	"github.com/google/uuid"
)

// TransactionStatus represents the status of a transaction.
type TransactionStatus string

const (
	StatusPending      TransactionStatus = "pending"
	StatusEscrowFunded TransactionStatus = "escrow_funded"
	StatusDelivered    TransactionStatus = "delivered"
	StatusCompleted    TransactionStatus = "completed"
	StatusDisputed     TransactionStatus = "disputed"
	StatusRefunded     TransactionStatus = "refunded"
	StatusCancelled    TransactionStatus = "cancelled"
)

// EscrowStatus represents the status of an escrow account.
type EscrowStatus string

const (
	EscrowPending  EscrowStatus = "pending"
	EscrowFunded   EscrowStatus = "funded"
	EscrowReleased EscrowStatus = "released"
	EscrowRefunded EscrowStatus = "refunded"
	EscrowDisputed EscrowStatus = "disputed"
)

// Transaction represents a marketplace transaction between buyer and seller.
type Transaction struct {
	ID                  uuid.UUID         `json:"id"`
	BuyerID             uuid.UUID         `json:"buyer_id"`
	SellerID            uuid.UUID         `json:"seller_id"`
	ListingID           *uuid.UUID        `json:"listing_id,omitempty"`
	RequestID           *uuid.UUID        `json:"request_id,omitempty"`
	OfferID             *uuid.UUID        `json:"offer_id,omitempty"`
	AuctionID           *uuid.UUID        `json:"auction_id,omitempty"`
	TaskID              *uuid.UUID        `json:"task_id,omitempty"`
	Amount              float64           `json:"amount"`
	Currency            string            `json:"currency"`
	PlatformFee         float64           `json:"platform_fee"`
	Status              TransactionStatus `json:"status"`
	DeliveryConfirmedAt *time.Time        `json:"delivery_confirmed_at,omitempty"`
	CompletedAt         *time.Time        `json:"completed_at,omitempty"`
	Metadata            map[string]any    `json:"metadata,omitempty"`
	CreatedAt           time.Time         `json:"created_at"`
	UpdatedAt           time.Time         `json:"updated_at"`

	// Joined fields for display
	BuyerName  string `json:"buyer_name,omitempty"`
	SellerName string `json:"seller_name,omitempty"`
}

// EscrowAccount holds funds during a transaction.
type EscrowAccount struct {
	ID                    uuid.UUID    `json:"id"`
	TransactionID         uuid.UUID    `json:"transaction_id"`
	Amount                float64      `json:"amount"`
	Currency              string       `json:"currency"`
	Status                EscrowStatus `json:"status"`
	FundedAt              *time.Time   `json:"funded_at,omitempty"`
	ReleasedAt            *time.Time   `json:"released_at,omitempty"`
	StripePaymentIntentID *string      `json:"stripe_payment_intent_id,omitempty"`
	Metadata              map[string]any `json:"metadata,omitempty"`
	CreatedAt             time.Time    `json:"created_at"`
	UpdatedAt             time.Time    `json:"updated_at"`
}

// Rating represents a rating given after a transaction.
type Rating struct {
	ID            uuid.UUID `json:"id"`
	TransactionID uuid.UUID `json:"transaction_id"`
	RaterID       uuid.UUID `json:"rater_id"`
	RatedAgentID  uuid.UUID `json:"rated_agent_id"`
	Score         int       `json:"score"` // 1-5
	Comment       string    `json:"comment,omitempty"`
	CreatedAt     time.Time `json:"created_at"`

	// Joined fields
	RaterName string `json:"rater_name,omitempty"`
}

// --- Request/Response DTOs ---

// CreateTransactionRequest is used internally when creating a transaction.
type CreateTransactionRequest struct {
	BuyerID   uuid.UUID
	SellerID  uuid.UUID
	ListingID *uuid.UUID
	RequestID *uuid.UUID
	OfferID   *uuid.UUID
	AuctionID *uuid.UUID
	TaskID    *uuid.UUID
	Amount    float64
	Currency  string
}

// ConfirmDeliveryRequest is the request body for confirming delivery.
type ConfirmDeliveryRequest struct {
	Notes string `json:"notes,omitempty"`
}

// SubmitRatingRequest is the request body for submitting a rating.
type SubmitRatingRequest struct {
	Score   int    `json:"score"` // 1-5
	Comment string `json:"comment,omitempty"`
}

// DisputeRequest is the request body for opening a dispute.
type DisputeRequest struct {
	Reason      string `json:"reason"`
	Description string `json:"description"`
}

// TransactionListResult is a paginated list of transactions.
type TransactionListResult struct {
	Items  []*Transaction `json:"items"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

// ListTransactionsParams contains filter parameters for listing transactions.
type ListTransactionsParams struct {
	AgentID *uuid.UUID         // Filter by buyer OR seller
	Status  *TransactionStatus // Filter by status
	Role    string             // "buyer", "seller", or "" for both
	Limit   int
	Offset  int
}

// EscrowFundingResult is returned when initiating escrow funding.
type EscrowFundingResult struct {
	TransactionID   uuid.UUID `json:"transaction_id"`
	PaymentIntentID string    `json:"payment_intent_id"`
	Amount          float64   `json:"amount"`
	Currency        string    `json:"currency"`
}
