package transaction

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

// mockRepository implements a mock repository for testing.
type mockRepository struct {
	transactions map[uuid.UUID]*Transaction
	escrows      map[uuid.UUID]*EscrowAccount
	ratings      map[uuid.UUID][]*Rating
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		transactions: make(map[uuid.UUID]*Transaction),
		escrows:      make(map[uuid.UUID]*EscrowAccount),
		ratings:      make(map[uuid.UUID][]*Rating),
	}
}

// mockPublisher implements EventPublisher for testing.
type mockPublisher struct {
	events []publishedEvent
}

type publishedEvent struct {
	eventType string
	payload   map[string]any
}

func (m *mockPublisher) Publish(ctx context.Context, eventType string, payload map[string]any) error {
	m.events = append(m.events, publishedEvent{eventType, payload})
	return nil
}

// mockPaymentService implements PaymentService for testing.
type mockPaymentService struct {
	paymentIntents map[string]bool
	captured       []string
	refunded       []string
}

func newMockPaymentService() *mockPaymentService {
	return &mockPaymentService{
		paymentIntents: make(map[string]bool),
	}
}

func (m *mockPaymentService) CreateEscrowPayment(ctx context.Context, transactionID, buyerID, sellerID string, amount float64, currency string) (string, string, error) {
	piID := "pi_test_" + transactionID[:8]
	m.paymentIntents[piID] = true
	return piID, piID + "_secret", nil
}

func (m *mockPaymentService) CapturePayment(ctx context.Context, paymentIntentID string) error {
	m.captured = append(m.captured, paymentIntentID)
	return nil
}

func (m *mockPaymentService) RefundPayment(ctx context.Context, paymentIntentID string) error {
	m.refunded = append(m.refunded, paymentIntentID)
	return nil
}

func TestTransactionStatus(t *testing.T) {
	tests := []struct {
		status   TransactionStatus
		expected string
	}{
		{StatusPending, "pending"},
		{StatusEscrowFunded, "escrow_funded"},
		{StatusDelivered, "delivered"},
		{StatusCompleted, "completed"},
		{StatusDisputed, "disputed"},
		{StatusRefunded, "refunded"},
	}

	for _, tt := range tests {
		if string(tt.status) != tt.expected {
			t.Errorf("expected status %s, got %s", tt.expected, tt.status)
		}
	}
}

func TestEscrowStatus(t *testing.T) {
	tests := []struct {
		status   EscrowStatus
		expected string
	}{
		{EscrowPending, "pending"},
		{EscrowFunded, "funded"},
		{EscrowReleased, "released"},
		{EscrowRefunded, "refunded"},
		{EscrowDisputed, "disputed"},
	}

	for _, tt := range tests {
		if string(tt.status) != tt.expected {
			t.Errorf("expected status %s, got %s", tt.expected, tt.status)
		}
	}
}

func TestEscrowFundingResult(t *testing.T) {
	result := &EscrowFundingResult{
		TransactionID:   uuid.New(),
		PaymentIntentID: "pi_test123",
		ClientSecret:    "pi_test123_secret",
		Amount:          100.00,
		Currency:        "USD",
	}

	if result.PaymentIntentID != "pi_test123" {
		t.Errorf("expected payment intent id pi_test123, got %s", result.PaymentIntentID)
	}

	if result.Amount != 100.00 {
		t.Errorf("expected amount 100.00, got %f", result.Amount)
	}
}

func TestCreateTransactionRequest(t *testing.T) {
	buyerID := uuid.New()
	sellerID := uuid.New()
	requestID := uuid.New()
	offerID := uuid.New()

	req := &CreateTransactionRequest{
		BuyerID:   buyerID,
		SellerID:  sellerID,
		RequestID: &requestID,
		OfferID:   &offerID,
		Amount:    50.00,
		Currency:  "USD",
	}

	if req.BuyerID != buyerID {
		t.Errorf("expected buyer id %s, got %s", buyerID, req.BuyerID)
	}

	if req.Amount != 50.00 {
		t.Errorf("expected amount 50.00, got %f", req.Amount)
	}
}

