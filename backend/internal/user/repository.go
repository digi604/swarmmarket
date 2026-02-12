package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

// Repository handles user database operations.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new user repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const userColumns = `id, clerk_user_id, email, name, avatar_url,
	COALESCE(stripe_connect_account_id, ''), COALESCE(stripe_connect_charges_enabled, false),
	COALESCE(stripe_customer_id, ''), COALESCE(stripe_default_payment_method_id, ''),
	created_at, updated_at`

func scanUser(row pgx.Row) (*User, error) {
	var u User
	err := row.Scan(
		&u.ID,
		&u.ClerkUserID,
		&u.Email,
		&u.Name,
		&u.AvatarURL,
		&u.StripeConnectAccountID,
		&u.StripeConnectChargesEnabled,
		&u.StripeCustomerID,
		&u.StripeDefaultPaymentMethodID,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	return &u, err
}

// CreateUser creates a new user in the database.
func (r *Repository) CreateUser(ctx context.Context, user *User) (*User, error) {
	query := `
		INSERT INTO users (clerk_user_id, email, name, avatar_url)
		VALUES ($1, $2, $3, $4)
		RETURNING ` + userColumns

	u, err := scanUser(r.pool.QueryRow(ctx, query,
		user.ClerkUserID,
		user.Email,
		user.Name,
		user.AvatarURL,
	))
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return u, nil
}

// GetUserByID retrieves a user by their internal ID.
func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	query := `SELECT ` + userColumns + ` FROM users WHERE id = $1`

	u, err := scanUser(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return u, nil
}

// GetUserByClerkID retrieves a user by their Clerk user ID.
func (r *Repository) GetUserByClerkID(ctx context.Context, clerkUserID string) (*User, error) {
	query := `SELECT ` + userColumns + ` FROM users WHERE clerk_user_id = $1`

	u, err := scanUser(r.pool.QueryRow(ctx, query, clerkUserID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by clerk id: %w", err)
	}
	return u, nil
}

// UpdateUser updates a user's profile information.
func (r *Repository) UpdateUser(ctx context.Context, user *User) error {
	query := `
		UPDATE users
		SET email = $2, name = $3, avatar_url = $4, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		user.ID,
		user.Email,
		user.Name,
		user.AvatarURL,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

// UpsertUser creates or updates a user based on Clerk ID.
func (r *Repository) UpsertUser(ctx context.Context, user *User) (*User, error) {
	query := `
		INSERT INTO users (clerk_user_id, email, name, avatar_url)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (clerk_user_id) DO UPDATE
		SET email = EXCLUDED.email,
		    name = EXCLUDED.name,
		    avatar_url = EXCLUDED.avatar_url,
		    updated_at = NOW()
		RETURNING ` + userColumns

	u, err := scanUser(r.pool.QueryRow(ctx, query,
		user.ClerkUserID,
		user.Email,
		user.Name,
		user.AvatarURL,
	))
	if err != nil {
		return nil, fmt.Errorf("failed to upsert user: %w", err)
	}
	return u, nil
}

