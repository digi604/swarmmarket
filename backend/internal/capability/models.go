package capability

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// VerificationLevel represents the trust level of a capability.
type VerificationLevel string

const (
	VerificationUnverified VerificationLevel = "unverified"
	VerificationTested     VerificationLevel = "tested"
	VerificationVerified   VerificationLevel = "verified"
	VerificationCertified  VerificationLevel = "certified"
)

// PricingModel represents how a capability is priced.
type PricingModel string

const (
	PricingFixed      PricingModel = "fixed"
	PricingPercentage PricingModel = "percentage"
	PricingTiered     PricingModel = "tiered"
	PricingCustom     PricingModel = "custom"
)

// GeographicScope represents the geographic reach of a capability.
type GeographicScope string

const (
	ScopeLocal         GeographicScope = "local"
	ScopeRegional      GeographicScope = "regional"
	ScopeNational      GeographicScope = "national"
	ScopeInternational GeographicScope = "international"
)

// Capability represents what an agent can do.
type Capability struct {
	ID      uuid.UUID `json:"id" db:"id"`
	AgentID uuid.UUID `json:"agent_id" db:"agent_id"`

	// Taxonomy
	Domain     string `json:"domain" db:"domain"`
	Type       string `json:"type" db:"type"`
	Subtype    string `json:"subtype,omitempty" db:"subtype"`
	DomainPath string `json:"domain_path" db:"domain_path"`

	// Metadata
	Name        string `json:"name" db:"name"`
	Description string `json:"description,omitempty" db:"description"`
	Version     string `json:"version" db:"version"`

	// Schemas
	InputSchema   json.RawMessage `json:"input_schema" db:"input_schema"`
	OutputSchema  json.RawMessage `json:"output_schema" db:"output_schema"`
	StatusEvents  json.RawMessage `json:"status_events,omitempty" db:"status_events"`

	// Geographic constraints
	GeographicScope GeographicScope `json:"geographic_scope" db:"geographic_scope"`
	GeoCenterLat    *float64        `json:"geo_center_lat,omitempty" db:"geo_center_lat"`
	GeoCenterLng    *float64        `json:"geo_center_lng,omitempty" db:"geo_center_lng"`
	GeoRadiusKM     *int            `json:"geo_radius_km,omitempty" db:"geo_radius_km"`
	GeoPolygon      json.RawMessage `json:"geo_polygon,omitempty" db:"geo_polygon"`

	// Temporal constraints
	AvailableHours string `json:"available_hours,omitempty" db:"available_hours"`
	AvailableDays  string `json:"available_days,omitempty" db:"available_days"`
	Timezone       string `json:"timezone" db:"timezone"`

	// Pricing
	PricingModel   PricingModel    `json:"pricing_model" db:"pricing_model"`
	BaseFee        *float64        `json:"base_fee,omitempty" db:"base_fee"`
	PercentageFee  *float64        `json:"percentage_fee,omitempty" db:"percentage_fee"`
	Currency       string          `json:"currency" db:"currency"`
	PricingDetails json.RawMessage `json:"pricing_details,omitempty" db:"pricing_details"`

	// SLA
	ResponseTimeSeconds *int   `json:"response_time_seconds,omitempty" db:"response_time_seconds"`
	CompletionTimeP50   string `json:"completion_time_p50,omitempty" db:"completion_time_p50"`
	CompletionTimeP95   string `json:"completion_time_p95,omitempty" db:"completion_time_p95"`

	// Status
	IsActive         bool `json:"is_active" db:"is_active"`
	IsAcceptingTasks bool `json:"is_accepting_tasks" db:"is_accepting_tasks"`

	// Stats
	TotalTasks      int     `json:"total_tasks" db:"total_tasks"`
	SuccessfulTasks int     `json:"successful_tasks" db:"successful_tasks"`
	FailedTasks     int     `json:"failed_tasks" db:"failed_tasks"`
	AverageRating   float64 `json:"average_rating" db:"average_rating"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Joined data (not stored, populated on queries)
	Verification *Verification `json:"verification,omitempty" db:"-"`
	AgentName    string        `json:"agent_name,omitempty" db:"agent_name"`
}

// Verification represents the verification status of a capability.
type Verification struct {
	ID           uuid.UUID         `json:"id" db:"id"`
	CapabilityID uuid.UUID         `json:"capability_id" db:"capability_id"`
	Level        VerificationLevel `json:"level" db:"level"`
	Method       string            `json:"method,omitempty" db:"method"`
	Proof        json.RawMessage   `json:"proof,omitempty" db:"proof"`
	TestResults  json.RawMessage   `json:"test_results,omitempty" db:"test_results"`

	// Metrics
	SuccessRate     *float64 `json:"success_rate,omitempty" db:"success_rate"`
	TotalTxns       *int     `json:"total_transactions,omitempty" db:"total_transactions"`
	AvgResponseTime *int     `json:"avg_response_time_ms,omitempty" db:"avg_response_time_ms"`

	// Validity
	VerifiedAt time.Time  `json:"verified_at,omitempty" db:"verified_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	VerifiedBy string     `json:"verified_by,omitempty" db:"verified_by"`
	IsCurrent  bool       `json:"is_current" db:"is_current"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// DomainTaxonomy represents a node in the capability taxonomy tree.
type DomainTaxonomy struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	Path           string          `json:"path" db:"path"`
	ParentPath     *string         `json:"parent_path,omitempty" db:"parent_path"`
	Name           string          `json:"name" db:"name"`
	Description    string          `json:"description,omitempty" db:"description"`
	Icon           string          `json:"icon,omitempty" db:"icon"`
	SchemaTemplate json.RawMessage `json:"schema_template,omitempty" db:"schema_template"`
	IsActive       bool            `json:"is_active" db:"is_active"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`

	// Computed (not stored)
	Children []DomainTaxonomy `json:"children,omitempty" db:"-"`
}

// StatusEvent represents a possible status update during task execution.
type StatusEvent struct {
	Event       string `json:"event"`
	Description string `json:"description"`
}

// GeoConstraint represents geographic constraints for a capability.
type GeoConstraint struct {
	Type     string    `json:"type"` // "radius", "polygon", "countries"
	Center   *GeoPoint `json:"center,omitempty"`
	RadiusKM int       `json:"radius_km,omitempty"`
	Polygon  []GeoPoint `json:"polygon,omitempty"`
	Countries []string  `json:"countries,omitempty"`
}

// GeoPoint represents a geographic coordinate.
type GeoPoint struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// PricingInfo represents detailed pricing information.
type PricingInfo struct {
	Model         PricingModel `json:"model"`
	BaseFee       float64      `json:"base_fee,omitempty"`
	PercentageFee float64      `json:"percentage_fee,omitempty"`
	Currency      string       `json:"currency"`
	MinFee        float64      `json:"min_fee,omitempty"`
	MaxFee        float64      `json:"max_fee,omitempty"`
	Tiers         []PricingTier `json:"tiers,omitempty"`
}

// PricingTier represents a tier in tiered pricing.
type PricingTier struct {
	UpTo float64 `json:"up_to"`
	Fee  float64 `json:"fee"`
}

// SLA represents service level agreement details.
type SLA struct {
	ResponseTimeSeconds int    `json:"response_time_seconds"`
	CompletionTimeP50   string `json:"completion_time_p50"`
	CompletionTimeP95   string `json:"completion_time_p95"`
}

// --- Request/Response DTOs ---

// CreateCapabilityRequest is the request to register a new capability.
type CreateCapabilityRequest struct {
	Domain      string `json:"domain" validate:"required"`
	Type        string `json:"type" validate:"required"`
	Subtype     string `json:"subtype,omitempty"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description,omitempty"`

	InputSchema  json.RawMessage `json:"input_schema" validate:"required"`
	OutputSchema json.RawMessage `json:"output_schema" validate:"required"`
	StatusEvents []StatusEvent   `json:"status_events,omitempty"`

	Geographic *GeoConstraint `json:"geographic,omitempty"`
	Temporal   *TemporalConstraint `json:"temporal,omitempty"`
	Pricing    *PricingInfo   `json:"pricing,omitempty"`
	SLA        *SLA           `json:"sla,omitempty"`
}

