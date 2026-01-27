package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/digi604/swarmmarket/backend/internal/capability"
)

// CapabilityHandlers handles capability-related HTTP requests.
type CapabilityHandlers struct {
	service *capability.Service
}

// NewCapabilityHandlers creates new capability handlers.
func NewCapabilityHandlers(service *capability.Service) *CapabilityHandlers {
	return &CapabilityHandlers{service: service}
}

// RegisterRoutes registers capability routes.
func (h *CapabilityHandlers) RegisterRoutes(r chi.Router) {
	r.Route("/capabilities", func(r chi.Router) {
		r.Get("/", h.Search)
		r.Get("/domains", h.GetDomains)
		r.Get("/domains/tree", h.GetDomainsTree)

		r.Group(func(r chi.Router) {
			// These routes require authentication
			r.Post("/", h.Create)
		})

		r.Route("/{capabilityID}", func(r chi.Router) {
			r.Get("/", h.GetByID)
			r.Put("/", h.Update)
			r.Delete("/", h.Delete)
			r.Post("/verify", h.RequestVerification)
			r.Get("/verification", h.GetVerification)
		})
	})

	// Agent-specific capability routes
	r.Get("/agents/{agentID}/capabilities", h.GetByAgentID)
}

// Create handles POST /capabilities
func (h *CapabilityHandlers) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get agent from context (set by auth middleware)
	agentID, ok := ctx.Value("agent_id").(uuid.UUID)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req capability.CreateCapabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cap, err := h.service.Create(ctx, agentID, &req)
	if err != nil {
		if errors.Is(err, capability.ErrInvalidDomain) {
			respondError(w, http.StatusBadRequest, "invalid domain/type/subtype")
			return
		}
		if errors.Is(err, capability.ErrInvalidSchema) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to create capability")
		return
	}

	respondJSON(w, http.StatusCreated, cap)
}

// GetByID handles GET /capabilities/{capabilityID}
func (h *CapabilityHandlers) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	capabilityID, err := uuid.Parse(chi.URLParam(r, "capabilityID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid capability ID")
		return
	}

	cap, err := h.service.GetByID(ctx, capabilityID)
	if err != nil {
		if errors.Is(err, capability.ErrCapabilityNotFound) {
			respondError(w, http.StatusNotFound, "capability not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get capability")
		return
	}

	respondJSON(w, http.StatusOK, cap)
}

// GetByAgentID handles GET /agents/{agentID}/capabilities
func (h *CapabilityHandlers) GetByAgentID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	agentID, err := uuid.Parse(chi.URLParam(r, "agentID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid agent ID")
		return
	}

	caps, err := h.service.GetByAgentID(ctx, agentID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get capabilities")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"capabilities": caps,
	})
}

// Update handles PUT /capabilities/{capabilityID}
func (h *CapabilityHandlers) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	agentID, ok := ctx.Value("agent_id").(uuid.UUID)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	capabilityID, err := uuid.Parse(chi.URLParam(r, "capabilityID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid capability ID")
		return
	}

	var req capability.UpdateCapabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cap, err := h.service.Update(ctx, agentID, capabilityID, &req)
	if err != nil {
		if errors.Is(err, capability.ErrCapabilityNotFound) {
			respondError(w, http.StatusNotFound, "capability not found")
			return
		}
		if errors.Is(err, capability.ErrUnauthorized) {
			respondError(w, http.StatusForbidden, "not your capability")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to update capability")
		return
	}

	respondJSON(w, http.StatusOK, cap)
}

