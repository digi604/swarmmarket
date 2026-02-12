package transaction

import (
	"context"
	"errors"

	"github.com/digi604/swarmmarket/backend/pkg/logger"
	"github.com/google/uuid"
)

var (
	ErrInvalidStatus       = errors.New("invalid transaction status for this operation")
	ErrNotAuthorized       = errors.New("not authorized to perform this action")
	ErrInvalidRating       = errors.New("rating score must be between 1 and 5")
	ErrCannotRateYourself  = errors.New("cannot rate yourself")
	ErrTransactionNotReady = errors.New("transaction is not ready for this operation")
)

// EventPublisher publishes events to the notification system.
type EventPublisher interface {
	Publish(ctx context.Context, eventType string, payload map[string]any) error
}

// PaymentService handles payment operations.
type PaymentService interface {
	CreateEscrowPayment(ctx context.Context, transactionID, buyerID, sellerID string, amount float64, currency string) (paymentIntentID string, err error)
	CapturePayment(ctx context.Context, paymentIntentID string) error
	RefundPayment(ctx context.Context, paymentIntentID string) error
}

// TrustHandler handles trust score updates.
type TrustHandler interface {
	OnTransactionCompleted(ctx context.Context, agentID uuid.UUID, transactionID uuid.UUID) error
	OnRatingReceived(ctx context.Context, agentID uuid.UUID, ratingScore int, transactionID uuid.UUID) error
}

// Service handles transaction business logic.
type Service struct {
	repo      RepositoryInterface
	publisher EventPublisher
	payment   PaymentService
	trust     TrustHandler
}

// NewService creates a new transaction service.
func NewService(repo RepositoryInterface, publisher EventPublisher) *Service {
	return &Service{
		repo:      repo,
		publisher: publisher,
	}
}

// SetPaymentService sets the payment service (optional, for escrow).
func (s *Service) SetPaymentService(payment PaymentService) {
	s.payment = payment
}

// SetTrustHandler sets the trust handler (optional, for trust score updates).
func (s *Service) SetTrustHandler(trust TrustHandler) {
	s.trust = trust
}

// CreateFromOffer creates a transaction from an accepted offer (implements marketplace.TransactionCreator).
func (s *Service) CreateFromOffer(ctx context.Context, buyerID, sellerID uuid.UUID, requestID, offerID *uuid.UUID, amount float64, currency string) (uuid.UUID, error) {
	tx, err := s.CreateTransaction(ctx, &CreateTransactionRequest{
		BuyerID:   buyerID,
		SellerID:  sellerID,
		RequestID: requestID,
		OfferID:   offerID,
		Amount:    amount,
		Currency:  currency,
	})
	if err != nil {
		return uuid.Nil, err
	}
	return tx.ID, nil
}

// CreateFromListing creates a transaction from a listing purchase (implements marketplace.ListingTransactionCreator).
func (s *Service) CreateFromListing(ctx context.Context, buyerID, sellerID uuid.UUID, listingID *uuid.UUID, amount float64, currency string) (uuid.UUID, error) {
	tx, err := s.CreateTransaction(ctx, &CreateTransactionRequest{
		BuyerID:   buyerID,
		SellerID:  sellerID,
		ListingID: listingID,
		Amount:    amount,
		Currency:  currency,
	})
	if err != nil {
		return uuid.Nil, err
	}
	return tx.ID, nil
}

// CreateFromTask creates a transaction from an accepted task (implements task.TransactionCreator).
func (s *Service) CreateFromTask(ctx context.Context, requesterID, executorID uuid.UUID, taskID *uuid.UUID, amount float64, currency string) (uuid.UUID, error) {
	tx, err := s.CreateTransaction(ctx, &CreateTransactionRequest{
		BuyerID:  requesterID, // Requester pays
		SellerID: executorID,  // Executor receives
		TaskID:   taskID,
		Amount:   amount,
		Currency: currency,
	})
	if err != nil {
		return uuid.Nil, err
	}
	return tx.ID, nil
}

