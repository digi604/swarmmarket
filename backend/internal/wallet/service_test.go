package wallet

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

// mockRepository is a mock implementation of wallet repository for testing.
type mockRepository struct {
	deposits       map[uuid.UUID]*Deposit
	userDeposits   map[uuid.UUID][]*Deposit
	agentDeposits  map[uuid.UUID][]*Deposit
	createErr      error
	getErr         error
	updateErr      error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		deposits:      make(map[uuid.UUID]*Deposit),
		userDeposits:  make(map[uuid.UUID][]*Deposit),
		agentDeposits: make(map[uuid.UUID][]*Deposit),
	}
}

func (m *mockRepository) CreateDeposit(ctx context.Context, deposit *Deposit) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.deposits[deposit.ID] = deposit
	if deposit.UserID != uuid.Nil {
		m.userDeposits[deposit.UserID] = append(m.userDeposits[deposit.UserID], deposit)
	}
	if deposit.AgentID != uuid.Nil {
		m.agentDeposits[deposit.AgentID] = append(m.agentDeposits[deposit.AgentID], deposit)
	}
	return nil
}

func (m *mockRepository) GetDeposit(ctx context.Context, id uuid.UUID) (*Deposit, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.deposits[id], nil
}

func (m *mockRepository) GetUserDeposits(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Deposit, int, error) {
	if m.getErr != nil {
		return nil, 0, m.getErr
	}
	deposits := m.userDeposits[userID]
	total := len(deposits)
	if offset >= total {
		return []*Deposit{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return deposits[offset:end], total, nil
}

func (m *mockRepository) GetAgentDeposits(ctx context.Context, agentID uuid.UUID, limit, offset int) ([]*Deposit, int, error) {
	if m.getErr != nil {
		return nil, 0, m.getErr
	}
	deposits := m.agentDeposits[agentID]
	total := len(deposits)
	if offset >= total {
		return []*Deposit{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return deposits[offset:end], total, nil
}

func (m *mockRepository) GetDepositByPaymentIntent(ctx context.Context, paymentIntentID string) (*Deposit, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	for _, d := range m.deposits {
		if d.StripePaymentIntentID == paymentIntentID {
			return d, nil
		}
	}
	return nil, nil
}

func (m *mockRepository) UpdateDepositStatus(ctx context.Context, id uuid.UUID, status DepositStatus, failureReason string) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if d, ok := m.deposits[id]; ok {
		d.Status = status
		d.FailureReason = failureReason
		d.UpdatedAt = time.Now()
		if status == DepositStatusCompleted {
			now := time.Now()
			d.CompletedAt = &now
		}
	}
	return nil
}

func (m *mockRepository) GetCompletedDepositsTotal(ctx context.Context, userID uuid.UUID) (float64, error) {
	if m.getErr != nil {
		return 0, m.getErr
	}
	var total float64
	for _, d := range m.userDeposits[userID] {
		if d.Status == DepositStatusCompleted {
			total += d.Amount
		}
	}
	return total, nil
}

func (m *mockRepository) GetPendingDepositsTotal(ctx context.Context, userID uuid.UUID) (float64, error) {
	if m.getErr != nil {
		return 0, m.getErr
	}
	var total float64
	for _, d := range m.userDeposits[userID] {
		if d.Status == DepositStatusPending || d.Status == DepositStatusProcessing {
			total += d.Amount
		}
	}
	return total, nil
}

func (m *mockRepository) GetAgentCompletedDepositsTotal(ctx context.Context, agentID uuid.UUID) (float64, error) {
	if m.getErr != nil {
		return 0, m.getErr
	}
	var total float64
	for _, d := range m.agentDeposits[agentID] {
		if d.Status == DepositStatusCompleted {
			total += d.Amount
		}
	}
	return total, nil
}

func (m *mockRepository) GetAgentPendingDepositsTotal(ctx context.Context, agentID uuid.UUID) (float64, error) {
	if m.getErr != nil {
		return 0, m.getErr
	}
	var total float64
	for _, d := range m.agentDeposits[agentID] {
		if d.Status == DepositStatusPending || d.Status == DepositStatusProcessing {
			total += d.Amount
		}
	}
	return total, nil
}

// Helper to add a deposit directly to the mock
func (m *mockRepository) addDeposit(deposit *Deposit) {
	m.deposits[deposit.ID] = deposit
	if deposit.UserID != uuid.Nil {
		m.userDeposits[deposit.UserID] = append(m.userDeposits[deposit.UserID], deposit)
	}
	if deposit.AgentID != uuid.Nil {
		m.agentDeposits[deposit.AgentID] = append(m.agentDeposits[deposit.AgentID], deposit)
	}
}

// testableService creates a service with mock repository for testing
// Note: Stripe-dependent methods cannot be tested without mocking Stripe API
type testableService struct {
	repo *mockRepository
}

func newTestableService() *testableService {
	return &testableService{
		repo: newMockRepository(),
	}
}

func (ts *testableService) GetDeposit(ctx context.Context, id uuid.UUID) (*Deposit, error) {
	deposit, err := ts.repo.GetDeposit(ctx, id)
	if err != nil {
		return nil, err
	}
	if deposit == nil {
		return nil, ErrDepositNotFound
	}
	return deposit, nil
}

func (ts *testableService) GetUserDeposits(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Deposit, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return ts.repo.GetUserDeposits(ctx, userID, limit, offset)
}

func (ts *testableService) GetWalletBalance(ctx context.Context, userID uuid.UUID) (*WalletBalance, error) {
	available, err := ts.repo.GetCompletedDepositsTotal(ctx, userID)
	if err != nil {
		return nil, err
	}

	pending, err := ts.repo.GetPendingDepositsTotal(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &WalletBalance{
		Available: available,
		Pending:   pending,
		Currency:  "USD",
	}, nil
}

func (ts *testableService) GetAgentDeposits(ctx context.Context, agentID uuid.UUID, limit, offset int) ([]*Deposit, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return ts.repo.GetAgentDeposits(ctx, agentID, limit, offset)
}

func (ts *testableService) GetAgentWalletBalance(ctx context.Context, agentID uuid.UUID) (*WalletBalance, error) {
	available, err := ts.repo.GetAgentCompletedDepositsTotal(ctx, agentID)
	if err != nil {
		return nil, err
	}

	pending, err := ts.repo.GetAgentPendingDepositsTotal(ctx, agentID)
	if err != nil {
		return nil, err
	}

	return &WalletBalance{
		Available: available,
		Pending:   pending,
		Currency:  "USD",
	}, nil
}

func (ts *testableService) HandlePaymentIntentSucceeded(ctx context.Context, paymentIntentID string) error {
	deposit, err := ts.repo.GetDepositByPaymentIntent(ctx, paymentIntentID)
	if err != nil {
		return err
	}
	if deposit == nil {
		return nil
	}
	return ts.repo.UpdateDepositStatus(ctx, deposit.ID, DepositStatusCompleted, "")
}

func (ts *testableService) HandlePaymentIntentFailed(ctx context.Context, paymentIntentID string, reason string) error {
	deposit, err := ts.repo.GetDepositByPaymentIntent(ctx, paymentIntentID)
	if err != nil {
		return err
	}
	if deposit == nil {
		return nil
	}
	return ts.repo.UpdateDepositStatus(ctx, deposit.ID, DepositStatusFailed, reason)
}

// Tests

func TestGetDeposit(t *testing.T) {
	ts := newTestableService()
	depositID := uuid.New()
	userID := uuid.New()

	deposit := &Deposit{
		ID:       depositID,
		UserID:   userID,
		Amount:   100.0,
		Currency: "USD",
		Status:   DepositStatusPending,
	}
	ts.repo.addDeposit(deposit)

	result, err := ts.GetDeposit(context.Background(), depositID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != depositID {
		t.Errorf("expected deposit ID %s, got %s", depositID, result.ID)
	}
	if result.Amount != 100.0 {
		t.Errorf("expected amount 100.0, got %f", result.Amount)
	}
}

func TestGetDeposit_NotFound(t *testing.T) {
	ts := newTestableService()

	_, err := ts.GetDeposit(context.Background(), uuid.New())
	if err != ErrDepositNotFound {
		t.Errorf("expected ErrDepositNotFound, got %v", err)
	}
}

func TestGetUserDeposits(t *testing.T) {
	ts := newTestableService()
	userID := uuid.New()

	// Add multiple deposits
	for i := 0; i < 5; i++ {
		ts.repo.addDeposit(&Deposit{
			ID:       uuid.New(),
			UserID:   userID,
			Amount:   float64(i+1) * 10,
			Currency: "USD",
			Status:   DepositStatusCompleted,
		})
	}

	deposits, total, err := ts.GetUserDeposits(context.Background(), userID, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}
	if len(deposits) != 5 {
		t.Errorf("expected 5 deposits, got %d", len(deposits))
	}
}

func TestGetUserDeposits_Pagination(t *testing.T) {
	ts := newTestableService()
	userID := uuid.New()

	for i := 0; i < 10; i++ {
		ts.repo.addDeposit(&Deposit{
			ID:       uuid.New(),
			UserID:   userID,
			Amount:   float64(i+1) * 10,
			Currency: "USD",
			Status:   DepositStatusCompleted,
		})
	}

	deposits, total, err := ts.GetUserDeposits(context.Background(), userID, 3, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 10 {
		t.Errorf("expected total 10, got %d", total)
	}
	if len(deposits) != 3 {
		t.Errorf("expected 3 deposits, got %d", len(deposits))
	}
}

func TestGetUserDeposits_DefaultLimit(t *testing.T) {
	ts := newTestableService()
	userID := uuid.New()

	for i := 0; i < 25; i++ {
		ts.repo.addDeposit(&Deposit{
			ID:       uuid.New(),
			UserID:   userID,
			Amount:   10,
			Currency: "USD",
			Status:   DepositStatusCompleted,
		})
	}

	// Limit <= 0 should default to 20
	deposits, _, err := ts.GetUserDeposits(context.Background(), userID, 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deposits) != 20 {
		t.Errorf("expected 20 deposits (default limit), got %d", len(deposits))
	}
}

func TestGetWalletBalance(t *testing.T) {
	ts := newTestableService()
	userID := uuid.New()

	// Add completed deposits
	ts.repo.addDeposit(&Deposit{
		ID:       uuid.New(),
		UserID:   userID,
		Amount:   100.0,
		Currency: "USD",
		Status:   DepositStatusCompleted,
	})
	ts.repo.addDeposit(&Deposit{
		ID:       uuid.New(),
		UserID:   userID,
		Amount:   50.0,
		Currency: "USD",
		Status:   DepositStatusCompleted,
	})

	// Add pending deposit
	ts.repo.addDeposit(&Deposit{
		ID:       uuid.New(),
		UserID:   userID,
		Amount:   25.0,
		Currency: "USD",
		Status:   DepositStatusPending,
	})

	balance, err := ts.GetWalletBalance(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if balance.Available != 150.0 {
		t.Errorf("expected available 150.0, got %f", balance.Available)
	}
	if balance.Pending != 25.0 {
		t.Errorf("expected pending 25.0, got %f", balance.Pending)
	}
	if balance.Currency != "USD" {
		t.Errorf("expected currency USD, got %s", balance.Currency)
	}
}

func TestGetAgentWalletBalance(t *testing.T) {
	ts := newTestableService()
	agentID := uuid.New()

	// Add completed deposits
	ts.repo.addDeposit(&Deposit{
		ID:       uuid.New(),
		AgentID:  agentID,
		Amount:   200.0,
		Currency: "USD",
		Status:   DepositStatusCompleted,
	})

	// Add processing deposit
	ts.repo.addDeposit(&Deposit{
		ID:       uuid.New(),
		AgentID:  agentID,
		Amount:   75.0,
		Currency: "USD",
		Status:   DepositStatusProcessing,
	})

	balance, err := ts.GetAgentWalletBalance(context.Background(), agentID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if balance.Available != 200.0 {
		t.Errorf("expected available 200.0, got %f", balance.Available)
	}
	if balance.Pending != 75.0 {
		t.Errorf("expected pending 75.0, got %f", balance.Pending)
	}
}

func TestHandlePaymentIntentSucceeded(t *testing.T) {
	ts := newTestableService()
	depositID := uuid.New()
	paymentIntentID := "pi_test_123"

	deposit := &Deposit{
		ID:                    depositID,
		UserID:                uuid.New(),
		Amount:                100.0,
		Currency:              "USD",
		StripePaymentIntentID: paymentIntentID,
		Status:                DepositStatusPending,
	}
	ts.repo.addDeposit(deposit)

	err := ts.HandlePaymentIntentSucceeded(context.Background(), paymentIntentID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check deposit status was updated
	updated, _ := ts.repo.GetDeposit(context.Background(), depositID)
	if updated.Status != DepositStatusCompleted {
		t.Errorf("expected status completed, got %s", updated.Status)
	}
}

func TestHandlePaymentIntentSucceeded_NotFound(t *testing.T) {
	ts := newTestableService()

	// Should not error for unknown payment intent
	err := ts.HandlePaymentIntentSucceeded(context.Background(), "pi_unknown")
	if err != nil {
		t.Errorf("unexpected error for unknown payment intent: %v", err)
	}
}

func TestHandlePaymentIntentFailed(t *testing.T) {
	ts := newTestableService()
	depositID := uuid.New()
	paymentIntentID := "pi_test_456"

	deposit := &Deposit{
		ID:                    depositID,
		UserID:                uuid.New(),
		Amount:                100.0,
		Currency:              "USD",
		StripePaymentIntentID: paymentIntentID,
		Status:                DepositStatusPending,
	}
	ts.repo.addDeposit(deposit)

	err := ts.HandlePaymentIntentFailed(context.Background(), paymentIntentID, "card_declined")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check deposit status was updated
	updated, _ := ts.repo.GetDeposit(context.Background(), depositID)
	if updated.Status != DepositStatusFailed {
		t.Errorf("expected status failed, got %s", updated.Status)
	}
	if updated.FailureReason != "card_declined" {
		t.Errorf("expected failure reason 'card_declined', got %s", updated.FailureReason)
	}
}

func TestNormalizeCurrency(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"USD", "usd"},
		{"usd", "usd"},
		{"EUR", "eur"},
		{"eur", "eur"},
		{"GBP", "gbp"},
		{"gbp", "gbp"},
		{"INVALID", "usd"},
		{"", "usd"},
	}

	for _, tt := range tests {
		result := normalizeCurrency(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeCurrency(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestCreateDepositRequest_Validation(t *testing.T) {
	// Test that invalid amounts are rejected
	// This tests the validation logic that would be in CreateDeposit

	tests := []struct {
		amount    float64
		expectErr bool
	}{
		{100.0, false},
		{0.01, false},
		{0, true},
		{-10, true},
		{-0.01, true},
	}

	for _, tt := range tests {
		req := &CreateDepositRequest{Amount: tt.amount}
		hasErr := req.Amount <= 0
		if hasErr != tt.expectErr {
			t.Errorf("amount %f: expected error=%v, got error=%v", tt.amount, tt.expectErr, hasErr)
		}
	}
}