// SetStripeConnectAccountID stores the Stripe Connect account ID for a user.
func (r *Repository) SetStripeConnectAccountID(ctx context.Context, userID uuid.UUID, accountID string) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE users SET stripe_connect_account_id = $2, updated_at = NOW() WHERE id = $1`,
		userID, accountID,
	)
	if err != nil {
		return fmt.Errorf("failed to set stripe connect account id: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

// SetStripeConnectChargesEnabled updates the cached charges_enabled flag.
func (r *Repository) SetStripeConnectChargesEnabled(ctx context.Context, userID uuid.UUID, enabled bool) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE users SET stripe_connect_charges_enabled = $2, updated_at = NOW() WHERE id = $1`,
		userID, enabled,
	)
	if err != nil {
		return fmt.Errorf("failed to set stripe connect charges enabled: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

// GetUserByStripeConnectAccountID looks up a user by their Connect account ID (for webhooks).
func (r *Repository) GetUserByStripeConnectAccountID(ctx context.Context, accountID string) (*User, error) {
	query := `SELECT ` + userColumns + ` FROM users WHERE stripe_connect_account_id = $1`

	u, err := scanUser(r.pool.QueryRow(ctx, query, accountID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by stripe connect account id: %w", err)
	}
	return u, nil
}

// GetConnectAccountIDForAgent resolves agent → owner → stripe_connect_account_id.
// Returns empty string if agent has no owner or owner has no Connect account.
func (r *Repository) GetConnectAccountIDForAgent(ctx context.Context, agentID uuid.UUID) (string, error) {
	var accountID *string
	err := r.pool.QueryRow(ctx, `
		SELECT u.stripe_connect_account_id
		FROM agents a
		JOIN users u ON u.id = a.owner_user_id
		WHERE a.id = $1
		  AND u.stripe_connect_charges_enabled = true
	`, agentID).Scan(&accountID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("failed to resolve connect account for agent: %w", err)
	}
	if accountID == nil {
		return "", nil
	}
	return *accountID, nil
}

// SetStripeCustomerID stores the Stripe Customer ID for a user.
func (r *Repository) SetStripeCustomerID(ctx context.Context, userID uuid.UUID, customerID string) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE users SET stripe_customer_id = $2, updated_at = NOW() WHERE id = $1`,
		userID, customerID,
	)
	if err != nil {
		return fmt.Errorf("failed to set stripe customer id: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

// SetStripeDefaultPaymentMethod updates the default payment method for a Stripe customer.
func (r *Repository) SetStripeDefaultPaymentMethod(ctx context.Context, customerID, paymentMethodID string) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE users SET stripe_default_payment_method_id = $2, updated_at = NOW() WHERE stripe_customer_id = $1`,
		customerID, paymentMethodID,
	)
	if err != nil {
		return fmt.Errorf("failed to set default payment method: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

// GetPaymentMethodForAgent resolves agent → owner → stripe customer + payment method.
func (r *Repository) GetPaymentMethodForAgent(ctx context.Context, agentID uuid.UUID) (customerID, pmID string, ownerUserID uuid.UUID, err error) {
	err = r.pool.QueryRow(ctx, `
		SELECT u.stripe_customer_id, u.stripe_default_payment_method_id, u.id
		FROM agents a
		JOIN users u ON u.id = a.owner_user_id
		WHERE a.id = $1
		  AND u.stripe_customer_id IS NOT NULL
		  AND u.stripe_default_payment_method_id IS NOT NULL
	`, agentID).Scan(&customerID, &pmID, &ownerUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", uuid.Nil, nil
		}
		return "", "", uuid.Nil, fmt.Errorf("failed to resolve payment method for agent: %w", err)
	}
	return customerID, pmID, ownerUserID, nil
}

// GetOwnedAgents retrieves all agents owned by a user.
func (r *Repository) GetOwnedAgents(ctx context.Context, userID uuid.UUID) ([]*OwnedAgentSummary, error) {
	query := `
		SELECT id, name, description, trust_score, total_transactions,
		       average_rating, is_active, created_at, last_seen_at
		FROM agents
		WHERE owner_user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get owned agents: %w", err)
	}
	defer rows.Close()

	var agents []*OwnedAgentSummary
	for rows.Next() {
		var a OwnedAgentSummary
		err := rows.Scan(
			&a.ID,
			&a.Name,
			&a.Description,
			&a.TrustScore,
			&a.TotalTransactions,
			&a.AverageRating,
			&a.IsActive,
			&a.CreatedAt,
			&a.LastSeenAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent: %w", err)
		}
		agents = append(agents, &a)
	}

	return agents, nil
}

// GetAgentMetrics retrieves detailed metrics for an owned agent.
func (r *Repository) GetAgentMetrics(ctx context.Context, userID, agentID uuid.UUID) (*AgentMetrics, error) {
	// First verify ownership
	var ownerID *uuid.UUID
	err := r.pool.QueryRow(ctx,
		`SELECT owner_user_id FROM agents WHERE id = $1`,
		agentID,
	).Scan(&ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("agent not found")
		}
		return nil, fmt.Errorf("failed to check agent ownership: %w", err)
	}

	if ownerID == nil || *ownerID != userID {
		return nil, errors.New("not authorized to view this agent's metrics")
	}

	// Get agent base info
	var m AgentMetrics
	err = r.pool.QueryRow(ctx, `
		SELECT id, name, trust_score, total_transactions, successful_trades,
		       average_rating, created_at, last_seen_at
		FROM agents
		WHERE id = $1
	`, agentID).Scan(
		&m.AgentID,
		&m.AgentName,
		&m.TrustScore,
		&m.TotalTransactions,
		&m.SuccessfulTrades,
		&m.AverageRating,
		&m.CreatedAt,
		&m.LastSeenAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent metrics: %w", err)
	}

	// Calculate failed trades
	m.FailedTrades = m.TotalTransactions - m.SuccessfulTrades

	// Get revenue (as seller)
	err = r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE seller_id = $1 AND status = 'completed'
	`, agentID).Scan(&m.TotalRevenue)
	if err != nil {
		return nil, fmt.Errorf("failed to get revenue: %w", err)
	}

	// Get spending (as buyer)
	err = r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE buyer_id = $1 AND status = 'completed'
	`, agentID).Scan(&m.TotalSpent)
	if err != nil {
		return nil, fmt.Errorf("failed to get spending: %w", err)
	}

	// Get active listings count
	err = r.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM listings
		WHERE seller_id = $1 AND status = 'active'
	`, agentID).Scan(&m.ActiveListings)
	if err != nil {
		return nil, fmt.Errorf("failed to get active listings: %w", err)
	}

	// Get active requests count
	err = r.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM requests
		WHERE requester_id = $1 AND status = 'open'
	`, agentID).Scan(&m.ActiveRequests)
	if err != nil {
		return nil, fmt.Errorf("failed to get active requests: %w", err)
	}

	// Get pending offers count (offers made by this agent)
	err = r.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM offers
		WHERE offerer_id = $1 AND status = 'pending'
	`, agentID).Scan(&m.PendingOffers)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending offers: %w", err)
	}

	// Get active auctions count
	err = r.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM auctions
		WHERE seller_id = $1 AND status = 'active'
	`, agentID).Scan(&m.ActiveAuctions)
	if err != nil {
		return nil, fmt.Errorf("failed to get active auctions: %w", err)
	}

	return &m, nil
}
