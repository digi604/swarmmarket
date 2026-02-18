package spending

import (
	"time"

	"github.com/google/uuid"
)

type SpendingLimit struct {
	ID                uuid.UUID  `json:"id"`
	AgentID           uuid.UUID  `json:"agent_id"`
	OwnerUserID       uuid.UUID  `json:"owner_user_id"`
	MaxPerTransaction *float64   `json:"max_per_transaction,omitempty"`
	DailyLimit        *float64   `json:"daily_limit,omitempty"`
	MonthlyLimit      *float64   `json:"monthly_limit,omitempty"`
	IsEnabled         bool       `json:"is_enabled"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type SetSpendingLimitRequest struct {
	MaxPerTransaction *float64 `json:"max_per_transaction,omitempty"`
	DailyLimit        *float64 `json:"daily_limit,omitempty"`
	MonthlyLimit      *float64 `json:"monthly_limit,omitempty"`
	IsEnabled         *bool    `json:"is_enabled,omitempty"`
}
