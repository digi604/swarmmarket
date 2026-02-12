package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/digi604/swarmmarket/backend/internal/agent"
	"github.com/digi604/swarmmarket/backend/internal/transaction"
	"github.com/digi604/swarmmarket/backend/pkg/middleware"
)

// mockTransactionService implements a mock transaction service for testing.
type mockTransactionService struct {
	transactions map[uuid.UUID]*transaction.Transaction
	escrows      map[uuid.UUID]*transaction.EscrowAccount
}

func newMockTransactionService() *mockTransactionService {
	return &mockTransactionService{
		transactions: make(map[uuid.UUID]*transaction.Transaction),
		escrows:      make(map[uuid.UUID]*transaction.EscrowAccount),
	}
}

func (m *mockTransactionService) GetTransaction(ctx context.Context, id uuid.UUID) (*transaction.Transaction, error) {
	tx, ok := m.transactions[id]
	if !ok {
		return nil, transaction.ErrTransactionNotFound
	}
	return tx, nil
}

func (m *mockTransactionService) ListTransactions(ctx context.Context, params transaction.ListTransactionsParams) (*transaction.TransactionListResult, error) {
	var items []*transaction.Transaction
	for _, tx := range m.transactions {
		if params.AgentID != nil {
			if tx.BuyerID == *params.AgentID || tx.SellerID == *params.AgentID {
				items = append(items, tx)
			}
		}
	}
	return &transaction.TransactionListResult{
		Items:  items,
		Total:  len(items),
		Limit:  params.Limit,
		Offset: params.Offset,
	}, nil
}

func (m *mockTransactionService) FundEscrow(ctx context.Context, transactionID, buyerID uuid.UUID) (*transaction.EscrowFundingResult, error) {
	tx, ok := m.transactions[transactionID]
	if !ok {
		return nil, transaction.ErrTransactionNotFound
	}
	if tx.BuyerID != buyerID {
		return nil, transaction.ErrNotAuthorized
	}
	if tx.Status != transaction.StatusPending {
		return nil, transaction.ErrInvalidStatus
	}
	return &transaction.EscrowFundingResult{
		TransactionID:   transactionID,
		PaymentIntentID: "pi_test_123",
		Amount:          tx.Amount,
		Currency:        tx.Currency,
	}, nil
}

func (m *mockTransactionService) MarkDelivered(ctx context.Context, transactionID, sellerID uuid.UUID, proof, message string) (*transaction.Transaction, error) {
	tx, ok := m.transactions[transactionID]
	if !ok {
		return nil, transaction.ErrTransactionNotFound
	}
	if tx.SellerID != sellerID {
		return nil, transaction.ErrNotAuthorized
	}
	if tx.Status != transaction.StatusPending && tx.Status != transaction.StatusEscrowFunded {
		return nil, transaction.ErrInvalidStatus
	}
	tx.Status = transaction.StatusDelivered
	return tx, nil
}

func (m *mockTransactionService) ConfirmDelivery(ctx context.Context, transactionID, buyerID uuid.UUID) (*transaction.Transaction, error) {
	tx, ok := m.transactions[transactionID]
	if !ok {
		return nil, transaction.ErrTransactionNotFound
	}
	if tx.BuyerID != buyerID {
		return nil, transaction.ErrNotAuthorized
	}
	if tx.Status != transaction.StatusDelivered {
		return nil, transaction.ErrInvalidStatus
	}
	tx.Status = transaction.StatusCompleted
	return tx, nil
}

func (m *mockTransactionService) SubmitRating(ctx context.Context, transactionID, raterID uuid.UUID, req *transaction.SubmitRatingRequest) (*transaction.Rating, error) {
	tx, ok := m.transactions[transactionID]
	if !ok {
		return nil, transaction.ErrTransactionNotFound
	}
	if tx.BuyerID != raterID && tx.SellerID != raterID {
		return nil, transaction.ErrNotAuthorized
	}
	if tx.Status != transaction.StatusCompleted && tx.Status != transaction.StatusDelivered {
		return nil, transaction.ErrTransactionNotReady
	}
	if req.Score < 1 || req.Score > 5 {
		return nil, transaction.ErrInvalidRating
	}

	ratedID := tx.SellerID
	if raterID == tx.SellerID {
		ratedID = tx.BuyerID
	}

	return &transaction.Rating{
		ID:            uuid.New(),
		TransactionID: transactionID,
		RaterID:       raterID,
		RatedAgentID:  ratedID,
		Score:         req.Score,
		Comment:       req.Comment,
	}, nil
}

func (m *mockTransactionService) GetTransactionRatings(ctx context.Context, transactionID uuid.UUID) ([]*transaction.Rating, error) {
	return []*transaction.Rating{}, nil
}

