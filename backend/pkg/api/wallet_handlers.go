package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/digi604/swarmmarket/backend/internal/wallet"
	"github.com/digi604/swarmmarket/backend/pkg/middleware"
)

// WalletHandler handles wallet-related HTTP requests for human users (Clerk auth).
type WalletHandler struct {
	service *wallet.Service
}

// NewWalletHandler creates a new wallet handler.
func NewWalletHandler(service *wallet.Service) *WalletHandler {
	return &WalletHandler{service: service}
}

// CreateDeposit handles POST /api/v1/dashboard/wallet/deposit (human users)
func (h *WalletHandler) CreateDeposit(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	if user == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req wallet.CreateDepositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Amount <= 0 {
		http.Error(w, `{"error":"amount must be greater than 0"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.service.CreateDeposit(r.Context(), user.ID, &req)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetBalance handles GET /api/v1/dashboard/wallet/balance (human users)
func (h *WalletHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	if user == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	balance, err := h.service.GetWalletBalance(r.Context(), user.ID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balance)
}

// GetDeposits handles GET /api/v1/dashboard/wallet/deposits (human users)
func (h *WalletHandler) GetDeposits(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	if user == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	limit := 20
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	deposits, total, err := h.service.GetUserDeposits(r.Context(), user.ID, limit, offset)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"deposits": deposits,
		"total":    total,
	})
}

// AgentWalletHandler handles wallet-related HTTP requests for agents (API key auth).
type AgentWalletHandler struct {
	service *wallet.Service
}

// NewAgentWalletHandler creates a new agent wallet handler.
func NewAgentWalletHandler(service *wallet.Service) *AgentWalletHandler {
	return &AgentWalletHandler{service: service}
}

// CreateDeposit handles POST /api/v1/wallet/deposit (agents)
func (h *AgentWalletHandler) CreateDeposit(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req wallet.CreateDepositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Amount <= 0 {
		http.Error(w, `{"error":"amount must be greater than 0"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.service.CreateAgentDeposit(r.Context(), agent.ID, &req)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetBalance handles GET /api/v1/wallet/balance (agents)
func (h *AgentWalletHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	balance, err := h.service.GetAgentWalletBalance(r.Context(), agent.ID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balance)
}

// GetDeposits handles GET /api/v1/wallet/deposits (agents)
func (h *AgentWalletHandler) GetDeposits(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	limit := 20
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	deposits, total, err := h.service.GetAgentDeposits(r.Context(), agent.ID, limit, offset)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"deposits": deposits,
		"total":    total,
	})
}
