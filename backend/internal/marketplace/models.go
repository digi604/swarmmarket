package marketplace

import (
	"time"

	"github.com/google/uuid"
)

// ListingType represents the type of listing.
type ListingType string

const (
	ListingTypeGoods    ListingType = "goods"
	ListingTypeServices ListingType = "services"
	ListingTypeData     ListingType = "data"
)

// ListingStatus represents the status of a listing.
type ListingStatus string

const (
	ListingStatusDraft   ListingStatus = "draft"
	ListingStatusActive  ListingStatus = "active"
	ListingStatusPaused  ListingStatus = "paused"
	ListingStatusSold    ListingStatus = "sold"
	ListingStatusExpired ListingStatus = "expired"
)

// GeographicScope represents the geographic reach of a listing/request.
type GeographicScope string

const (
	ScopeLocal         GeographicScope = "local"
	ScopeRegional      GeographicScope = "regional"
	ScopeNational      GeographicScope = "national"
	ScopeInternational GeographicScope = "international"
)

// Listing represents something an agent is offering (goods/services/data).
type Listing struct {
	ID              uuid.UUID       `json:"id"`
	SellerID        uuid.UUID       `json:"seller_id"`
	CategoryID      *uuid.UUID      `json:"category_id,omitempty"`
	Title           string          `json:"title"`
	Description     string          `json:"description,omitempty"`
	ListingType     ListingType     `json:"listing_type"`
	PriceAmount     *float64        `json:"price_amount,omitempty"`
	PriceCurrency   string          `json:"price_currency"`
	Quantity        int             `json:"quantity"`
	GeographicScope GeographicScope `json:"geographic_scope"`
	LocationLat     *float64        `json:"location_lat,omitempty"`
	LocationLng     *float64        `json:"location_lng,omitempty"`
	LocationRadius  *int            `json:"location_radius_km,omitempty"`
	Status          ListingStatus   `json:"status"`
	ExpiresAt       *time.Time      `json:"expires_at,omitempty"`
	Metadata        map[string]any  `json:"metadata,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// RequestStatus represents the status of a request.
type RequestStatus string

const (
	RequestStatusOpen       RequestStatus = "open"
	RequestStatusInProgress RequestStatus = "in_progress"
	RequestStatusFulfilled  RequestStatus = "fulfilled"
	RequestStatusCancelled  RequestStatus = "cancelled"
	RequestStatusExpired    RequestStatus = "expired"
)

// Request represents something an agent needs (reverse auction style).
// Examples: "I need a pizza delivered", "I need 200kg of sugar"
type Request struct {
	ID              uuid.UUID       `json:"id"`
	RequesterID     uuid.UUID       `json:"requester_id"`
	CategoryID      *uuid.UUID      `json:"category_id,omitempty"`
	Title           string          `json:"title"`
	Description     string          `json:"description,omitempty"`
	RequestType     ListingType     `json:"request_type"` // goods, services, data
	BudgetMin       *float64        `json:"budget_min,omitempty"`
	BudgetMax       *float64        `json:"budget_max,omitempty"`
	BudgetCurrency  string          `json:"budget_currency"`
	Quantity        int             `json:"quantity"`
	GeographicScope GeographicScope `json:"geographic_scope"`
	LocationLat     *float64        `json:"location_lat,omitempty"`
	LocationLng     *float64        `json:"location_lng,omitempty"`
	LocationRadius  *int            `json:"location_radius_km,omitempty"`
	Status          RequestStatus   `json:"status"`
	ExpiresAt       *time.Time      `json:"expires_at,omitempty"`
	Metadata        map[string]any  `json:"metadata,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`

	// Counts for display
	OfferCount int `json:"offer_count,omitempty"`
}

// OfferStatus represents the status of an offer.
type OfferStatus string

const (
	OfferStatusPending   OfferStatus = "pending"
	OfferStatusAccepted  OfferStatus = "accepted"
	OfferStatusRejected  OfferStatus = "rejected"
	OfferStatusWithdrawn OfferStatus = "withdrawn"
	OfferStatusExpired   OfferStatus = "expired"
)

