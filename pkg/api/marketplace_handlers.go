package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/swarmmarket/swarmmarket/internal/common"
	"github.com/swarmmarket/swarmmarket/internal/marketplace"
	"github.com/swarmmarket/swarmmarket/pkg/middleware"
)

// MarketplaceHandler handles marketplace HTTP requests.
type MarketplaceHandler struct {
	service *marketplace.Service
}

// NewMarketplaceHandler creates a new marketplace handler.
func NewMarketplaceHandler(service *marketplace.Service) *MarketplaceHandler {
	return &MarketplaceHandler{service: service}
}

// --- Listings ---

// CreateListing handles creating a new listing.
func (h *MarketplaceHandler) CreateListing(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	var req marketplace.CreateListingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid request body"))
		return
	}

	listing, err := h.service.CreateListing(r.Context(), agent.ID, &req)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest(err.Error()))
		return
	}

	common.WriteJSON(w, http.StatusCreated, listing)
}

// GetListing handles getting a listing by ID.
func (h *MarketplaceHandler) GetListing(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid listing id"))
		return
	}

	listing, err := h.service.GetListing(r.Context(), id)
	if err != nil {
		if err == marketplace.ErrListingNotFound {
			common.WriteError(w, http.StatusNotFound, common.ErrNotFound("listing not found"))
			return
		}
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to get listing"))
		return
	}

	common.WriteJSON(w, http.StatusOK, listing)
}

// SearchListings handles searching for listings.
func (h *MarketplaceHandler) SearchListings(w http.ResponseWriter, r *http.Request) {
	params := marketplace.SearchListingsParams{
		Query:  r.URL.Query().Get("q"),
		Limit:  parseIntParam(r, "limit", 20),
		Offset: parseIntParam(r, "offset", 0),
	}

	if typeStr := r.URL.Query().Get("type"); typeStr != "" {
		t := marketplace.ListingType(typeStr)
		params.ListingType = &t
	}
	if scopeStr := r.URL.Query().Get("scope"); scopeStr != "" {
		s := marketplace.GeographicScope(scopeStr)
		params.GeographicScope = &s
	}
	if categoryStr := r.URL.Query().Get("category"); categoryStr != "" {
		if catID, err := uuid.Parse(categoryStr); err == nil {
			params.CategoryID = &catID
		}
	}
	if minStr := r.URL.Query().Get("min_price"); minStr != "" {
		if min, err := strconv.ParseFloat(minStr, 64); err == nil {
			params.MinPrice = &min
		}
	}
	if maxStr := r.URL.Query().Get("max_price"); maxStr != "" {
		if max, err := strconv.ParseFloat(maxStr, 64); err == nil {
			params.MaxPrice = &max
		}
	}

	result, err := h.service.SearchListings(r.Context(), params)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to search listings"))
		return
	}

	common.WriteJSON(w, http.StatusOK, result)
}

// DeleteListing handles deleting a listing.
func (h *MarketplaceHandler) DeleteListing(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid listing id"))
		return
	}

	if err := h.service.DeleteListing(r.Context(), id, agent.ID); err != nil {
		if err == marketplace.ErrListingNotFound {
			common.WriteError(w, http.StatusNotFound, common.ErrNotFound("listing not found"))
			return
		}
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to delete listing"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Requests ---

// CreateRequest handles creating a new request.
func (h *MarketplaceHandler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	var req marketplace.CreateRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid request body"))
		return
	}

	request, err := h.service.CreateRequest(r.Context(), agent.ID, &req)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest(err.Error()))
		return
	}

	common.WriteJSON(w, http.StatusCreated, request)
}

// GetRequest handles getting a request by ID.
func (h *MarketplaceHandler) GetRequest(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid request id"))
		return
	}

	request, err := h.service.GetRequest(r.Context(), id)
	if err != nil {
		if err == marketplace.ErrRequestNotFound {
			common.WriteError(w, http.StatusNotFound, common.ErrNotFound("request not found"))
			return
		}
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to get request"))
		return
	}

	common.WriteJSON(w, http.StatusOK, request)
}

// SearchRequests handles searching for requests.
func (h *MarketplaceHandler) SearchRequests(w http.ResponseWriter, r *http.Request) {
	params := marketplace.SearchRequestsParams{
		Query:  r.URL.Query().Get("q"),
		Limit:  parseIntParam(r, "limit", 20),
		Offset: parseIntParam(r, "offset", 0),
	}

	if typeStr := r.URL.Query().Get("type"); typeStr != "" {
		t := marketplace.ListingType(typeStr)
		params.RequestType = &t
	}
	if scopeStr := r.URL.Query().Get("scope"); scopeStr != "" {
		s := marketplace.GeographicScope(scopeStr)
		params.GeographicScope = &s
	}

	result, err := h.service.SearchRequests(r.Context(), params)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to search requests"))
		return
	}

	common.WriteJSON(w, http.StatusOK, result)
}

// --- Offers ---

// SubmitOffer handles submitting an offer to a request.
func (h *MarketplaceHandler) SubmitOffer(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	requestIDStr := chi.URLParam(r, "id")
	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid request id"))
		return
	}

	var req marketplace.CreateOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid request body"))
		return
	}

	offer, err := h.service.SubmitOffer(r.Context(), agent.ID, requestID, &req)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest(err.Error()))
		return
	}

	common.WriteJSON(w, http.StatusCreated, offer)
}

// GetOffers handles getting all offers for a request.
func (h *MarketplaceHandler) GetOffers(w http.ResponseWriter, r *http.Request) {
	requestIDStr := chi.URLParam(r, "id")
	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid request id"))
		return
	}

	offers, err := h.service.GetOffersByRequest(r.Context(), requestID)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to get offers"))
		return
	}

	common.WriteJSON(w, http.StatusOK, map[string]any{"offers": offers})
}

// AcceptOffer handles accepting an offer.
func (h *MarketplaceHandler) AcceptOffer(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	offerIDStr := chi.URLParam(r, "offerId")
	offerID, err := uuid.Parse(offerIDStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid offer id"))
		return
	}

	offer, err := h.service.AcceptOffer(r.Context(), agent.ID, offerID)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest(err.Error()))
		return
	}

	common.WriteJSON(w, http.StatusOK, offer)
}

// --- Categories ---

// GetCategories handles getting all categories.
func (h *MarketplaceHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.service.GetCategories(r.Context())
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to get categories"))
		return
	}

	common.WriteJSON(w, http.StatusOK, map[string]any{"categories": categories})
}

// --- Helpers ---

func parseIntParam(r *http.Request, name string, defaultVal int) int {
	valStr := r.URL.Query().Get(name)
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}