// CreateTransaction creates a new transaction (called when offer is accepted).
func (s *Service) CreateTransaction(ctx context.Context, req *CreateTransactionRequest) (*Transaction, error) {
	// Create the transaction
	tx, err := s.repo.CreateTransaction(ctx, req)
	if err != nil {
		return nil, err
	}

	// Create escrow account
	_, err = s.repo.CreateEscrowAccount(ctx, tx.ID, tx.Amount, tx.Currency)
	if err != nil {
		// Log but don't fail - escrow can be created later
	}

	// Log transaction creation
	logger.Info("transaction_created", map[string]interface{}{
		"transaction_id": tx.ID.String(),
		"buyer_id":       tx.BuyerID.String(),
		"seller_id":      tx.SellerID.String(),
		"amount":         tx.Amount,
		"currency":       tx.Currency,
		"status":         string(tx.Status),
	})

	// Publish event
	s.publishEvent(ctx, "transaction.created", map[string]any{
		"transaction_id": tx.ID,
		"buyer_id":       tx.BuyerID,
		"seller_id":      tx.SellerID,
		"amount":         tx.Amount,
		"currency":       tx.Currency,
	})
	s.publishEvent(ctx, "order.created", map[string]any{
		"transaction_id": tx.ID,
		"buyer_id":       tx.BuyerID,
		"seller_id":      tx.SellerID,
		"amount":         tx.Amount,
		"currency":       tx.Currency,
	})

	return tx, nil
}

// GetTransaction retrieves a transaction by ID.
func (s *Service) GetTransaction(ctx context.Context, id uuid.UUID) (*Transaction, error) {
	return s.repo.GetTransactionByID(ctx, id)
}

// ListTransactions retrieves transactions for an agent.
func (s *Service) ListTransactions(ctx context.Context, params ListTransactionsParams) (*TransactionListResult, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	return s.repo.ListTransactions(ctx, params)
}

// FundEscrow charges the buyer's saved payment method off-session.
func (s *Service) FundEscrow(ctx context.Context, transactionID, buyerID uuid.UUID) (*EscrowFundingResult, error) {
	tx, err := s.repo.GetTransactionByID(ctx, transactionID)
	if err != nil {
		return nil, err
	}

	if tx.BuyerID != buyerID {
		return nil, ErrNotAuthorized
	}

	if tx.Status != StatusPending {
		return nil, ErrInvalidStatus
	}

	if s.payment == nil {
		return nil, errors.New("payment service not configured")
	}

	paymentIntentID, err := s.payment.CreateEscrowPayment(
		ctx,
		transactionID.String(),
		tx.BuyerID.String(),
		tx.SellerID.String(),
		tx.Amount,
		tx.Currency,
	)
	if err != nil {
		return nil, err
	}

	escrow, err := s.repo.GetEscrowByTransactionID(ctx, transactionID)
	if err == nil {
		s.repo.UpdateEscrowPaymentIntent(ctx, escrow.ID, paymentIntentID)
	}

	return &EscrowFundingResult{
		TransactionID:   transactionID,
		PaymentIntentID: paymentIntentID,
		Amount:          tx.Amount,
		Currency:        tx.Currency,
	}, nil
}

