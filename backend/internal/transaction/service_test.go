package transaction

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func strPtr(s string) *string { return &s }

// mockRepository implements RepositoryInterface for testing.
type mockRepository struct {
	transactions    map[uuid.UUID]*Transaction
	escrows         map[uuid.UUID]*EscrowAccount
	ratings         map[uuid.UUID][]*Rating
	agentRatings    map[uuid.UUID]float64
	agentStats      map[uuid.UUID]struct{ total, successful int }
	createErr       error
	getByIDErr      error
	listErr         error
	updateStatusErr error
	escrowErr       error
	ratingErr       error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		transactions: make(map[uuid.UUID]*Transaction),
		escrows:      make(map[uuid.UUID]*EscrowAccount),
		ratings:      make(map[uuid.UUID][]*Rating),
		agentRatings: make(map[uuid.UUID]float64),
		agentStats:   make(map[uuid.UUID]struct{ total, successful int }),
	}
}

// Verify mockRepository implements RepositoryInterface
var _ RepositoryInterface = (*mockRepository)(nil)

func (m *mockRepository) CreateTransaction(ctx context.Context, req *CreateTransactionRequest) (*Transaction, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
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
	m.transactions[tx.ID] = tx
	return tx, nil
}

func (m *mockRepository) GetTransactionByID(ctx context.Context, id uuid.UUID) (*Transaction, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	tx, ok := m.transactions[id]
	if !ok {
		return nil, ErrTransactionNotFound
	}
	return tx, nil
}

func (m *mockRepository) ListTransactions(ctx context.Context, params ListTransactionsParams) (*TransactionListResult, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var items []*Transaction
	for _, tx := range m.transactions {
		if params.AgentID != nil {
			if params.Role == "buyer" && tx.BuyerID != *params.AgentID {
				continue
			}
			if params.Role == "seller" && tx.SellerID != *params.AgentID {
				continue
			}
			if params.Role == "" && tx.BuyerID != *params.AgentID && tx.SellerID != *params.AgentID {
				continue
			}
		}
		if params.Status != nil && tx.Status != *params.Status {
			continue
		}
		items = append(items, tx)
	}
	return &TransactionListResult{
		Items:  items,
		Total:  len(items),
		Limit:  params.Limit,
		Offset: params.Offset,
	}, nil
}

func (m *mockRepository) UpdateTransactionStatus(ctx context.Context, id uuid.UUID, status TransactionStatus) error {
	if m.updateStatusErr != nil {
		return m.updateStatusErr
	}
	tx, ok := m.transactions[id]
	if !ok {
		return ErrTransactionNotFound
	}
	tx.Status = status
	tx.UpdatedAt = time.Now()
	return nil
}

func (m *mockRepository) ConfirmDelivery(ctx context.Context, id uuid.UUID) error {
	tx, ok := m.transactions[id]
	if !ok {
		return ErrTransactionNotFound
	}
	tx.Status = StatusCompleted
	now := time.Now()
	tx.DeliveryConfirmedAt = &now
	tx.UpdatedAt = now
	return nil
}

func (m *mockRepository) CompleteTransaction(ctx context.Context, id uuid.UUID) error {
	tx, ok := m.transactions[id]
	if !ok {
		return ErrTransactionNotFound
	}
	tx.Status = StatusCompleted
	now := time.Now()
	tx.CompletedAt = &now
	tx.UpdatedAt = now
	return nil
}

