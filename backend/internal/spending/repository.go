package spending

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) GetByAgentID(ctx context.Context, agentID uuid.UUID) (*SpendingLimit, error) {
	var sl SpendingLimit
	err := r.pool.QueryRow(ctx, `
		SELECT id, agent_id, owner_user_id, max_per_transaction, daily_limit, monthly_limit,
		       is_enabled, created_at, updated_at
		FROM agent_spending_limits
		WHERE agent_id = $1
	`, agentID).Scan(
		&sl.ID, &sl.AgentID, &sl.OwnerUserID,
		&sl.MaxPerTransaction, &sl.DailyLimit, &sl.MonthlyLimit,
		&sl.IsEnabled, &sl.CreatedAt, &sl.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get spending limit: %w", err)
	}
	return &sl, nil
}

func (r *Repository) Upsert(ctx context.Context, sl *SpendingLimit) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO agent_spending_limits (agent_id, owner_user_id, max_per_transaction, daily_limit, monthly_limit, is_enabled)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (agent_id) DO UPDATE SET
			max_per_transaction = EXCLUDED.max_per_transaction,
			daily_limit = EXCLUDED.daily_limit,
			monthly_limit = EXCLUDED.monthly_limit,
			is_enabled = EXCLUDED.is_enabled,
			updated_at = NOW()
	`, sl.AgentID, sl.OwnerUserID, sl.MaxPerTransaction, sl.DailyLimit, sl.MonthlyLimit, sl.IsEnabled)
	if err != nil {
		return fmt.Errorf("failed to upsert spending limit: %w", err)
	}
	return nil
}

// GetAgentSpendSince returns the total amount spent by an agent since a given time.
func (r *Repository) GetAgentSpendSince(ctx context.Context, agentID uuid.UUID, since time.Time) (float64, error) {
	var total float64
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE buyer_id = $1
		  AND status NOT IN ('cancelled', 'refunded')
		  AND created_at >= $2
	`, agentID, since).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to get agent spend: %w", err)
	}
	return total, nil
}