// Delete handles DELETE /capabilities/{capabilityID}
func (h *CapabilityHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	agentID, ok := ctx.Value("agent_id").(uuid.UUID)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	capabilityID, err := uuid.Parse(chi.URLParam(r, "capabilityID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid capability ID")
		return
	}

	err = h.service.Delete(ctx, agentID, capabilityID)
	if err != nil {
		if errors.Is(err, capability.ErrCapabilityNotFound) {
			respondError(w, http.StatusNotFound, "capability not found")
			return
		}
		if errors.Is(err, capability.ErrUnauthorized) {
			respondError(w, http.StatusForbidden, "not your capability")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to delete capability")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// Search handles GET /capabilities
func (h *CapabilityHandlers) Search(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()

	req := &capability.SearchCapabilitiesRequest{
		Domain:       q.Get("domain"),
		Type:         q.Get("type"),
		Subtype:      q.Get("subtype"),
		DomainPath:   q.Get("domain_path"),
		Query:        q.Get("q"),
		Currency:     q.Get("currency"),
		SortBy:       q.Get("sort_by"),
		SortOrder:    q.Get("sort_order"),
		VerifiedOnly: q.Get("verified_only") == "true",
	}

	// Parse location
	if lat := q.Get("lat"); lat != "" {
		if v, err := strconv.ParseFloat(lat, 64); err == nil {
			req.Lat = &v
		}
	}
	if lng := q.Get("lng"); lng != "" {
		if v, err := strconv.ParseFloat(lng, 64); err == nil {
			req.Lng = &v
		}
	}
	if radius := q.Get("radius_km"); radius != "" {
		if v, err := strconv.Atoi(radius); err == nil {
			req.RadiusKM = &v
		}
	}

	// Parse filters
	if minRating := q.Get("min_rating"); minRating != "" {
		if v, err := strconv.ParseFloat(minRating, 64); err == nil {
			req.MinRating = &v
		}
	}
	if maxPrice := q.Get("max_price"); maxPrice != "" {
		if v, err := strconv.ParseFloat(maxPrice, 64); err == nil {
			req.MaxPrice = &v
		}
	}

	// Parse pagination
	if limit := q.Get("limit"); limit != "" {
		if v, err := strconv.Atoi(limit); err == nil {
			req.Limit = v
		}
	}
	if offset := q.Get("offset"); offset != "" {
		if v, err := strconv.Atoi(offset); err == nil {
			req.Offset = v
		}
	}

	result, err := h.service.Search(ctx, req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to search capabilities")
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// GetDomains handles GET /capabilities/domains
func (h *CapabilityHandlers) GetDomains(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	domains, err := h.service.GetDomainTaxonomy(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get domains")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"domains": domains,
	})
}

// GetDomainsTree handles GET /capabilities/domains/tree
func (h *CapabilityHandlers) GetDomainsTree(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tree, err := h.service.GetDomainTaxonomyTree(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get domain tree")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"domains": tree,
	})
}

// RequestVerification handles POST /capabilities/{capabilityID}/verify
func (h *CapabilityHandlers) RequestVerification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	agentID, ok := ctx.Value("agent_id").(uuid.UUID)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	capabilityID, err := uuid.Parse(chi.URLParam(r, "capabilityID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid capability ID")
		return
	}

	var req struct {
		Method string `json:"method"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Method = "api_test" // default
	}

	verification, err := h.service.RequestVerification(ctx, agentID, capabilityID, req.Method)
	if err != nil {
		if errors.Is(err, capability.ErrCapabilityNotFound) {
			respondError(w, http.StatusNotFound, "capability not found")
			return
		}
		if errors.Is(err, capability.ErrUnauthorized) {
			respondError(w, http.StatusForbidden, "not your capability")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to request verification")
		return
	}

	respondJSON(w, http.StatusOK, verification)
}

// GetVerification handles GET /capabilities/{capabilityID}/verification
func (h *CapabilityHandlers) GetVerification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	capabilityID, err := uuid.Parse(chi.URLParam(r, "capabilityID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid capability ID")
		return
	}

	verification, err := h.service.GetVerification(ctx, capabilityID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get verification")
		return
	}

	if verification == nil {
		respondError(w, http.StatusNotFound, "no verification found")
		return
	}

	respondJSON(w, http.StatusOK, verification)
}

// Helper functions
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
