package user

import (
	"time"

	"github.com/google/uuid"
)

// User represents a human user authenticated via Clerk.
type User struct {
	ID                          uuid.UUID `json:"id"`
	ClerkUserID                 string    `json:"clerk_user_id"`
	Email                       string    `json:"email"`
	Name                        string    `json:"name,omitempty"`
	AvatarURL                   string    `json:"avatar_url,omitempty"`
	StripeConnectAccountID        string    `json:"stripe_connect_account_id,omitempty"`
	StripeConnectChargesEnabled   bool      `json:"stripe_connect_charges_enabled"`
	StripeCustomerID              string    `json:"stripe_customer_id,omitempty"`
	StripeDefaultPaymentMethodID  string    `json:"stripe_default_payment_method_id,omitempty"`
	CreatedAt                   time.Time `json:"created_at"`
	UpdatedAt                   time.Time `json:"updated_at"`
}

// OwnershipToken represents a token for claiming agent ownership.
type OwnershipToken struct {
	ID           uuid.UUID  `json:"id"`
	AgentID      uuid.UUID  `json:"agent_id"`
	Token        string     `json:"token,omitempty"` // Only set on creation, never stored
	TokenHash    string     `json:"-"`               // Stored hash
	ExpiresAt    time.Time  `json:"expires_at"`
	UsedAt       *time.Time `json:"used_at,omitempty"`
	UsedByUserID *uuid.UUID `json:"used_by_user_id,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// IsExpired returns true if the token has expired.
func (t *OwnershipToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsUsed returns true if the token has been used.
func (t *OwnershipToken) IsUsed() bool {
	return t.UsedAt != nil
}

// ClaimOwnershipRequest is the request body for claiming agent ownership.
type ClaimOwnershipRequest struct {
	Token string `json:"token"`
}

// OwnershipTokenResponse is returned when generating an ownership token.
type OwnershipTokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// AgentMetrics contains metrics for an agent visible to owners.
type AgentMetrics struct {
	AgentID           uuid.UUID `json:"agent_id"`
	AgentName         string    `json:"agent_name"`
	TotalTransactions int       `json:"total_transactions"`
	SuccessfulTrades  int       `json:"successful_trades"`
	FailedTrades      int       `json:"failed_trades"`
	TotalRevenue      float64   `json:"total_revenue"`
	TotalSpent        float64   `json:"total_spent"`
	AverageRating     float64   `json:"average_rating"`
	ActiveListings    int       `json:"active_listings"`
	ActiveRequests    int       `json:"active_requests"`
	PendingOffers     int       `json:"pending_offers"`
	ActiveAuctions    int       `json:"active_auctions"`
	TrustScore        float64   `json:"trust_score"`
	CreatedAt         time.Time `json:"created_at"`
	LastSeenAt        *time.Time `json:"last_seen_at,omitempty"`
}

// OwnedAgentSummary is a summary of an owned agent for listing.
type OwnedAgentSummary struct {
	ID                uuid.UUID  `json:"id"`
	Name              string     `json:"name"`
	Description       string     `json:"description,omitempty"`
	TrustScore        float64    `json:"trust_score"`
	TotalTransactions int        `json:"total_transactions"`
	AverageRating     float64    `json:"average_rating"`
	IsActive          bool       `json:"is_active"`
	CreatedAt         time.Time  `json:"created_at"`
	LastSeenAt        *time.Time `json:"last_seen_at,omitempty"`
}
