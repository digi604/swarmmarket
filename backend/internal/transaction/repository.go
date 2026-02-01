package transaction

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrTransactionNotFound = errors.New("transaction not found")
	ErrRatingNotFound      = errors.New("rating not found")
	ErrRatingAlreadyExists = errors.New("rating already submitted for this transaction")
	ErrEscrowNotFound      = errors.New("escrow account not found")
)

// Repository handles transaction database operations.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new transaction repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// CreateTransaction creates a new transaction.
func (r *Repository) CreateTransaction(ctx context.Context, req *CreateTransactionRequest) (*Transaction, error) {
	tx := &Transaction{
		ID:        uuid.New(),
		BuyerID:   req.BuyerID,
		SellerID:  req.SellerID,
		ListingID: req.ListingID,
		RequestID: req.RequestID,
		OfferID:   req.OfferID,
		AuctionID: req.AuctionID,
		Amount:    req.Amount,
		Currency:  req.Currency,
		Status:    StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if tx.Currency == "" {
		tx.Currency = "USD"
	}

	query := `
		INSERT INTO transactions (id, buyer_id, seller_id, listing_id, request_id, offer_id, auction_id,
			amount, currency, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING platform_fee`

	err := r.pool.QueryRow(ctx, query,
		tx.ID, tx.BuyerID, tx.SellerID, tx.ListingID, tx.RequestID, tx.OfferID, tx.AuctionID,
		tx.Amount, tx.Currency, tx.Status, tx.CreatedAt, tx.UpdatedAt,
	).Scan(&tx.PlatformFee)

	if err != nil {
		return nil, err
	}

	return tx, nil
}

// GetTransactionByID retrieves a transaction by ID.
func (r *Repository) GetTransactionByID(ctx context.Context, id uuid.UUID) (*Transaction, error) {
	query := `
		SELECT t.id, t.buyer_id, t.seller_id, t.listing_id, t.request_id, t.offer_id, t.auction_id,
			t.amount, t.currency, t.platform_fee, t.status, t.delivery_confirmed_at, t.completed_at,
			t.metadata, t.created_at, t.updated_at,
			COALESCE(b.name, '') as buyer_name, COALESCE(s.name, '') as seller_name
		FROM transactions t
		LEFT JOIN agents b ON t.buyer_id = b.id
		LEFT JOIN agents s ON t.seller_id = s.id
		WHERE t.id = $1`

	tx := &Transaction{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&tx.ID, &tx.BuyerID, &tx.SellerID, &tx.ListingID, &tx.RequestID, &tx.OfferID, &tx.AuctionID,
		&tx.Amount, &tx.Currency, &tx.PlatformFee, &tx.Status, &tx.DeliveryConfirmedAt, &tx.CompletedAt,
		&tx.Metadata, &tx.CreatedAt, &tx.UpdatedAt,
		&tx.BuyerName, &tx.SellerName,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTransactionNotFound
		}
		return nil, err
	}

	return tx, nil
}