// TemporalConstraint represents time-based constraints.
type TemporalConstraint struct {
	AvailableHours string `json:"available_hours,omitempty"` // "09:00-18:00"
	AvailableDays  string `json:"available_days,omitempty"`  // "mon,tue,wed,thu,fri"
	Timezone       string `json:"timezone,omitempty"`
}

// UpdateCapabilityRequest is the request to update a capability.
type UpdateCapabilityRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`

	InputSchema  json.RawMessage `json:"input_schema,omitempty"`
	OutputSchema json.RawMessage `json:"output_schema,omitempty"`
	StatusEvents []StatusEvent   `json:"status_events,omitempty"`

	Geographic *GeoConstraint `json:"geographic,omitempty"`
	Temporal   *TemporalConstraint `json:"temporal,omitempty"`
	Pricing    *PricingInfo   `json:"pricing,omitempty"`
	SLA        *SLA           `json:"sla,omitempty"`

	IsActive         *bool `json:"is_active,omitempty"`
	IsAcceptingTasks *bool `json:"is_accepting_tasks,omitempty"`
}

// SearchCapabilitiesRequest is the request to search capabilities.
type SearchCapabilitiesRequest struct {
	Domain        string   `json:"domain,omitempty"`
	Type          string   `json:"type,omitempty"`
	Subtype       string   `json:"subtype,omitempty"`
	DomainPath    string   `json:"domain_path,omitempty"`
	Query         string   `json:"query,omitempty"` // full-text search

	// Location
	Lat      *float64 `json:"lat,omitempty"`
	Lng      *float64 `json:"lng,omitempty"`
	RadiusKM *int     `json:"radius_km,omitempty"`

	// Filters
	VerifiedOnly  bool     `json:"verified_only,omitempty"`
	MinRating     *float64 `json:"min_rating,omitempty"`
	MaxPrice      *float64 `json:"max_price,omitempty"`
	Currency      string   `json:"currency,omitempty"`
	RequiredInput []string `json:"required_input,omitempty"`

	// Sorting
	SortBy    string `json:"sort_by,omitempty"` // reputation, price, response_time
	SortOrder string `json:"sort_order,omitempty"` // asc, desc

	// Pagination
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

// SearchCapabilitiesResponse is the response from searching capabilities.
type SearchCapabilitiesResponse struct {
	Capabilities []CapabilityMatch `json:"capabilities"`
	Total        int               `json:"total"`
	Limit        int               `json:"limit"`
	Offset       int               `json:"offset"`
	Facets       *SearchFacets     `json:"facets,omitempty"`
}

// CapabilityMatch is a capability with match metadata.
type CapabilityMatch struct {
	Capability
	MatchScore     float64        `json:"match_score,omitempty"`
	DistanceKM     *float64       `json:"distance_km,omitempty"`
	EstimatedPrice *PriceEstimate `json:"estimated_price,omitempty"`
}

// PriceEstimate is an estimated price for a capability.
type PriceEstimate struct {
	MinFee   float64 `json:"min_fee"`
	MaxFee   float64 `json:"max_fee"`
	Currency string  `json:"currency"`
}

// SearchFacets contains aggregated counts for search refinement.
type SearchFacets struct {
	ByDomain       map[string]int            `json:"by_domain,omitempty"`
	ByVerification map[VerificationLevel]int `json:"by_verification,omitempty"`
	ByPricingModel map[PricingModel]int      `json:"by_pricing_model,omitempty"`
}
