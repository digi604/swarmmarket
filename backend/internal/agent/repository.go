package agent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrAgentNotFound = errors.New("agent not found")
	ErrDuplicateKey  = errors.New("duplicate api key")
)

// Repository handles agent data persistence.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new agent repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create inserts a new agent into the database.
func (r *Repository) Create(ctx context.Context, agent *Agent) error {
	query := `
		INSERT INTO agents (
			id, name, description, owner_email, api_key_hash, api_key_prefix,
			verification_level, trust_score, total_transactions, successful_trades,
			average_rating, is_active, metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
	`

	_, err := r.pool.Exec(ctx, query,
		agent.ID,
		agent.Name,
		agent.Description,
		agent.OwnerEmail,
		agent.APIKeyHash,
		agent.APIKeyPrefix,
		agent.VerificationLevel,
		agent.TrustScore,
		agent.TotalTransactions,
		agent.SuccessfulTrades,
		agent.AverageRating,
		agent.IsActive,
		agent.Metadata,
		agent.CreatedAt,
		agent.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}

	return nil
}

// GetByID retrieves an agent by ID.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Agent, error) {
	query := `
		SELECT id, name, description, owner_email, api_key_hash, api_key_prefix,
			verification_level, trust_score, total_transactions, successful_trades,
			average_rating, is_active, metadata, created_at, updated_at, last_seen_at
		FROM agents
		WHERE id = $1
	`

	agent := &Agent{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&agent.ID,
		&agent.Name,
		&agent.Description,
		&agent.OwnerEmail,
		&agent.APIKeyHash,
		&agent.APIKeyPrefix,
		&agent.VerificationLevel,
		&agent.TrustScore,
		&agent.TotalTransactions,
		&agent.SuccessfulTrades,
		&agent.AverageRating,
		&agent.IsActive,
		&agent.Metadata,
		&agent.CreatedAt,
		&agent.UpdatedAt,
		&agent.LastSeenAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAgentNotFound
		}
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	return agent, nil
}

// GetByAPIKeyHash retrieves an agent by API key hash.
func (r *Repository) GetByAPIKeyHash(ctx context.Context, hash string) (*Agent, error) {
	query := `
		SELECT id, name, description, owner_email, api_key_hash, api_key_prefix,
			verification_level, trust_score, total_transactions, successful_trades,
			average_rating, is_active, metadata, created_at, updated_at, last_seen_at
		FROM agents
		WHERE api_key_hash = $1 AND is_active = true
	`

	agent := &Agent{}
	err := r.pool.QueryRow(ctx, query, hash).Scan(
		&agent.ID,
		&agent.Name,
		&agent.Description,
		&agent.OwnerEmail,
		&agent.APIKeyHash,
		&agent.APIKeyPrefix,
		&agent.VerificationLevel,
		&agent.TrustScore,
		&agent.TotalTransactions,
		&agent.SuccessfulTrades,
		&agent.AverageRating,
		&agent.IsActive,
		&agent.Metadata,
		&agent.CreatedAt,
		&agent.UpdatedAt,
		&agent.LastSeenAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAgentNotFound
		}
		return nil, fmt.Errorf("failed to get agent by api key: %w", err)
	}

	return agent, nil
}

// Update updates an existing agent.
func (r *Repository) Update(ctx context.Context, agent *Agent) error {
	query := `
		UPDATE agents
		SET name = $2, description = $3, metadata = $4, updated_at = $5
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		agent.ID,
		agent.Name,
		agent.Description,
		agent.Metadata,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("failed to update agent: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrAgentNotFound
	}

	return nil
}

// UpdateLastSeen updates the agent's last seen timestamp.
func (r *Repository) UpdateLastSeen(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE agents SET last_seen_at = $2 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id, time.Now().UTC())
	return err
}

// Deactivate deactivates an agent (soft delete).
func (r *Repository) Deactivate(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE agents SET is_active = false, updated_at = $2 WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to deactivate agent: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrAgentNotFound
	}

	return nil
}

// GetReputation retrieves the reputation details for an agent.
func (r *Repository) GetReputation(ctx context.Context, agentID uuid.UUID) (*Reputation, error) {
	// First verify agent exists
	agent, err := r.GetByID(ctx, agentID)
	if err != nil {
		return nil, err
	}

	rep := &Reputation{
		AgentID:           agent.ID,
		TrustScore:        agent.TrustScore,
		TotalTransactions: agent.TotalTransactions,
		SuccessfulTrades:  agent.SuccessfulTrades,
		AverageRating:     agent.AverageRating,
	}

	// Get recent ratings
	ratingsQuery := `
		SELECT transaction_id, rater_id, score, comment, created_at
		FROM ratings
		WHERE rated_agent_id = $1
		ORDER BY created_at DESC
		LIMIT 10
	`

	rows, err := r.pool.Query(ctx, ratingsQuery, agentID)
	if err != nil {
		// Ratings table might not exist yet, continue without ratings
		return rep, nil
	}
	defer rows.Close()

	for rows.Next() {
		var rating Rating
		if err := rows.Scan(
			&rating.TransactionID,
			&rating.RaterID,
			&rating.Score,
			&rating.Comment,
			&rating.CreatedAt,
		); err != nil {
			continue
		}
		rep.RecentRatings = append(rep.RecentRatings, rating)
	}

	rep.RatingCount = len(rep.RecentRatings)

	return rep, nil
}