// ListTransactions retrieves transactions for an agent.
func (r *Repository) ListTransactions(ctx context.Context, params ListTransactionsParams) (*TransactionListResult, error) {
	var conditions []string
	var args []interface{}
	argNum := 1

	if params.AgentID != nil {
		switch params.Role {
		case "buyer":
			conditions = append(conditions, "t.buyer_id = $"+string(rune('0'+argNum)))
		case "seller":
			conditions = append(conditions, "t.seller_id = $"+string(rune('0'+argNum)))
		default:
			conditions = append(conditions, "(t.buyer_id = $"+string(rune('0'+argNum))+" OR t.seller_id = $"+string(rune('0'+argNum))+")")
		}
		args = append(args, *params.AgentID)
		argNum++
	}

	if params.Status != nil {
		conditions = append(conditions, "t.status = $"+string(rune('0'+argNum)))
		args = append(args, *params.Status)
		argNum++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			whereClause += " AND " + conditions[i]
		}
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM transactions t " + whereClause
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	// Get items
	query := `
		SELECT t.id, t.buyer_id, t.seller_id, t.listing_id, t.request_id, t.offer_id, t.auction_id,
			t.amount, t.currency, t.platform_fee, t.status, t.delivery_confirmed_at, t.completed_at,
			t.metadata, t.created_at, t.updated_at,
			COALESCE(b.name, '') as buyer_name, COALESCE(s.name, '') as seller_name
		FROM transactions t
		LEFT JOIN agents b ON t.buyer_id = b.id
		LEFT JOIN agents s ON t.seller_id = s.id
		` + whereClause + `
		ORDER BY t.created_at DESC
		LIMIT $` + string(rune('0'+argNum)) + ` OFFSET $` + string(rune('0'+argNum+1))

	args = append(args, params.Limit, params.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*Transaction
	for rows.Next() {
		tx := &Transaction{}
		err := rows.Scan(
			&tx.ID, &tx.BuyerID, &tx.SellerID, &tx.ListingID, &tx.RequestID, &tx.OfferID, &tx.AuctionID,
			&tx.Amount, &tx.Currency, &tx.PlatformFee, &tx.Status, &tx.DeliveryConfirmedAt, &tx.CompletedAt,
			&tx.Metadata, &tx.CreatedAt, &tx.UpdatedAt,
			&tx.BuyerName, &tx.SellerName,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, tx)
	}

	return &TransactionListResult{
		Items:  items,
		Total:  total,
		Limit:  params.Limit,
		Offset: params.Offset,
	}, nil
}

// UpdateTransactionStatus updates a transaction's status.
func (r *Repository) UpdateTransactionStatus(ctx context.Context, id uuid.UUID, status TransactionStatus) error {
	query := `UPDATE transactions SET status = $1, updated_at = NOW() WHERE id = $2`
	result, err := r.pool.Exec(ctx, query, status, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrTransactionNotFound
	}
	return nil
}

// ConfirmDelivery marks a transaction as completed (buyer confirms receipt).
func (r *Repository) ConfirmDelivery(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE transactions
		SET status = $1, delivery_confirmed_at = NOW(), updated_at = NOW()
		WHERE id = $2`
	result, err := r.pool.Exec(ctx, query, StatusCompleted, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrTransactionNotFound
	}
	return nil
}

// CompleteTransaction marks a transaction as completed.
func (r *Repository) CompleteTransaction(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE transactions
		SET status = $1, completed_at = NOW(), updated_at = NOW()
		WHERE id = $2`
	result, err := r.pool.Exec(ctx, query, StatusCompleted, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrTransactionNotFound
	}
	return nil
}

// --- Escrow Operations ---

// CreateEscrowAccount creates an escrow account for a transaction.
func (r *Repository) CreateEscrowAccount(ctx context.Context, transactionID uuid.UUID, amount float64, currency string) (*EscrowAccount, error) {
	escrow := &EscrowAccount{
		ID:            uuid.New(),
		TransactionID: transactionID,
		Amount:        amount,
		Currency:      currency,
		Status:        EscrowPending,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	query := `
		INSERT INTO escrow_accounts (id, transaction_id, amount, currency, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.pool.Exec(ctx, query,
		escrow.ID, escrow.TransactionID, escrow.Amount, escrow.Currency, escrow.Status,
		escrow.CreatedAt, escrow.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return escrow, nil
}

// GetEscrowByTransactionID retrieves the escrow account for a transaction.
func (r *Repository) GetEscrowByTransactionID(ctx context.Context, transactionID uuid.UUID) (*EscrowAccount, error) {
	query := `
		SELECT id, transaction_id, amount, currency, status, funded_at, released_at,
			stripe_payment_intent_id, metadata, created_at, updated_at
		FROM escrow_accounts
		WHERE transaction_id = $1`

	escrow := &EscrowAccount{}
	err := r.pool.QueryRow(ctx, query, transactionID).Scan(
		&escrow.ID, &escrow.TransactionID, &escrow.Amount, &escrow.Currency, &escrow.Status,
		&escrow.FundedAt, &escrow.ReleasedAt, &escrow.StripePaymentIntentID,
		&escrow.Metadata, &escrow.CreatedAt, &escrow.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrEscrowNotFound
		}
		return nil, err
	}

	return escrow, nil
}

// UpdateEscrowStatus updates an escrow account's status.
func (r *Repository) UpdateEscrowStatus(ctx context.Context, id uuid.UUID, status EscrowStatus) error {
	var updateField string
	switch status {
	case EscrowFunded:
		updateField = ", funded_at = NOW()"
	case EscrowReleased:
		updateField = ", released_at = NOW()"
	}

	query := `UPDATE escrow_accounts SET status = $1, updated_at = NOW()` + updateField + ` WHERE id = $2`
	result, err := r.pool.Exec(ctx, query, status, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrEscrowNotFound
	}
	return nil
}

// UpdateEscrowPaymentIntent updates the Stripe payment intent ID on an escrow.
func (r *Repository) UpdateEscrowPaymentIntent(ctx context.Context, id uuid.UUID, paymentIntentID string) error {
	query := `UPDATE escrow_accounts SET stripe_payment_intent_id = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, paymentIntentID, id)
	return err
}

// --- Rating Operations ---

// CreateRating creates a rating for a transaction.
func (r *Repository) CreateRating(ctx context.Context, rating *Rating) error {
	rating.ID = uuid.New()
	rating.CreatedAt = time.Now()

	query := `
		INSERT INTO ratings (id, transaction_id, rater_id, rated_agent_id, score, comment, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.pool.Exec(ctx, query,
		rating.ID, rating.TransactionID, rating.RaterID, rating.RatedAgentID,
		rating.Score, rating.Comment, rating.CreatedAt,
	)
	if err != nil {
		// Check for unique constraint violation
		if err.Error() == "ERROR: duplicate key value violates unique constraint \"ratings_transaction_id_rater_id_key\" (SQLSTATE 23505)" {
			return ErrRatingAlreadyExists
		}
		return err
	}

	return nil
}

// GetRatingsByTransactionID retrieves all ratings for a transaction.
func (r *Repository) GetRatingsByTransactionID(ctx context.Context, transactionID uuid.UUID) ([]*Rating, error) {
	query := `
		SELECT r.id, r.transaction_id, r.rater_id, r.rated_agent_id, r.score, r.comment, r.created_at,
			COALESCE(a.name, '') as rater_name
		FROM ratings r
		LEFT JOIN agents a ON r.rater_id = a.id
		WHERE r.transaction_id = $1
		ORDER BY r.created_at DESC`

	rows, err := r.pool.Query(ctx, query, transactionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ratings []*Rating
	for rows.Next() {
		rating := &Rating{}
		err := rows.Scan(
			&rating.ID, &rating.TransactionID, &rating.RaterID, &rating.RatedAgentID,
			&rating.Score, &rating.Comment, &rating.CreatedAt, &rating.RaterName,
		)
		if err != nil {
			return nil, err
		}
		ratings = append(ratings, rating)
	}

	return ratings, nil
}

// HasRated checks if an agent has already rated a transaction.
func (r *Repository) HasRated(ctx context.Context, transactionID, raterID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM ratings WHERE transaction_id = $1 AND rater_id = $2)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, transactionID, raterID).Scan(&exists)
	return exists, err
}

// UpdateAgentStats updates an agent's transaction stats after a completed transaction.
func (r *Repository) UpdateAgentStats(ctx context.Context, agentID uuid.UUID, successful bool) error {
	var query string
	if successful {
		query = `
			UPDATE agents
			SET total_transactions = total_transactions + 1,
				successful_trades = successful_trades + 1,
				updated_at = NOW()
			WHERE id = $1`
	} else {
		query = `
			UPDATE agents
			SET total_transactions = total_transactions + 1,
				updated_at = NOW()
			WHERE id = $1`
	}

	_, err := r.pool.Exec(ctx, query, agentID)
	return err
}

// RecalculateAgentRating recalculates an agent's average rating.
func (r *Repository) RecalculateAgentRating(ctx context.Context, agentID uuid.UUID) error {
	query := `
		UPDATE agents
		SET average_rating = COALESCE((
			SELECT AVG(score)::DECIMAL(3,2) FROM ratings WHERE rated_agent_id = $1
		), 0),
		updated_at = NOW()
		WHERE id = $1`

	_, err := r.pool.Exec(ctx, query, agentID)
	return err
}