// ConfirmEscrowFunded is called when payment succeeds (via webhook or confirmation).
func (s *Service) ConfirmEscrowFunded(ctx context.Context, transactionID uuid.UUID, paymentIntentID string) error {
	tx, err := s.repo.GetTransactionByID(ctx, transactionID)
	if err != nil {
		return err
	}

	// Update transaction status
	if err := s.repo.UpdateTransactionStatus(ctx, transactionID, StatusEscrowFunded); err != nil {
		return err
	}

	// Update escrow
	escrow, err := s.repo.GetEscrowByTransactionID(ctx, transactionID)
	if err == nil {
		s.repo.UpdateEscrowStatus(ctx, escrow.ID, EscrowFunded)
	}

	// Log escrow funded
	logger.Info("escrow_funded", map[string]interface{}{
		"transaction_id":    transactionID.String(),
		"buyer_id":          tx.BuyerID.String(),
		"seller_id":         tx.SellerID.String(),
		"amount":            tx.Amount,
		"currency":          tx.Currency,
		"payment_intent_id": paymentIntentID,
	})

	// Publish event
	s.publishEvent(ctx, "transaction.escrow_funded", map[string]any{
		"transaction_id":    transactionID,
		"buyer_id":          tx.BuyerID,
		"seller_id":         tx.SellerID,
		"amount":            tx.Amount,
		"currency":          tx.Currency,
		"payment_intent_id": paymentIntentID,
	})
	s.publishEvent(ctx, "escrow.funded", map[string]any{
		"transaction_id":    transactionID,
		"buyer_id":          tx.BuyerID,
		"seller_id":         tx.SellerID,
		"amount":            tx.Amount,
		"currency":          tx.Currency,
		"payment_intent_id": paymentIntentID,
	})

	return nil
}

// MarkDelivered marks a transaction as delivered by the seller.
// Only the seller can mark as delivered.
func (s *Service) MarkDelivered(ctx context.Context, transactionID, agentID uuid.UUID, deliveryProof, message string) (*Transaction, error) {
	// Get transaction
	tx, err := s.repo.GetTransactionByID(ctx, transactionID)
	if err != nil {
		return nil, err
	}

	// Only seller can mark as delivered
	if tx.SellerID != agentID {
		return nil, ErrNotAuthorized
	}

	// Check status - must be pending or escrow_funded
	if tx.Status != StatusPending && tx.Status != StatusEscrowFunded {
		return nil, ErrInvalidStatus
	}

	// Update transaction status to delivered
	if err := s.repo.UpdateTransactionStatus(ctx, transactionID, StatusDelivered); err != nil {
		return nil, err
	}

	// Get updated transaction
	tx, _ = s.repo.GetTransactionByID(ctx, transactionID)

	// Log delivery marked
	logger.Info("delivery_marked", map[string]interface{}{
		"transaction_id": transactionID.String(),
		"buyer_id":       tx.BuyerID.String(),
		"seller_id":      tx.SellerID.String(),
		"status":         string(tx.Status),
	})

	// Publish event
	s.publishEvent(ctx, "transaction.delivered", map[string]any{
		"transaction_id": transactionID,
		"buyer_id":       tx.BuyerID,
		"seller_id":      tx.SellerID,
		"delivery_proof": deliveryProof,
		"message":        message,
	})

	return tx, nil
}

// ConfirmDelivery confirms that goods/services have been delivered.
// Only the buyer can confirm delivery.
func (s *Service) ConfirmDelivery(ctx context.Context, transactionID, agentID uuid.UUID) (*Transaction, error) {
	// Get transaction
	tx, err := s.repo.GetTransactionByID(ctx, transactionID)
	if err != nil {
		return nil, err
	}

	// Only buyer can confirm delivery
	if tx.BuyerID != agentID {
		return nil, ErrNotAuthorized
	}

	// Check status - must be delivered (seller has marked it as delivered)
	if tx.Status != StatusDelivered {
		return nil, ErrInvalidStatus
	}

	// Update transaction
	if err := s.repo.ConfirmDelivery(ctx, transactionID); err != nil {
		return nil, err
	}

	// Get updated transaction
	tx, _ = s.repo.GetTransactionByID(ctx, transactionID)

	// Log delivery confirmed
	logger.Info("delivery_confirmed", map[string]interface{}{
		"transaction_id": transactionID.String(),
		"buyer_id":       tx.BuyerID.String(),
		"seller_id":      tx.SellerID.String(),
		"amount":         tx.Amount,
		"status":         string(tx.Status),
	})

	// Publish event
	s.publishEvent(ctx, "transaction.completed", map[string]any{
		"transaction_id": transactionID,
		"buyer_id":       tx.BuyerID,
		"seller_id":      tx.SellerID,
	})
	s.publishEvent(ctx, "delivery.confirmed", map[string]any{
		"transaction_id": transactionID,
		"buyer_id":       tx.BuyerID,
		"seller_id":      tx.SellerID,
	})

	return tx, nil
}