func TestSubmitRatingRequest(t *testing.T) {
	// Valid rating
	req := &SubmitRatingRequest{
		Score:   5,
		Comment: "Excellent service!",
	}

	if req.Score < 1 || req.Score > 5 {
		t.Error("rating score should be between 1 and 5")
	}

	// Test boundary values
	validScores := []int{1, 2, 3, 4, 5}
	for _, score := range validScores {
		if score < 1 || score > 5 {
			t.Errorf("score %d should be valid", score)
		}
	}
}

func TestDisputeRequest(t *testing.T) {
	req := &DisputeRequest{
		Reason:      "Item not as described",
		Description: "The data provided was incomplete and missing key fields.",
	}

	if req.Reason == "" {
		t.Error("dispute reason should not be empty")
	}
}

func TestListTransactionsParams(t *testing.T) {
	agentID := uuid.New()
	status := StatusPending

	params := ListTransactionsParams{
		AgentID: &agentID,
		Status:  &status,
		Role:    "buyer",
		Limit:   20,
		Offset:  0,
	}

	if *params.AgentID != agentID {
		t.Errorf("expected agent id %s, got %s", agentID, *params.AgentID)
	}

	if *params.Status != StatusPending {
		t.Errorf("expected status pending, got %s", *params.Status)
	}

	if params.Role != "buyer" {
		t.Errorf("expected role buyer, got %s", params.Role)
	}
}

func TestTransactionListResult(t *testing.T) {
	result := &TransactionListResult{
		Items:  []*Transaction{},
		Total:  0,
		Limit:  20,
		Offset: 0,
	}

	if result.Limit != 20 {
		t.Errorf("expected limit 20, got %d", result.Limit)
	}

	if len(result.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(result.Items))
	}
}

func TestTransaction(t *testing.T) {
	buyerID := uuid.New()
	sellerID := uuid.New()

	tx := &Transaction{
		ID:       uuid.New(),
		BuyerID:  buyerID,
		SellerID: sellerID,
		Amount:   100.00,
		Currency: "USD",
		Status:   StatusPending,
	}

	if tx.BuyerID != buyerID {
		t.Errorf("expected buyer id %s, got %s", buyerID, tx.BuyerID)
	}

	if tx.Status != StatusPending {
		t.Errorf("expected status pending, got %s", tx.Status)
	}
}

func TestEscrowAccount(t *testing.T) {
	txID := uuid.New()

	escrow := &EscrowAccount{
		ID:            uuid.New(),
		TransactionID: txID,
		Amount:        100.00,
		Currency:      "USD",
		Status:        EscrowPending,
	}

	if escrow.TransactionID != txID {
		t.Errorf("expected transaction id %s, got %s", txID, escrow.TransactionID)
	}

	if escrow.Status != EscrowPending {
		t.Errorf("expected status pending, got %s", escrow.Status)
	}
}

func TestRating(t *testing.T) {
	txID := uuid.New()
	raterID := uuid.New()
	ratedID := uuid.New()

	rating := &Rating{
		ID:            uuid.New(),
		TransactionID: txID,
		RaterID:       raterID,
		RatedAgentID:  ratedID,
		Score:         5,
		Comment:       "Great transaction!",
	}

	if rating.Score != 5 {
		t.Errorf("expected score 5, got %d", rating.Score)
	}

	if rating.RaterID != raterID {
		t.Errorf("expected rater id %s, got %s", raterID, rating.RaterID)
	}
}

func TestServiceErrors(t *testing.T) {
	// Test error messages
	if ErrInvalidStatus.Error() != "invalid transaction status for this operation" {
		t.Errorf("unexpected error message: %s", ErrInvalidStatus.Error())
	}

	if ErrNotAuthorized.Error() != "not authorized to perform this action" {
		t.Errorf("unexpected error message: %s", ErrNotAuthorized.Error())
	}

	if ErrInvalidRating.Error() != "rating score must be between 1 and 5" {
		t.Errorf("unexpected error message: %s", ErrInvalidRating.Error())
	}

	if ErrCannotRateYourself.Error() != "cannot rate yourself" {
		t.Errorf("unexpected error message: %s", ErrCannotRateYourself.Error())
	}

	if ErrTransactionNotReady.Error() != "transaction is not ready for this operation" {
		t.Errorf("unexpected error message: %s", ErrTransactionNotReady.Error())
	}
}