func (m *mockTransactionService) DisputeTransaction(ctx context.Context, transactionID, agentID uuid.UUID, req *transaction.DisputeRequest) (*transaction.Transaction, error) {
	tx, ok := m.transactions[transactionID]
	if !ok {
		return nil, transaction.ErrTransactionNotFound
	}
	if tx.BuyerID != agentID && tx.SellerID != agentID {
		return nil, transaction.ErrNotAuthorized
	}
	tx.Status = transaction.StatusDisputed
	return tx, nil
}

func (m *mockTransactionService) addTransaction(tx *transaction.Transaction) {
	m.transactions[tx.ID] = tx
}

// Test helper to create authenticated request
func createAuthenticatedRequest(t *testing.T, method, path string, body interface{}, agentID uuid.UUID) *http.Request {
	var bodyReader *bytes.Buffer
	if body != nil {
		bodyBytes, _ := json.Marshal(body)
		bodyReader = bytes.NewBuffer(bodyBytes)
	} else {
		bodyReader = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")

	// Add agent to context
	testAgent := &agent.Agent{ID: agentID, Name: "TestAgent"}
	ctx := context.WithValue(req.Context(), middleware.AgentContextKey, testAgent)
	return req.WithContext(ctx)
}

func TestOrderHandler_ListOrders(t *testing.T) {
	mockService := newMockTransactionService()
	handler := NewOrderHandler(mockService)

	buyerID := uuid.New()
	sellerID := uuid.New()

	// Add test transaction
	tx := &transaction.Transaction{
		ID:       uuid.New(),
		BuyerID:  buyerID,
		SellerID: sellerID,
		Amount:   100.00,
		Currency: "USD",
		Status:   transaction.StatusPending,
	}
	mockService.addTransaction(tx)

	req := createAuthenticatedRequest(t, "GET", "/orders", nil, buyerID)
	rr := httptest.NewRecorder()

	handler.ListOrders(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var result transaction.TransactionListResult
	json.NewDecoder(rr.Body).Decode(&result)

	if len(result.Items) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(result.Items))
	}
}

func TestOrderHandler_GetOrder(t *testing.T) {
	mockService := newMockTransactionService()
	handler := NewOrderHandler(mockService)

	buyerID := uuid.New()
	sellerID := uuid.New()
	txID := uuid.New()

	tx := &transaction.Transaction{
		ID:       txID,
		BuyerID:  buyerID,
		SellerID: sellerID,
		Amount:   50.00,
		Currency: "USD",
		Status:   transaction.StatusPending,
	}
	mockService.addTransaction(tx)

	// Test as buyer
	req := createAuthenticatedRequest(t, "GET", "/orders/"+txID.String(), nil, buyerID)

	// Set up chi URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", txID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.GetOrder(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestOrderHandler_GetOrder_NotFound(t *testing.T) {
	mockService := newMockTransactionService()
	handler := NewOrderHandler(mockService)

	agentID := uuid.New()
	nonExistentID := uuid.New()

	req := createAuthenticatedRequest(t, "GET", "/orders/"+nonExistentID.String(), nil, agentID)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", nonExistentID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.GetOrder(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}
}

func TestOrderHandler_FundEscrow(t *testing.T) {
	mockService := newMockTransactionService()
	handler := NewOrderHandler(mockService)

	buyerID := uuid.New()
	sellerID := uuid.New()
	txID := uuid.New()

	tx := &transaction.Transaction{
		ID:       txID,
		BuyerID:  buyerID,
		SellerID: sellerID,
		Amount:   100.00,
		Currency: "USD",
		Status:   transaction.StatusPending,
	}
	mockService.addTransaction(tx)

	req := createAuthenticatedRequest(t, "POST", "/orders/"+txID.String()+"/fund", nil, buyerID)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", txID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.FundEscrow(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result transaction.EscrowFundingResult
	json.NewDecoder(rr.Body).Decode(&result)

	if result.PaymentIntentID == "" {
		t.Error("expected payment intent ID")
	}
}

func TestOrderHandler_FundEscrow_NotBuyer(t *testing.T) {
	mockService := newMockTransactionService()
	handler := NewOrderHandler(mockService)

	buyerID := uuid.New()
	sellerID := uuid.New()
	txID := uuid.New()

	tx := &transaction.Transaction{
		ID:       txID,
		BuyerID:  buyerID,
		SellerID: sellerID,
		Amount:   100.00,
		Currency: "USD",
		Status:   transaction.StatusPending,
	}
	mockService.addTransaction(tx)

	// Try to fund as seller (should fail)
	req := createAuthenticatedRequest(t, "POST", "/orders/"+txID.String()+"/fund", nil, sellerID)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", txID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.FundEscrow(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rr.Code)
	}
}