// CompleteTransaction marks a transaction as completed and releases escrow.
func (s *Service) CompleteTransaction(ctx context.Context, transactionID uuid.UUID) (*Transaction, error) {
	// Get transaction
	tx, err := s.repo.GetTransactionByID(ctx, transactionID)
	if err != nil {
		return nil, err
	}

	// Check status - must be delivered
	if tx.Status != StatusDelivered {
		return nil, ErrInvalidStatus
	}

	// Get escrow and capture payment if we have a payment intent
	escrow, err := s.repo.GetEscrowByTransactionID(ctx, transactionID)
	if err == nil && escrow.StripePaymentIntentID != "" && s.payment != nil {
		// Capture the held payment (releases funds to seller)
		if err := s.payment.CapturePayment(ctx, escrow.StripePaymentIntentID); err != nil {
			// Log error but don't fail - manual resolution needed
			s.publishEvent(ctx, "payment.capture_failed", map[string]any{
				"transaction_id":    transactionID,
				"payment_intent_id": escrow.StripePaymentIntentID,
				"error":             err.Error(),
			})
		}
		s.repo.UpdateEscrowStatus(ctx, escrow.ID, EscrowReleased)
	}

	// Complete transaction
	if err := s.repo.CompleteTransaction(ctx, transactionID); err != nil {
		return nil, err
	}

	// Update agent stats
	s.repo.UpdateAgentStats(ctx, tx.BuyerID, true)
	s.repo.UpdateAgentStats(ctx, tx.SellerID, true)

	// Update trust scores (async)
	if s.trust != nil {
		go s.trust.OnTransactionCompleted(context.Background(), tx.BuyerID, transactionID)
		go s.trust.OnTransactionCompleted(context.Background(), tx.SellerID, transactionID)
	}

	// Get updated transaction
	tx, _ = s.repo.GetTransactionByID(ctx, transactionID)

	// Publish event
	s.publishEvent(ctx, "transaction.completed", map[string]any{
		"transaction_id": transactionID,
		"buyer_id":       tx.BuyerID,
		"seller_id":      tx.SellerID,
		"amount":         tx.Amount,
	})
	s.publishEvent(ctx, "payment.released", map[string]any{
		"transaction_id": transactionID,
		"buyer_id":       tx.BuyerID,
		"seller_id":      tx.SellerID,
		"amount":         tx.Amount,
		"currency":       tx.Currency,
	})

	return tx, nil
}

// SubmitRating submits a rating for a completed transaction.
func (s *Service) SubmitRating(ctx context.Context, transactionID, raterID uuid.UUID, req *SubmitRatingRequest) (*Rating, error) {
	// Validate rating score
	if req.Score < 1 || req.Score > 5 {
		return nil, ErrInvalidRating
	}

	// Get transaction
	tx, err := s.repo.GetTransactionByID(ctx, transactionID)
	if err != nil {
		return nil, err
	}

	// Check transaction is completed or delivered
	if tx.Status != StatusCompleted && tx.Status != StatusDelivered {
		return nil, ErrTransactionNotReady
	}

	// Determine who is being rated
	var ratedAgentID uuid.UUID
	if raterID == tx.BuyerID {
		ratedAgentID = tx.SellerID // Buyer rates seller
	} else if raterID == tx.SellerID {
		ratedAgentID = tx.BuyerID // Seller rates buyer
	} else {
		return nil, ErrNotAuthorized
	}

	// Check hasn't already rated
	hasRated, err := s.repo.HasRated(ctx, transactionID, raterID)
	if err != nil {
		return nil, err
	}
	if hasRated {
		return nil, ErrRatingAlreadyExists
	}

	// Create rating
	rating := &Rating{
		TransactionID: transactionID,
		RaterID:       raterID,
		RatedAgentID:  ratedAgentID,
		Score:         req.Score,
		Comment:       req.Comment,
	}

	if err := s.repo.CreateRating(ctx, rating); err != nil {
		return nil, err
	}

	// Recalculate rated agent's average rating
	s.repo.RecalculateAgentRating(ctx, ratedAgentID)

	// Update trust score for rated agent (async)
	if s.trust != nil {
		go s.trust.OnRatingReceived(context.Background(), ratedAgentID, req.Score, transactionID)
	}

	// If transaction was delivered and both parties have rated, complete it
	if tx.Status == StatusDelivered {
		ratings, _ := s.repo.GetRatingsByTransactionID(ctx, transactionID)
		if len(ratings) >= 2 {
			s.CompleteTransaction(ctx, transactionID)
		}
	}

	// Publish event
	s.publishEvent(ctx, "rating.submitted", map[string]any{
		"transaction_id": transactionID,
		"rater_id":       raterID,
		"rated_agent_id": ratedAgentID,
		"score":          req.Score,
	})

	return rating, nil
}

