package wallet

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles wallet database operations.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new wallet repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// CreateDeposit creates a new deposit record.
func (r *Repository) CreateDeposit(ctx context.Context, deposit *Deposit) error {
	query := `
		INSERT INTO wallet_deposits (
			id, user_id, agent_id, amount, currency, stripe_payment_intent_id,
			stripe_client_secret, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	var userID, agentID *uuid.UUID
	if deposit.UserID != uuid.Nil {
		userID = &deposit.UserID
	}
	if deposit.AgentID != uuid.Nil {
		agentID = &deposit.AgentID
	}

	_, err := r.pool.Exec(ctx, query,
		deposit.ID,
		userID,
		agentID,
		deposit.Amount,
		deposit.Currency,
		deposit.StripePaymentIntentID,
		deposit.StripeClientSecret,
		deposit.Status,
		deposit.CreatedAt,
		deposit.UpdatedAt,
	)

	return err
}

// GetDeposit retrieves a deposit by ID.
func (r *Repository) GetDeposit(ctx context.Context, id uuid.UUID) (*Deposit, error) {
	query := `
		SELECT id, user_id, agent_id, amount, currency, stripe_payment_intent_id,
			stripe_client_secret, status, failure_reason, created_at, updated_at, completed_at
		FROM wallet_deposits
		WHERE id = $1
	`

	deposit := &Deposit{}
	var failureReason *string
	var stripePI, stripeCS *string
	var userID, agentID *uuid.UUID

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&deposit.ID,
		&userID,
		&agentID,
		&deposit.Amount,
		&deposit.Currency,
		&stripePI,
		&stripeCS,
		&deposit.Status,
		&failureReason,
		&deposit.CreatedAt,
		&deposit.UpdatedAt,
		&deposit.CompletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if userID != nil {
		deposit.UserID = *userID
	}
	if agentID != nil {
		deposit.AgentID = *agentID
	}
	if stripePI != nil {
		deposit.StripePaymentIntentID = *stripePI
	}
	if stripeCS != nil {
		deposit.StripeClientSecret = *stripeCS
	}
	if failureReason != nil {
		deposit.FailureReason = *failureReason
	}

	return deposit, nil
}

// GetDepositByPaymentIntent retrieves a deposit by Stripe payment intent ID.
func (r *Repository) GetDepositByPaymentIntent(ctx context.Context, paymentIntentID string) (*Deposit, error) {
	query := `
		SELECT id, user_id, agent_id, amount, currency, stripe_payment_intent_id,
			stripe_client_secret, status, failure_reason, created_at, updated_at, completed_at
		FROM wallet_deposits
		WHERE stripe_payment_intent_id = $1
	`

	deposit := &Deposit{}
	var failureReason *string
	var stripePI, stripeCS *string
	var userID, agentID *uuid.UUID

	err := r.pool.QueryRow(ctx, query, paymentIntentID).Scan(
		&deposit.ID,
		&userID,
		&agentID,
		&deposit.Amount,
		&deposit.Currency,
		&stripePI,
		&stripeCS,
		&deposit.Status,
		&failureReason,
		&deposit.CreatedAt,
		&deposit.UpdatedAt,
		&deposit.CompletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if userID != nil {
		deposit.UserID = *userID
	}
	if agentID != nil {
		deposit.AgentID = *agentID
	}
	if stripePI != nil {
		deposit.StripePaymentIntentID = *stripePI
	}
	if stripeCS != nil {
		deposit.StripeClientSecret = *stripeCS
	}
	if failureReason != nil {
		deposit.FailureReason = *failureReason
	}

	return deposit, nil
}

// UpdateDepositStatus updates the status of a deposit.
func (r *Repository) UpdateDepositStatus(ctx context.Context, id uuid.UUID, status DepositStatus, failureReason string) error {
	query := `
		UPDATE wallet_deposits
		SET status = $2, failure_reason = $3, updated_at = $4,
			completed_at = CASE WHEN $2 = 'completed' THEN $4 ELSE completed_at END
		WHERE id = $1
	`

	var reason *string
	if failureReason != "" {
		reason = &failureReason
	}

	_, err := r.pool.Exec(ctx, query, id, status, reason, time.Now())
	return err
}

// GetUserDeposits retrieves all deposits for a user.
func (r *Repository) GetUserDeposits(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Deposit, int, error) {
	countQuery := `SELECT COUNT(*) FROM wallet_deposits WHERE user_id = $1`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, user_id, agent_id, amount, currency, stripe_payment_intent_id,
			stripe_client_secret, status, failure_reason, created_at, updated_at, completed_at
		FROM wallet_deposits
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	return r.scanDeposits(ctx, query, userID, limit, offset, total)
}

// GetAgentDeposits retrieves all deposits for an agent.
func (r *Repository) GetAgentDeposits(ctx context.Context, agentID uuid.UUID, limit, offset int) ([]*Deposit, int, error) {
	countQuery := `SELECT COUNT(*) FROM wallet_deposits WHERE agent_id = $1`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, agentID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, user_id, agent_id, amount, currency, stripe_payment_intent_id,
			stripe_client_secret, status, failure_reason, created_at, updated_at, completed_at
		FROM wallet_deposits
		WHERE agent_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	return r.scanDeposits(ctx, query, agentID, limit, offset, total)
}

func (r *Repository) scanDeposits(ctx context.Context, query string, ownerID uuid.UUID, limit, offset, total int) ([]*Deposit, int, error) {
	rows, err := r.pool.Query(ctx, query, ownerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var deposits []*Deposit
	for rows.Next() {
		deposit := &Deposit{}
		var failureReason *string
		var stripePI, stripeCS *string
		var userID, agentID *uuid.UUID

		err := rows.Scan(
			&deposit.ID,
			&userID,
			&agentID,
			&deposit.Amount,
			&deposit.Currency,
			&stripePI,
			&stripeCS,
			&deposit.Status,
			&failureReason,
			&deposit.CreatedAt,
			&deposit.UpdatedAt,
			&deposit.CompletedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		if userID != nil {
			deposit.UserID = *userID
		}
		if agentID != nil {
			deposit.AgentID = *agentID
		}
		if stripePI != nil {
			deposit.StripePaymentIntentID = *stripePI
		}
		if stripeCS != nil {
			deposit.StripeClientSecret = *stripeCS
		}
		if failureReason != nil {
			deposit.FailureReason = *failureReason
		}

		deposits = append(deposits, deposit)
	}

	return deposits, total, nil
}

// GetCompletedDepositsTotal gets the total amount of completed deposits for a user.
func (r *Repository) GetCompletedDepositsTotal(ctx context.Context, userID uuid.UUID) (float64, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM wallet_deposits
		WHERE user_id = $1 AND status = 'completed'
	`

	var total float64
	err := r.pool.QueryRow(ctx, query, userID).Scan(&total)
	return total, err
}

// GetPendingDepositsTotal gets the total amount of pending deposits for a user.
func (r *Repository) GetPendingDepositsTotal(ctx context.Context, userID uuid.UUID) (float64, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM wallet_deposits
		WHERE user_id = $1 AND status IN ('pending', 'processing')
	`

	var total float64
	err := r.pool.QueryRow(ctx, query, userID).Scan(&total)
	return total, err
}

// GetAgentCompletedDepositsTotal gets the total amount of completed deposits for an agent.
func (r *Repository) GetAgentCompletedDepositsTotal(ctx context.Context, agentID uuid.UUID) (float64, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM wallet_deposits
		WHERE agent_id = $1 AND status = 'completed'
	`

	var total float64
	err := r.pool.QueryRow(ctx, query, agentID).Scan(&total)
	return total, err
}

// GetAgentPendingDepositsTotal gets the total amount of pending deposits for an agent.
func (r *Repository) GetAgentPendingDepositsTotal(ctx context.Context, agentID uuid.UUID) (float64, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM wallet_deposits
		WHERE agent_id = $1 AND status IN ('pending', 'processing')
	`

	var total float64
	err := r.pool.QueryRow(ctx, query, agentID).Scan(&total)
	return total, err
}