func (m *mockRepository) CreateEscrowAccount(ctx context.Context, transactionID uuid.UUID, amount float64, currency string) (*EscrowAccount, error) {
	if m.escrowErr != nil {
		return nil, m.escrowErr
	}
	escrow := &EscrowAccount{
		ID:            uuid.New(),
		TransactionID: transactionID,
		Amount:        amount,
		Currency:      currency,
		Status:        EscrowPending,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	m.escrows[transactionID] = escrow
	return escrow, nil
}

func (m *mockRepository) GetEscrowByTransactionID(ctx context.Context, transactionID uuid.UUID) (*EscrowAccount, error) {
	escrow, ok := m.escrows[transactionID]
	if !ok {
		return nil, ErrEscrowNotFound
	}
	return escrow, nil
}

func (m *mockRepository) UpdateEscrowStatus(ctx context.Context, id uuid.UUID, status EscrowStatus) error {
	for _, escrow := range m.escrows {
		if escrow.ID == id {
			escrow.Status = status
			escrow.UpdatedAt = time.Now()
			if status == EscrowFunded {
				now := time.Now()
				escrow.FundedAt = &now
			}
			if status == EscrowReleased {
				now := time.Now()
				escrow.ReleasedAt = &now
			}
			return nil
		}
	}
	return ErrEscrowNotFound
}

func (m *mockRepository) UpdateEscrowPaymentIntent(ctx context.Context, id uuid.UUID, paymentIntentID string) error {
	for _, escrow := range m.escrows {
		if escrow.ID == id {
			escrow.StripePaymentIntentID = &paymentIntentID
			escrow.UpdatedAt = time.Now()
			return nil
		}
	}
	return nil
}

func (m *mockRepository) CreateRating(ctx context.Context, rating *Rating) error {
	if m.ratingErr != nil {
		return m.ratingErr
	}
	ratings := m.ratings[rating.TransactionID]
	for _, r := range ratings {
		if r.RaterID == rating.RaterID {
			return ErrRatingAlreadyExists
		}
	}
	rating.ID = uuid.New()
	rating.CreatedAt = time.Now()
	m.ratings[rating.TransactionID] = append(m.ratings[rating.TransactionID], rating)
	return nil
}

func (m *mockRepository) GetRatingsByTransactionID(ctx context.Context, transactionID uuid.UUID) ([]*Rating, error) {
	return m.ratings[transactionID], nil
}

func (m *mockRepository) HasRated(ctx context.Context, transactionID, raterID uuid.UUID) (bool, error) {
	ratings := m.ratings[transactionID]
	for _, r := range ratings {
		if r.RaterID == raterID {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockRepository) UpdateAgentStats(ctx context.Context, agentID uuid.UUID, successful bool) error {
	stats := m.agentStats[agentID]
	stats.total++
	if successful {
		stats.successful++
	}
	m.agentStats[agentID] = stats
	return nil
}

func (m *mockRepository) RecalculateAgentRating(ctx context.Context, agentID uuid.UUID) error {
	return nil
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

func (m *mockPaymentService) CreateEscrowPayment(ctx context.Context, transactionID, buyerID, sellerID string, amount float64, currency string) (string, error) {
	piID := "pi_test_" + transactionID[:8]
	m.paymentIntents[piID] = true
	return piID, nil
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

func TestNewService(t *testing.T) {
	publisher := &mockPublisher{}
	service := NewService(nil, publisher)

	if service == nil {
		t.Fatal("expected service to be created")
	}
	if service.publisher != publisher {
		t.Error("publisher not set correctly")
	}
	if service.payment != nil {
		t.Error("payment should be nil initially")
	}
}

func TestSetPaymentService(t *testing.T) {
	service := NewService(nil, nil)
	paymentService := newMockPaymentService()

	service.SetPaymentService(paymentService)

	if service.payment == nil {
		t.Error("payment service should be set")
	}
}

func TestPublishEvent(t *testing.T) {
	publisher := &mockPublisher{}
	service := &Service{publisher: publisher}

	ctx := context.Background()
	service.publishEvent(ctx, "test.event", map[string]any{"key": "value"})

	// Give the goroutine time to execute
	time.Sleep(10 * time.Millisecond)

	if len(publisher.events) != 1 {
		t.Errorf("expected 1 event, got %d", len(publisher.events))
	}
	if publisher.events[0].eventType != "test.event" {
		t.Errorf("expected event type 'test.event', got %s", publisher.events[0].eventType)
	}
}

func TestPublishEventNilPublisher(t *testing.T) {
	service := &Service{publisher: nil}

	// Should not panic with nil publisher
	ctx := context.Background()
	service.publishEvent(ctx, "test.event", map[string]any{"key": "value"})
}

func TestTransactionAllFields(t *testing.T) {
	now := time.Now()
	requestID := uuid.New()
	offerID := uuid.New()
	listingID := uuid.New()
	auctionID := uuid.New()
	completedAt := now.Add(24 * time.Hour)
	deliveryConfirmedAt := now.Add(12 * time.Hour)

	tx := &Transaction{
		ID:                  uuid.New(),
		BuyerID:             uuid.New(),
		SellerID:            uuid.New(),
		ListingID:           &listingID,
		RequestID:           &requestID,
		OfferID:             &offerID,
		AuctionID:           &auctionID,
		Amount:              500.0,
		Currency:            "USD",
		PlatformFee:         25.0,
		Status:              StatusCompleted,
		DeliveryConfirmedAt: &deliveryConfirmedAt,
		CompletedAt:         &completedAt,
		Metadata:            map[string]any{"notes": "test transaction"},
		CreatedAt:           now,
		UpdatedAt:           now,
		BuyerName:           "Test Buyer",
		SellerName:          "Test Seller",
	}

	if tx.Amount != 500.0 {
		t.Errorf("expected amount 500.0, got %f", tx.Amount)
	}
	if tx.Status != StatusCompleted {
		t.Error("status not set correctly")
	}
	if tx.PlatformFee != 25.0 {
		t.Errorf("expected platform fee 25.0, got %f", tx.PlatformFee)
	}
	if tx.BuyerName != "Test Buyer" {
		t.Error("buyer name not set correctly")
	}
	if tx.DeliveryConfirmedAt == nil {
		t.Error("delivery confirmed at should be set")
	}
}

func TestEscrowAccountAllFields(t *testing.T) {
	now := time.Now()
	releasedAt := now.Add(48 * time.Hour)

	escrow := &EscrowAccount{
		ID:                    uuid.New(),
		TransactionID:         uuid.New(),
		Amount:                250.0,
		Currency:              "EUR",
		Status:                EscrowFunded,
		StripePaymentIntentID: strPtr("pi_test123"),
		FundedAt:              &now,
		ReleasedAt:            &releasedAt,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	if escrow.Amount != 250.0 {
		t.Errorf("expected amount 250.0, got %f", escrow.Amount)
	}
	if escrow.Status != EscrowFunded {
		t.Error("status not set correctly")
	}
	if escrow.StripePaymentIntentID == nil || *escrow.StripePaymentIntentID != "pi_test123" {
		t.Error("payment intent ID not set correctly")
	}
	if escrow.FundedAt == nil {
		t.Error("funded at should be set")
	}
}

func TestRatingAllFields(t *testing.T) {
	now := time.Now()
	rating := &Rating{
		ID:            uuid.New(),
		TransactionID: uuid.New(),
		RaterID:       uuid.New(),
		RatedAgentID:  uuid.New(),
		Score:         5,
		Comment:       "Excellent service, highly recommended!",
		CreatedAt:     now,
	}

	if rating.Score != 5 {
		t.Errorf("expected score 5, got %d", rating.Score)
	}
	if rating.Comment != "Excellent service, highly recommended!" {
		t.Error("comment not set correctly")
	}
}

func TestRatingScoreValidation(t *testing.T) {
	validScores := []int{1, 2, 3, 4, 5}
	for _, score := range validScores {
		if score < 1 || score > 5 {
			t.Errorf("score %d should be valid", score)
		}
	}

	invalidScores := []int{0, -1, 6, 10, 100}
	for _, score := range invalidScores {
		if score >= 1 && score <= 5 {
			t.Errorf("score %d should be invalid", score)
		}
	}
}

func TestListTransactionsParamsDefaults(t *testing.T) {
	params := ListTransactionsParams{}

	if params.Limit != 0 {
		t.Errorf("expected default limit 0, got %d", params.Limit)
	}
	if params.Offset != 0 {
		t.Errorf("expected default offset 0, got %d", params.Offset)
	}
}

func TestListTransactionsParamsWithFilters(t *testing.T) {
	agentID := uuid.New()
	status := StatusDelivered

	params := ListTransactionsParams{
		AgentID: &agentID,
		Status:  &status,
		Role:    "seller",
		Limit:   50,
		Offset:  100,
	}

	if *params.AgentID != agentID {
		t.Error("agent ID not set correctly")
	}
	if *params.Status != StatusDelivered {
		t.Error("status not set correctly")
	}
	if params.Role != "seller" {
		t.Errorf("expected role 'seller', got %s", params.Role)
	}
	if params.Limit != 50 {
		t.Errorf("expected limit 50, got %d", params.Limit)
	}
	if params.Offset != 100 {
		t.Errorf("expected offset 100, got %d", params.Offset)
	}
}

func TestTransactionListResultEmpty(t *testing.T) {
	result := &TransactionListResult{
		Items:  []*Transaction{},
		Total:  0,
		Limit:  20,
		Offset: 0,
	}

	if len(result.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(result.Items))
	}
	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}
}

func TestTransactionListResultWithItems(t *testing.T) {
	items := []*Transaction{
		{ID: uuid.New(), Amount: 100.0},
		{ID: uuid.New(), Amount: 200.0},
		{ID: uuid.New(), Amount: 300.0},
	}

	result := &TransactionListResult{
		Items:  items,
		Total:  100,
		Limit:  20,
		Offset: 40,
	}

	if len(result.Items) != 3 {
		t.Errorf("expected 3 items, got %d", len(result.Items))
	}
	if result.Total != 100 {
		t.Errorf("expected total 100, got %d", result.Total)
	}
	if result.Offset != 40 {
		t.Errorf("expected offset 40, got %d", result.Offset)
	}
}

func TestCreateTransactionRequestValidation(t *testing.T) {
	requestID := uuid.New()
	offerID := uuid.New()

	req := &CreateTransactionRequest{
		BuyerID:   uuid.New(),
		SellerID:  uuid.New(),
		RequestID: &requestID,
		OfferID:   &offerID,
		Amount:    100.0,
		Currency:  "USD",
	}

	// Validate required fields
	if req.BuyerID == uuid.Nil {
		t.Error("buyer ID should not be nil")
	}
	if req.SellerID == uuid.Nil {
		t.Error("seller ID should not be nil")
	}
	if req.Amount <= 0 {
		t.Error("amount should be positive")
	}
}

func TestSubmitRatingRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     SubmitRatingRequest
		wantErr bool
	}{
		{
			name:    "valid rating",
			req:     SubmitRatingRequest{Score: 5, Comment: "Great!"},
			wantErr: false,
		},
		{
			name:    "score too low",
			req:     SubmitRatingRequest{Score: 0},
			wantErr: true,
		},
		{
			name:    "score too high",
			req:     SubmitRatingRequest{Score: 6},
			wantErr: true,
		},
		{
			name:    "min valid score",
			req:     SubmitRatingRequest{Score: 1},
			wantErr: false,
		},
		{
			name:    "max valid score",
			req:     SubmitRatingRequest{Score: 5},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasErr := tt.req.Score < 1 || tt.req.Score > 5
			if hasErr != tt.wantErr {
				t.Errorf("validation error = %v, wantErr %v", hasErr, tt.wantErr)
			}
		})
	}
}

func TestDisputeRequestValidation(t *testing.T) {
	req := &DisputeRequest{
		Reason:      "Item not received",
		Description: "I never received the data package that was promised.",
	}

	if req.Reason == "" {
		t.Error("reason should not be empty")
	}
	if len(req.Description) == 0 {
		t.Error("description should be provided")
	}
}

func TestDisputeRequestEmptyReason(t *testing.T) {
	req := &DisputeRequest{
		Reason:      "",
		Description: "Some description",
	}

	if req.Reason != "" {
		t.Error("expected empty reason")
	}
}

func TestEscrowFundingResultAllFields(t *testing.T) {
	result := &EscrowFundingResult{
		TransactionID:   uuid.New(),
		PaymentIntentID: "pi_test_abc123",
		Amount:          150.0,
		Currency:        "USD",
	}

	if result.PaymentIntentID != "pi_test_abc123" {
		t.Error("payment intent ID not set correctly")
	}
	if result.Amount != 150.0 {
		t.Errorf("expected amount 150.0, got %f", result.Amount)
	}
}

// --- Service Method Tests ---

func TestService_CreateTransaction(t *testing.T) {
	repo := newMockRepository()
	publisher := &mockPublisher{}
	service := NewService(repo, publisher)

	buyerID := uuid.New()
	sellerID := uuid.New()

	tx, err := service.CreateTransaction(context.Background(), &CreateTransactionRequest{
		BuyerID:  buyerID,
		SellerID: sellerID,
		Amount:   100.0,
		Currency: "USD",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx.BuyerID != buyerID {
		t.Errorf("expected buyer ID %s, got %s", buyerID, tx.BuyerID)
	}
	if tx.Status != StatusPending {
		t.Errorf("expected status pending, got %s", tx.Status)
	}
	if tx.Amount != 100.0 {
		t.Errorf("expected amount 100.0, got %f", tx.Amount)
	}

	// Verify escrow was created
	if len(repo.escrows) != 1 {
		t.Errorf("expected 1 escrow, got %d", len(repo.escrows))
	}

	// Wait for async event publishing
	time.Sleep(10 * time.Millisecond)
	if len(publisher.events) < 1 {
		t.Error("expected transaction.created event to be published")
	}
}

func TestService_CreateTransaction_Error(t *testing.T) {
	repo := newMockRepository()
	repo.createErr = ErrTransactionNotFound
	service := NewService(repo, nil)

	_, err := service.CreateTransaction(context.Background(), &CreateTransactionRequest{
		BuyerID:  uuid.New(),
		SellerID: uuid.New(),
		Amount:   100.0,
	})

	if err == nil {
		t.Error("expected error")
	}
}

func TestService_CreateFromOffer(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	buyerID := uuid.New()
	sellerID := uuid.New()
	requestID := uuid.New()
	offerID := uuid.New()

	txID, err := service.CreateFromOffer(context.Background(), buyerID, sellerID, &requestID, &offerID, 150.0, "EUR")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if txID == uuid.Nil {
		t.Error("expected non-nil transaction ID")
	}

	// Verify transaction exists
	tx, err := service.GetTransaction(context.Background(), txID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx.Amount != 150.0 {
		t.Errorf("expected amount 150.0, got %f", tx.Amount)
	}
	if tx.Currency != "EUR" {
		t.Errorf("expected currency EUR, got %s", tx.Currency)
	}
}

func TestService_GetTransaction(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	// Create a transaction first
	tx, _ := repo.CreateTransaction(context.Background(), &CreateTransactionRequest{
		BuyerID:  uuid.New(),
		SellerID: uuid.New(),
		Amount:   100.0,
	})

	// Get it back
	retrieved, err := service.GetTransaction(context.Background(), tx.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retrieved.ID != tx.ID {
		t.Errorf("expected ID %s, got %s", tx.ID, retrieved.ID)
	}
}

func TestService_GetTransaction_NotFound(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	_, err := service.GetTransaction(context.Background(), uuid.New())
	if err != ErrTransactionNotFound {
		t.Errorf("expected ErrTransactionNotFound, got %v", err)
	}
}

func TestService_ListTransactions(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	buyerID := uuid.New()

	// Create some transactions
	for i := 0; i < 5; i++ {
		repo.CreateTransaction(context.Background(), &CreateTransactionRequest{
			BuyerID:  buyerID,
			SellerID: uuid.New(),
			Amount:   float64(100 * (i + 1)),
		})
	}

	// List all for buyer
	result, err := service.ListTransactions(context.Background(), ListTransactionsParams{
		AgentID: &buyerID,
		Role:    "buyer",
		Limit:   10,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 5 {
		t.Errorf("expected 5 transactions, got %d", result.Total)
	}
}

func TestService_ListTransactions_DefaultLimit(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	// Test default limit application
	result, err := service.ListTransactions(context.Background(), ListTransactionsParams{
		Limit: 0, // Should default to 20
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Limit != 20 {
		t.Errorf("expected default limit 20, got %d", result.Limit)
	}
}

func TestService_ListTransactions_MaxLimit(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	// Test max limit capping
	result, err := service.ListTransactions(context.Background(), ListTransactionsParams{
		Limit: 500, // Should be capped to 100
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Limit != 100 {
		t.Errorf("expected max limit 100, got %d", result.Limit)
	}
}

func TestService_FundEscrow(t *testing.T) {
	repo := newMockRepository()
	payment := newMockPaymentService()
	service := NewService(repo, nil)
	service.SetPaymentService(payment)

	buyerID := uuid.New()
	tx, _ := repo.CreateTransaction(context.Background(), &CreateTransactionRequest{
		BuyerID:  buyerID,
		SellerID: uuid.New(),
		Amount:   100.0,
		Currency: "USD",
	})
	repo.CreateEscrowAccount(context.Background(), tx.ID, 100.0, "USD")

	result, err := service.FundEscrow(context.Background(), tx.ID, buyerID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PaymentIntentID == "" {
		t.Error("expected payment intent ID")
	}
	if result.Amount != 100.0 {
		t.Errorf("expected amount 100.0, got %f", result.Amount)
	}
}

func TestService_FundEscrow_NotBuyer(t *testing.T) {
	repo := newMockRepository()
	payment := newMockPaymentService()
	service := NewService(repo, nil)
	service.SetPaymentService(payment)

	tx, _ := repo.CreateTransaction(context.Background(), &CreateTransactionRequest{
		BuyerID:  uuid.New(),
		SellerID: uuid.New(),
		Amount:   100.0,
	})

	_, err := service.FundEscrow(context.Background(), tx.ID, uuid.New()) // Different agent

	if err != ErrNotAuthorized {
		t.Errorf("expected ErrNotAuthorized, got %v", err)
	}
}

func TestService_FundEscrow_WrongStatus(t *testing.T) {
	repo := newMockRepository()
	payment := newMockPaymentService()
	service := NewService(repo, nil)
	service.SetPaymentService(payment)

	buyerID := uuid.New()
	tx, _ := repo.CreateTransaction(context.Background(), &CreateTransactionRequest{
		BuyerID:  buyerID,
		SellerID: uuid.New(),
		Amount:   100.0,
	})
	tx.Status = StatusCompleted // Wrong status

	_, err := service.FundEscrow(context.Background(), tx.ID, buyerID)

	if err != ErrInvalidStatus {
		t.Errorf("expected ErrInvalidStatus, got %v", err)
	}
}

func TestService_FundEscrow_NoPaymentService(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil) // No payment service

	buyerID := uuid.New()
	tx, _ := repo.CreateTransaction(context.Background(), &CreateTransactionRequest{
		BuyerID:  buyerID,
		SellerID: uuid.New(),
		Amount:   100.0,
	})

	_, err := service.FundEscrow(context.Background(), tx.ID, buyerID)

	if err == nil {
		t.Error("expected error for missing payment service")
	}
}

func TestService_ConfirmEscrowFunded(t *testing.T) {
	repo := newMockRepository()
	publisher := &mockPublisher{}
	service := NewService(repo, publisher)

	tx, _ := repo.CreateTransaction(context.Background(), &CreateTransactionRequest{
		BuyerID:  uuid.New(),
		SellerID: uuid.New(),
		Amount:   100.0,
	})
	repo.CreateEscrowAccount(context.Background(), tx.ID, 100.0, "USD")

	err := service.ConfirmEscrowFunded(context.Background(), tx.ID, "pi_test123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check transaction status
	updated, _ := repo.GetTransactionByID(context.Background(), tx.ID)
	if updated.Status != StatusEscrowFunded {
		t.Errorf("expected status escrow_funded, got %s", updated.Status)
	}

	// Wait for async event
	time.Sleep(10 * time.Millisecond)
	if len(publisher.events) < 1 {
		t.Error("expected event to be published")
	}
}

func TestService_MarkDelivered(t *testing.T) {
	repo := newMockRepository()
	publisher := &mockPublisher{}
	service := NewService(repo, publisher)

	sellerID := uuid.New()
	tx, _ := repo.CreateTransaction(context.Background(), &CreateTransactionRequest{
		BuyerID:  uuid.New(),
		SellerID: sellerID,
		Amount:   100.0,
	})

	updated, err := service.MarkDelivered(context.Background(), tx.ID, sellerID, "proof123", "Delivered successfully")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Status != StatusDelivered {
		t.Errorf("expected status delivered, got %s", updated.Status)
	}
}

func TestService_MarkDelivered_NotSeller(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	tx, _ := repo.CreateTransaction(context.Background(), &CreateTransactionRequest{
		BuyerID:  uuid.New(),
		SellerID: uuid.New(),
		Amount:   100.0,
	})

	_, err := service.MarkDelivered(context.Background(), tx.ID, uuid.New(), "", "") // Different agent

	if err != ErrNotAuthorized {
		t.Errorf("expected ErrNotAuthorized, got %v", err)
	}
}

func TestService_MarkDelivered_InvalidStatus(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	sellerID := uuid.New()
	tx, _ := repo.CreateTransaction(context.Background(), &CreateTransactionRequest{
		BuyerID:  uuid.New(),
		SellerID: sellerID,
		Amount:   100.0,
	})
	tx.Status = StatusCompleted // Invalid for marking delivered

	_, err := service.MarkDelivered(context.Background(), tx.ID, sellerID, "", "")

	if err != ErrInvalidStatus {
		t.Errorf("expected ErrInvalidStatus, got %v", err)
	}
}

func TestService_ConfirmDelivery(t *testing.T) {
	repo := newMockRepository()
	publisher := &mockPublisher{}
	service := NewService(repo, publisher)

	buyerID := uuid.New()
	tx, _ := repo.CreateTransaction(context.Background(), &CreateTransactionRequest{
		BuyerID:  buyerID,
		SellerID: uuid.New(),
		Amount:   100.0,
	})

	// Transaction must be in 'delivered' status for buyer to confirm
	repo.UpdateTransactionStatus(context.Background(), tx.ID, StatusDelivered)

	updated, err := service.ConfirmDelivery(context.Background(), tx.ID, buyerID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Status != StatusCompleted {
		t.Errorf("expected status completed, got %s", updated.Status)
	}
}

func TestService_ConfirmDelivery_NotBuyer(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	tx, _ := repo.CreateTransaction(context.Background(), &CreateTransactionRequest{
		BuyerID:  uuid.New(),
		SellerID: uuid.New(),
		Amount:   100.0,
	})

	// Transaction must be in 'delivered' status
	repo.UpdateTransactionStatus(context.Background(), tx.ID, StatusDelivered)

	_, err := service.ConfirmDelivery(context.Background(), tx.ID, uuid.New()) // Different agent

	if err != ErrNotAuthorized {
		t.Errorf("expected ErrNotAuthorized, got %v", err)
	}
}

func TestService_CompleteTransaction(t *testing.T) {
	repo := newMockRepository()
	publisher := &mockPublisher{}
	service := NewService(repo, publisher)

	buyerID := uuid.New()
	sellerID := uuid.New()
	tx, _ := repo.CreateTransaction(context.Background(), &CreateTransactionRequest{
		BuyerID:  buyerID,
		SellerID: sellerID,
		Amount:   100.0,
	})
	// Must be delivered first (seller has marked as delivered)
	repo.UpdateTransactionStatus(context.Background(), tx.ID, StatusDelivered)

	updated, err := service.CompleteTransaction(context.Background(), tx.ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Status != StatusCompleted {
		t.Errorf("expected status completed, got %s", updated.Status)
	}

	// Check agent stats were updated
	buyerStats := repo.agentStats[buyerID]
	sellerStats := repo.agentStats[sellerID]
	if buyerStats.total != 1 || buyerStats.successful != 1 {
		t.Error("buyer stats not updated correctly")
	}
	if sellerStats.total != 1 || sellerStats.successful != 1 {
		t.Error("seller stats not updated correctly")
	}
}

func TestService_CompleteTransaction_NotDelivered(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	tx, _ := repo.CreateTransaction(context.Background(), &CreateTransactionRequest{
		BuyerID:  uuid.New(),
		SellerID: uuid.New(),
		Amount:   100.0,
	})
	// Status is pending, not delivered

	_, err := service.CompleteTransaction(context.Background(), tx.ID)

	if err != ErrInvalidStatus {
		t.Errorf("expected ErrInvalidStatus, got %v", err)
	}
}

func TestService_SubmitRating(t *testing.T) {
	repo := newMockRepository()
	publisher := &mockPublisher{}
	service := NewService(repo, publisher)

	buyerID := uuid.New()
	sellerID := uuid.New()
	tx, _ := repo.CreateTransaction(context.Background(), &CreateTransactionRequest{
		BuyerID:  buyerID,
		SellerID: sellerID,
		Amount:   100.0,
	})
	repo.ConfirmDelivery(context.Background(), tx.ID)

	rating, err := service.SubmitRating(context.Background(), tx.ID, buyerID, &SubmitRatingRequest{
		Score:   5,
		Comment: "Excellent service!",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rating.Score != 5 {
		t.Errorf("expected score 5, got %d", rating.Score)
	}
	if rating.RaterID != buyerID {
		t.Errorf("expected rater ID %s, got %s", buyerID, rating.RaterID)
	}
	if rating.RatedAgentID != sellerID {
		t.Errorf("expected rated agent ID %s, got %s", sellerID, rating.RatedAgentID)
	}
}

func TestService_SubmitRating_InvalidScore(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	buyerID := uuid.New()
	tx, _ := repo.CreateTransaction(context.Background(), &CreateTransactionRequest{
		BuyerID:  buyerID,
		SellerID: uuid.New(),
		Amount:   100.0,
	})
	repo.ConfirmDelivery(context.Background(), tx.ID)

	tests := []struct {
		score int
	}{
		{0},
		{-1},
		{6},
		{100},
	}

	for _, tt := range tests {
		_, err := service.SubmitRating(context.Background(), tx.ID, buyerID, &SubmitRatingRequest{
			Score: tt.score,
		})
		if err != ErrInvalidRating {
			t.Errorf("expected ErrInvalidRating for score %d, got %v", tt.score, err)
		}
	}
}

func TestService_SubmitRating_NotParticipant(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	tx, _ := repo.CreateTransaction(context.Background(), &CreateTransactionRequest{
		BuyerID:  uuid.New(),
		SellerID: uuid.New(),
		Amount:   100.0,
	})
	repo.ConfirmDelivery(context.Background(), tx.ID)

	_, err := service.SubmitRating(context.Background(), tx.ID, uuid.New(), &SubmitRatingRequest{
		Score: 5,
	})

	if err != ErrNotAuthorized {
		t.Errorf("expected ErrNotAuthorized, got %v", err)
	}
}

func TestService_SubmitRating_TransactionNotReady(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	buyerID := uuid.New()
	tx, _ := repo.CreateTransaction(context.Background(), &CreateTransactionRequest{
		BuyerID:  buyerID,
		SellerID: uuid.New(),
		Amount:   100.0,
	})
	// Status is pending, not delivered or completed

	_, err := service.SubmitRating(context.Background(), tx.ID, buyerID, &SubmitRatingRequest{
		Score: 5,
	})

	if err != ErrTransactionNotReady {
		t.Errorf("expected ErrTransactionNotReady, got %v", err)
	}
}

func TestService_SubmitRating_AlreadyRated(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	buyerID := uuid.New()
	tx, _ := repo.CreateTransaction(context.Background(), &CreateTransactionRequest{
		BuyerID:  buyerID,
		SellerID: uuid.New(),
		Amount:   100.0,
	})
	repo.ConfirmDelivery(context.Background(), tx.ID)

	// First rating
	_, err := service.SubmitRating(context.Background(), tx.ID, buyerID, &SubmitRatingRequest{
		Score: 5,
	})
	if err != nil {
		t.Fatalf("unexpected error on first rating: %v", err)
	}

	// Second rating - should fail
	_, err = service.SubmitRating(context.Background(), tx.ID, buyerID, &SubmitRatingRequest{
		Score: 4,
	})

	if err != ErrRatingAlreadyExists {
		t.Errorf("expected ErrRatingAlreadyExists, got %v", err)
	}
}

func TestService_SubmitRating_AutoComplete(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	buyerID := uuid.New()
	sellerID := uuid.New()
	tx, _ := repo.CreateTransaction(context.Background(), &CreateTransactionRequest{
		BuyerID:  buyerID,
		SellerID: sellerID,
		Amount:   100.0,
	})
	repo.ConfirmDelivery(context.Background(), tx.ID)

	// Both parties rate
	service.SubmitRating(context.Background(), tx.ID, buyerID, &SubmitRatingRequest{Score: 5})
	service.SubmitRating(context.Background(), tx.ID, sellerID, &SubmitRatingRequest{Score: 4})

	// Transaction should be auto-completed
	updated, _ := repo.GetTransactionByID(context.Background(), tx.ID)
	if updated.Status != StatusCompleted {
		t.Errorf("expected transaction to be auto-completed, got status %s", updated.Status)
	}
}

func TestService_GetTransactionRatings(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	buyerID := uuid.New()
	sellerID := uuid.New()
	tx, _ := repo.CreateTransaction(context.Background(), &CreateTransactionRequest{
		BuyerID:  buyerID,
		SellerID: sellerID,
		Amount:   100.0,
	})
	repo.ConfirmDelivery(context.Background(), tx.ID)

	// Submit ratings
	service.SubmitRating(context.Background(), tx.ID, buyerID, &SubmitRatingRequest{Score: 5})
	service.SubmitRating(context.Background(), tx.ID, sellerID, &SubmitRatingRequest{Score: 4})

	// Get ratings
	ratings, err := service.GetTransactionRatings(context.Background(), tx.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ratings) != 2 {
		t.Errorf("expected 2 ratings, got %d", len(ratings))
	}
}

func TestService_DisputeTransaction(t *testing.T) {
	repo := newMockRepository()
	publisher := &mockPublisher{}
	service := NewService(repo, publisher)

	buyerID := uuid.New()
	tx, _ := repo.CreateTransaction(context.Background(), &CreateTransactionRequest{
		BuyerID:  buyerID,
		SellerID: uuid.New(),
		Amount:   100.0,
	})
	repo.CreateEscrowAccount(context.Background(), tx.ID, 100.0, "USD")

	updated, err := service.DisputeTransaction(context.Background(), tx.ID, buyerID, &DisputeRequest{
		Reason:      "Item not as described",
		Description: "The data was incomplete",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Status != StatusDisputed {
		t.Errorf("expected status disputed, got %s", updated.Status)
	}
}

func TestService_DisputeTransaction_NotParticipant(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	tx, _ := repo.CreateTransaction(context.Background(), &CreateTransactionRequest{
		BuyerID:  uuid.New(),
		SellerID: uuid.New(),
		Amount:   100.0,
	})

	_, err := service.DisputeTransaction(context.Background(), tx.ID, uuid.New(), &DisputeRequest{
		Reason: "Test",
	})

	if err != ErrNotAuthorized {
		t.Errorf("expected ErrNotAuthorized, got %v", err)
	}
}

func TestService_DisputeTransaction_InvalidStatus(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	buyerID := uuid.New()
	tx, _ := repo.CreateTransaction(context.Background(), &CreateTransactionRequest{
		BuyerID:  buyerID,
		SellerID: uuid.New(),
		Amount:   100.0,
	})
	tx.Status = StatusCompleted // Can't dispute completed

	_, err := service.DisputeTransaction(context.Background(), tx.ID, buyerID, &DisputeRequest{
		Reason: "Test",
	})

	if err != ErrInvalidStatus {
		t.Errorf("expected ErrInvalidStatus, got %v", err)
	}
}

func TestService_RefundTransaction(t *testing.T) {
	repo := newMockRepository()
	publisher := &mockPublisher{}
	service := NewService(repo, publisher)

	tx, _ := repo.CreateTransaction(context.Background(), &CreateTransactionRequest{
		BuyerID:  uuid.New(),
		SellerID: uuid.New(),
		Amount:   100.0,
	})
	repo.CreateEscrowAccount(context.Background(), tx.ID, 100.0, "USD")

	err := service.RefundTransaction(context.Background(), tx.ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check status
	updated, _ := repo.GetTransactionByID(context.Background(), tx.ID)
	if updated.Status != StatusRefunded {
		t.Errorf("expected status refunded, got %s", updated.Status)
	}

	// Wait for async event
	time.Sleep(10 * time.Millisecond)
	hasRefundEvent := false
	for _, e := range publisher.events {
		if e.eventType == "transaction.refunded" {
			hasRefundEvent = true
			break
		}
	}
	if !hasRefundEvent {
		t.Error("expected transaction.refunded event")
	}
}

func TestRepositoryErrors(t *testing.T) {
	if ErrTransactionNotFound.Error() != "transaction not found" {
		t.Errorf("unexpected error message: %s", ErrTransactionNotFound.Error())
	}
	if ErrRatingNotFound.Error() != "rating not found" {
		t.Errorf("unexpected error message: %s", ErrRatingNotFound.Error())
	}
	if ErrRatingAlreadyExists.Error() != "rating already submitted for this transaction" {
		t.Errorf("unexpected error message: %s", ErrRatingAlreadyExists.Error())
	}
	if ErrEscrowNotFound.Error() != "escrow account not found" {
		t.Errorf("unexpected error message: %s", ErrEscrowNotFound.Error())
	}
}