func TestOrderHandler_MarkDelivered(t *testing.T) {
	mockService := newMockTransactionService()
	handler := NewOrderHandler(mockService)

	buyerID := uuid.New()
	sellerID := uuid.New()
	txID := uuid.New()

	tx := &transaction.Transaction{
		ID:       txID,
		BuyerID:  buyerID,
		SellerID: sellerID,
		Amount:   100.00,
		Currency: "USD",
		Status:   transaction.StatusEscrowFunded,
	}
	mockService.addTransaction(tx)

	body := map[string]string{
		"delivery_proof": "https://example.com/data",
		"message":        "Data ready",
	}
	req := createAuthenticatedRequest(t, "POST", "/orders/"+txID.String()+"/deliver", body, sellerID)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", txID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.MarkDelivered(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestOrderHandler_MarkDelivered_NotSeller(t *testing.T) {
	mockService := newMockTransactionService()
	handler := NewOrderHandler(mockService)

	buyerID := uuid.New()
	sellerID := uuid.New()
	txID := uuid.New()

	tx := &transaction.Transaction{
		ID:       txID,
		BuyerID:  buyerID,
		SellerID: sellerID,
		Amount:   100.00,
		Currency: "USD",
		Status:   transaction.StatusEscrowFunded,
	}
	mockService.addTransaction(tx)

	// Try to deliver as buyer (should fail)
	req := createAuthenticatedRequest(t, "POST", "/orders/"+txID.String()+"/deliver", nil, buyerID)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", txID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.MarkDelivered(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rr.Code)
	}
}

func TestOrderHandler_ConfirmDelivery(t *testing.T) {
	mockService := newMockTransactionService()
	handler := NewOrderHandler(mockService)

	buyerID := uuid.New()
	sellerID := uuid.New()
	txID := uuid.New()

	tx := &transaction.Transaction{
		ID:       txID,
		BuyerID:  buyerID,
		SellerID: sellerID,
		Amount:   100.00,
		Currency: "USD",
		Status:   transaction.StatusDelivered,
	}
	mockService.addTransaction(tx)

	req := createAuthenticatedRequest(t, "POST", "/orders/"+txID.String()+"/confirm", nil, buyerID)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", txID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.ConfirmDelivery(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestOrderHandler_SubmitRating(t *testing.T) {
	mockService := newMockTransactionService()
	handler := NewOrderHandler(mockService)

	buyerID := uuid.New()
	sellerID := uuid.New()
	txID := uuid.New()

	tx := &transaction.Transaction{
		ID:       txID,
		BuyerID:  buyerID,
		SellerID: sellerID,
		Amount:   100.00,
		Currency: "USD",
		Status:   transaction.StatusCompleted,
	}
	mockService.addTransaction(tx)

	body := transaction.SubmitRatingRequest{
		Score:   5,
		Comment: "Excellent service!",
	}
	req := createAuthenticatedRequest(t, "POST", "/orders/"+txID.String()+"/rating", body, buyerID)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", txID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.SubmitRating(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestOrderHandler_SubmitRating_InvalidScore(t *testing.T) {
	mockService := newMockTransactionService()
	handler := NewOrderHandler(mockService)

	buyerID := uuid.New()
	sellerID := uuid.New()
	txID := uuid.New()

	tx := &transaction.Transaction{
		ID:       txID,
		BuyerID:  buyerID,
		SellerID: sellerID,
		Amount:   100.00,
		Currency: "USD",
		Status:   transaction.StatusCompleted,
	}
	mockService.addTransaction(tx)

	body := transaction.SubmitRatingRequest{
		Score:   6, // Invalid - should be 1-5
		Comment: "Invalid rating",
	}
	req := createAuthenticatedRequest(t, "POST", "/orders/"+txID.String()+"/rating", body, buyerID)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", txID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.SubmitRating(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestOrderHandler_DisputeOrder(t *testing.T) {
	mockService := newMockTransactionService()
	handler := NewOrderHandler(mockService)

	buyerID := uuid.New()
	sellerID := uuid.New()
	txID := uuid.New()

	tx := &transaction.Transaction{
		ID:       txID,
		BuyerID:  buyerID,
		SellerID: sellerID,
		Amount:   100.00,
		Currency: "USD",
		Status:   transaction.StatusDelivered,
	}
	mockService.addTransaction(tx)

	body := transaction.DisputeRequest{
		Reason:      "Item not as described",
		Description: "Missing data fields",
	}
	req := createAuthenticatedRequest(t, "POST", "/orders/"+txID.String()+"/dispute", body, buyerID)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", txID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.DisputeOrder(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestOrderHandler_Unauthenticated(t *testing.T) {
	mockService := newMockTransactionService()
	handler := NewOrderHandler(mockService)

	// Request without authentication
	req := httptest.NewRequest("GET", "/orders", nil)
	rr := httptest.NewRecorder()

	handler.ListOrders(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}
