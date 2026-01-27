package agent

import (
	"time"

	"github.com/google/uuid"
)

// VerificationLevel represents the verification status of an agent.
type VerificationLevel string

const (
	VerificationBasic    VerificationLevel = "basic"
	VerificationVerified VerificationLevel = "verified"
	VerificationPremium  VerificationLevel = "premium"
)

// Agent represents an AI agent registered in the marketplace.
type Agent struct {
	ID                uuid.UUID         `json:"id"`
	Name              string            `json:"name"`
	Description       string            `json:"description,omitempty"`
	OwnerEmail        string            `json:"owner_email,omitempty"`
	APIKeyHash        string            `json:"-"` // Never expose
	APIKeyPrefix      string            `json:"api_key_prefix,omitempty"`
	VerificationLevel VerificationLevel `json:"verification_level"`
	TrustScore        float64           `json:"trust_score"`
	TotalTransactions int               `json:"total_transactions"`
	SuccessfulTrades  int               `json:"successful_trades"`
	AverageRating     float64           `json:"average_rating"`
	IsActive          bool              `json:"is_active"`
	Metadata          map[string]any    `json:"metadata,omitempty"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
	LastSeenAt        *time.Time        `json:"last_seen_at,omitempty"`
}

// PublicProfile returns a sanitized version of the agent for public display.
func (a *Agent) PublicProfile() *AgentPublicProfile {
	return &AgentPublicProfile{
		ID:                a.ID,
		Name:              a.Name,
		Description:       a.Description,
		VerificationLevel: a.VerificationLevel,
		TrustScore:        a.TrustScore,
		TotalTransactions: a.TotalTransactions,
		SuccessfulTrades:  a.SuccessfulTrades,
		AverageRating:     a.AverageRating,
		CreatedAt:         a.CreatedAt,
	}
}

// AgentPublicProfile is the public-facing agent information.
type AgentPublicProfile struct {
	ID                uuid.UUID         `json:"id"`
	Name              string            `json:"name"`
	Description       string            `json:"description,omitempty"`
	VerificationLevel VerificationLevel `json:"verification_level"`
	TrustScore        float64           `json:"trust_score"`
	TotalTransactions int               `json:"total_transactions"`
	SuccessfulTrades  int               `json:"successful_trades"`
	AverageRating     float64           `json:"average_rating"`
	CreatedAt         time.Time         `json:"created_at"`
}

// Reputation holds detailed reputation information for an agent.
type Reputation struct {
	AgentID           uuid.UUID        `json:"agent_id"`
	TrustScore        float64          `json:"trust_score"`
	TotalTransactions int              `json:"total_transactions"`
	SuccessfulTrades  int              `json:"successful_trades"`
	FailedTrades      int              `json:"failed_trades"`
	DisputesWon       int              `json:"disputes_won"`
	DisputesLost      int              `json:"disputes_lost"`
	AverageRating     float64          `json:"average_rating"`
	RatingCount       int              `json:"rating_count"`
	RecentRatings     []Rating         `json:"recent_ratings,omitempty"`
	CategoryScores    map[string]float64 `json:"category_scores,omitempty"`
}

// Rating represents a single rating from a transaction.
type Rating struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	RaterID       uuid.UUID `json:"rater_id"`
	Score         int       `json:"score"` // 1-5
	Comment       string    `json:"comment,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// RegisterRequest is the request body for agent registration.
type RegisterRequest struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	OwnerEmail  string         `json:"owner_email"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// RegisterResponse is the response for agent registration.
type RegisterResponse struct {
	Agent  *Agent `json:"agent"`
	APIKey string `json:"api_key"` // Only returned once at registration
}

// UpdateRequest is the request body for updating an agent.
type UpdateRequest struct {
	Name        *string        `json:"name,omitempty"`
	Description *string        `json:"description,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}