// Offer represents a response to a request.
// Example: Agent B offers to fulfill Agent A's request for sugar.
type Offer struct {
	ID            uuid.UUID      `json:"id"`
	RequestID     uuid.UUID      `json:"request_id"`
	OffererID     uuid.UUID      `json:"offerer_id"`
	ListingID     *uuid.UUID     `json:"listing_id,omitempty"` // Optional link to existing listing
	PriceAmount   float64        `json:"price_amount"`
	PriceCurrency string         `json:"price_currency"`
	Description   string         `json:"description,omitempty"`
	DeliveryTerms string         `json:"delivery_terms,omitempty"`
	ValidUntil    *time.Time     `json:"valid_until,omitempty"`
	Status        OfferStatus    `json:"status"`
	Metadata      map[string]any `json:"metadata,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// Category represents a category in the taxonomy.
type Category struct {
	ID          uuid.UUID      `json:"id"`
	ParentID    *uuid.UUID     `json:"parent_id,omitempty"`
	Name        string         `json:"name"`
	Slug        string         `json:"slug"`
	Description string         `json:"description,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}

// --- Request/Response DTOs ---

// CreateListingRequest is the request body for creating a listing.
type CreateListingRequest struct {
	CategoryID      *uuid.UUID      `json:"category_id,omitempty"`
	Title           string          `json:"title"`
	Description     string          `json:"description,omitempty"`
	ListingType     ListingType     `json:"listing_type"`
	PriceAmount     *float64        `json:"price_amount,omitempty"`
	PriceCurrency   string          `json:"price_currency,omitempty"`
	Quantity        int             `json:"quantity,omitempty"`
	GeographicScope GeographicScope `json:"geographic_scope,omitempty"`
	LocationLat     *float64        `json:"location_lat,omitempty"`
	LocationLng     *float64        `json:"location_lng,omitempty"`
	LocationRadius  *int            `json:"location_radius_km,omitempty"`
	ExpiresAt       *time.Time      `json:"expires_at,omitempty"`
	Metadata        map[string]any  `json:"metadata,omitempty"`
}

// CreateRequestRequest is the request body for creating a request.
type CreateRequestRequest struct {
	CategoryID      *uuid.UUID      `json:"category_id,omitempty"`
	Title           string          `json:"title"`
	Description     string          `json:"description,omitempty"`
	RequestType     ListingType     `json:"request_type"`
	BudgetMin       *float64        `json:"budget_min,omitempty"`
	BudgetMax       *float64        `json:"budget_max,omitempty"`
	BudgetCurrency  string          `json:"budget_currency,omitempty"`
	Quantity        int             `json:"quantity,omitempty"`
	GeographicScope GeographicScope `json:"geographic_scope,omitempty"`
	LocationLat     *float64        `json:"location_lat,omitempty"`
	LocationLng     *float64        `json:"location_lng,omitempty"`
	LocationRadius  *int            `json:"location_radius_km,omitempty"`
	ExpiresAt       *time.Time      `json:"expires_at,omitempty"`
	Metadata        map[string]any  `json:"metadata,omitempty"`
}

// CreateOfferRequest is the request body for submitting an offer.
type CreateOfferRequest struct {
	ListingID     *uuid.UUID `json:"listing_id,omitempty"`
	PriceAmount   float64    `json:"price_amount"`
	PriceCurrency string     `json:"price_currency,omitempty"`
	Description   string     `json:"description,omitempty"`
	DeliveryTerms string     `json:"delivery_terms,omitempty"`
	ValidUntil    *time.Time `json:"valid_until,omitempty"`
}

// SearchListingsParams contains search/filter parameters for listings.
type SearchListingsParams struct {
	CategoryID      *uuid.UUID
	ListingType     *ListingType
	MinPrice        *float64
	MaxPrice        *float64
	GeographicScope *GeographicScope
	SellerID        *uuid.UUID
	Status          *ListingStatus
	Query           string // Full-text search
	Limit           int
	Offset          int
}

// SearchRequestsParams contains search/filter parameters for requests.
type SearchRequestsParams struct {
	CategoryID      *uuid.UUID
	RequestType     *ListingType
	MinBudget       *float64
	MaxBudget       *float64
	GeographicScope *GeographicScope
	RequesterID     *uuid.UUID
	Status          *RequestStatus
	Query           string
	Limit           int
	Offset          int
}

// ListResult is a generic paginated list result.
type ListResult[T any] struct {
	Items  []T `json:"items"`
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}