// GetTransactionRatings retrieves all ratings for a transaction.
func (s *Service) GetTransactionRatings(ctx context.Context, transactionID uuid.UUID) ([]*Rating, error) {
	return s.repo.GetRatingsByTransactionID(ctx, transactionID)
}

// DisputeTransaction opens a dispute on a transaction.
func (s *Service) DisputeTransaction(ctx context.Context, transactionID, agentID uuid.UUID, req *DisputeRequest) (*Transaction, error) {
	// Get transaction
	tx, err := s.repo.GetTransactionByID(ctx, transactionID)
	if err != nil {
		return nil, err
	}

	// Must be buyer or seller
	if tx.BuyerID != agentID && tx.SellerID != agentID {
		return nil, ErrNotAuthorized
	}

	// Can only dispute certain statuses
	if tx.Status != StatusPending && tx.Status != StatusEscrowFunded && tx.Status != StatusDelivered {
		return nil, ErrInvalidStatus
	}

	// Update status
	if err := s.repo.UpdateTransactionStatus(ctx, transactionID, StatusDisputed); err != nil {
		return nil, err
	}

	// Update escrow
	escrow, err := s.repo.GetEscrowByTransactionID(ctx, transactionID)
	if err == nil {
		s.repo.UpdateEscrowStatus(ctx, escrow.ID, EscrowDisputed)
	}

	// Get updated transaction
	tx, _ = s.repo.GetTransactionByID(ctx, transactionID)

	// Publish event
	s.publishEvent(ctx, "dispute.opened", map[string]any{
		"transaction_id": transactionID,
		"opened_by":      agentID,
		"reason":         req.Reason,
	})

	return tx, nil
}

// RefundTransaction marks a transaction as refunded (called after Stripe refund).
func (s *Service) RefundTransaction(ctx context.Context, transactionID uuid.UUID) error {
	if err := s.repo.UpdateTransactionStatus(ctx, transactionID, StatusRefunded); err != nil {
		return err
	}

	// Update escrow
	escrow, err := s.repo.GetEscrowByTransactionID(ctx, transactionID)
	if err == nil {
		s.repo.UpdateEscrowStatus(ctx, escrow.ID, EscrowRefunded)
	}

	// Publish event
	s.publishEvent(ctx, "transaction.refunded", map[string]any{
		"transaction_id": transactionID,
	})

	return nil
}

// PublishPaymentFailed emits a payment failure event for a transaction.
func (s *Service) PublishPaymentFailed(ctx context.Context, tx *Transaction, paymentIntentID, reason string) {
	if tx == nil {
		return
	}

	s.publishEvent(ctx, "payment.failed", map[string]any{
		"transaction_id":    tx.ID,
		"buyer_id":          tx.BuyerID,
		"seller_id":         tx.SellerID,
		"amount":            tx.Amount,
		"currency":          tx.Currency,
		"payment_intent_id": paymentIntentID,
		"reason":            reason,
	})
}

// Helper to publish events asynchronously
func (s *Service) publishEvent(ctx context.Context, eventType string, payload map[string]any) {
	if s.publisher != nil {
		go s.publisher.Publish(ctx, eventType, payload)
	}
}
