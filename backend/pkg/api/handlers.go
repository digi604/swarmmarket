package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/digi604/swarmmarket/backend/internal/agent"
	"github.com/digi604/swarmmarket/backend/internal/common"
	"github.com/digi604/swarmmarket/backend/pkg/middleware"
)

// AgentHandler handles agent-related HTTP requests.
type AgentHandler struct {
	service *agent.Service
}

// NewAgentHandler creates a new agent handler.
func NewAgentHandler(service *agent.Service) *AgentHandler {
	return &AgentHandler{service: service}
}

// Register handles agent registration.
func (h *AgentHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req agent.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid request body"))
		return
	}

	if req.Name == "" {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("name is required"))
		return
	}
	if req.OwnerEmail == "" {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("owner_email is required"))
		return
	}

	resp, err := h.service.Register(r.Context(), &req)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to register agent"))
		return
	}

	common.WriteJSON(w, http.StatusCreated, resp)
}

// GetMe returns the current authenticated agent.
func (h *AgentHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	ag := middleware.GetAgent(r.Context())
	if ag == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	common.WriteJSON(w, http.StatusOK, ag)
}

// GetByID returns an agent's public profile by ID.
func (h *AgentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid agent id"))
		return
	}

	profile, err := h.service.GetPublicProfile(r.Context(), id)
	if err != nil {
		if err == agent.ErrAgentNotFound {
			common.WriteError(w, http.StatusNotFound, common.ErrNotFound("agent not found"))
			return
		}
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to get agent"))
		return
	}

	common.WriteJSON(w, http.StatusOK, profile)
}

// GetReputation returns an agent's reputation details.
func (h *AgentHandler) GetReputation(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid agent id"))
		return
	}

	rep, err := h.service.GetReputation(r.Context(), id)
	if err != nil {
		if err == agent.ErrAgentNotFound {
			common.WriteError(w, http.StatusNotFound, common.ErrNotFound("agent not found"))
			return
		}
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to get reputation"))
		return
	}

	common.WriteJSON(w, http.StatusOK, rep)
}

// Update updates the current agent's profile.
func (h *AgentHandler) Update(w http.ResponseWriter, r *http.Request) {
	ag := middleware.GetAgent(r.Context())
	if ag == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	var req agent.UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid request body"))
		return
	}

	updated, err := h.service.Update(r.Context(), ag.ID, &req)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to update agent"))
		return
	}

	common.WriteJSON(w, http.StatusOK, updated)
}

// HealthHandler handles health check requests.
type HealthHandler struct {
	db    HealthChecker
	redis HealthChecker
}

// HealthChecker interface for health checks.
type HealthChecker interface {
	Health(ctx context.Context) error
}

// NewHealthHandler creates a new health handler.
func NewHealthHandler(db, redis HealthChecker) *HealthHandler {
	return &HealthHandler{db: db, redis: redis}
}

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status   string            `json:"status"`
	Services map[string]string `json:"services"`
}

// Check performs a health check.
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	resp := HealthResponse{
		Status:   "healthy",
		Services: make(map[string]string),
	}

	// Check database
	if h.db != nil {
		if err := h.db.Health(r.Context()); err != nil {
			resp.Status = "unhealthy"
			resp.Services["database"] = "unhealthy: " + err.Error()
		} else {
			resp.Services["database"] = "healthy"
		}
	}

	// Check Redis
	if h.redis != nil {
		if err := h.redis.Health(r.Context()); err != nil {
			resp.Status = "unhealthy"
			resp.Services["redis"] = "unhealthy: " + err.Error()
		} else {
			resp.Services["redis"] = "healthy"
		}
	}

	statusCode := http.StatusOK
	if resp.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	common.WriteJSON(w, statusCode, resp)
}

// Ready checks if the service is ready to accept traffic.
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	common.WriteJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

// Live checks if the service is alive.
func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	common.WriteJSON(w, http.StatusOK, map[string]string{"status": "alive"})
}
