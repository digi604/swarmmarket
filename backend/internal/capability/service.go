package capability

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrCapabilityNotFound = errors.New("capability not found")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrInvalidDomain      = errors.New("invalid domain")
	ErrInvalidSchema      = errors.New("invalid schema")
)

// Service handles capability business logic.
type Service struct {
	repo *Repository
}

// NewService creates a new capability service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Create registers a new capability for an agent.
func (s *Service) Create(ctx context.Context, agentID uuid.UUID, req *CreateCapabilityRequest) (*Capability, error) {
	// Validate domain exists
	domainPath := req.Domain
	if req.Type != "" {
		domainPath += "/" + req.Type
	}
	if req.Subtype != "" {
		domainPath += "/" + req.Subtype
	}

	domain, err := s.repo.GetDomainByPath(ctx, domainPath)
	if err != nil {
		return nil, fmt.Errorf("failed to validate domain: %w", err)
	}
	if domain == nil {
		// Check if parent domain exists
		parentPath := req.Domain
		if req.Type != "" {
			parentPath += "/" + req.Type
		}
		parent, _ := s.repo.GetDomainByPath(ctx, parentPath)
		if parent == nil {
			return nil, ErrInvalidDomain
		}
	}

	// Validate schemas are valid JSON
	if !json.Valid(req.InputSchema) {
		return nil, fmt.Errorf("%w: input_schema is not valid JSON", ErrInvalidSchema)
	}
	if !json.Valid(req.OutputSchema) {
		return nil, fmt.Errorf("%w: output_schema is not valid JSON", ErrInvalidSchema)
	}

	// Build capability
	cap := &Capability{
		AgentID:          agentID,
		Domain:           req.Domain,
		Type:             req.Type,
		Subtype:          req.Subtype,
		Name:             req.Name,
		Description:      req.Description,
		InputSchema:      req.InputSchema,
		OutputSchema:     req.OutputSchema,
		IsActive:         true,
		IsAcceptingTasks: true,
	}

	// Status events
	if len(req.StatusEvents) > 0 {
		eventsJSON, _ := json.Marshal(req.StatusEvents)
		cap.StatusEvents = eventsJSON
	}

	// Geographic constraints
	if req.Geographic != nil {
		switch req.Geographic.Type {
		case "radius":
			cap.GeographicScope = ScopeLocal
			if req.Geographic.Center != nil {
				cap.GeoCenterLat = &req.Geographic.Center.Lat
				cap.GeoCenterLng = &req.Geographic.Center.Lng
			}
			if req.Geographic.RadiusKM > 0 {
				cap.GeoRadiusKM = &req.Geographic.RadiusKM
			}
		case "polygon":
			cap.GeographicScope = ScopeRegional
			if len(req.Geographic.Polygon) > 0 {
				polygonJSON, _ := json.Marshal(req.Geographic.Polygon)
				cap.GeoPolygon = polygonJSON
			}
		default:
			cap.GeographicScope = ScopeInternational
		}
	}

	// Temporal constraints
	if req.Temporal != nil {
		cap.AvailableHours = req.Temporal.AvailableHours
		cap.AvailableDays = req.Temporal.AvailableDays
		if req.Temporal.Timezone != "" {
			cap.Timezone = req.Temporal.Timezone
		}
	}

	// Pricing
	if req.Pricing != nil {
		cap.PricingModel = req.Pricing.Model
		if req.Pricing.BaseFee > 0 {
			cap.BaseFee = &req.Pricing.BaseFee
		}
		if req.Pricing.PercentageFee > 0 {
			cap.PercentageFee = &req.Pricing.PercentageFee
		}
		cap.Currency = req.Pricing.Currency
		if cap.Currency == "" {
			cap.Currency = "USD"
		}
		if len(req.Pricing.Tiers) > 0 {
			detailsJSON, _ := json.Marshal(req.Pricing)
			cap.PricingDetails = detailsJSON
		}
	}

	// SLA
	if req.SLA != nil {
		cap.ResponseTimeSeconds = &req.SLA.ResponseTimeSeconds
		cap.CompletionTimeP50 = req.SLA.CompletionTimeP50
		cap.CompletionTimeP95 = req.SLA.CompletionTimeP95
	}

	// Create in database
	if err := s.repo.Create(ctx, cap); err != nil {
		return nil, fmt.Errorf("failed to create capability: %w", err)
	}

	// Create initial unverified verification record
	v := &Verification{
		CapabilityID: cap.ID,
		Level:        VerificationUnverified,
		VerifiedAt:   time.Now(),
		VerifiedBy:   "system",
	}
	_ = s.repo.CreateVerification(ctx, v)
	cap.Verification = v

	return cap, nil
}

// GetByID retrieves a capability by ID.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*Capability, error) {
	cap, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get capability: %w", err)
	}
	if cap == nil {
		return nil, ErrCapabilityNotFound
	}
	return cap, nil
}

// GetByAgentID retrieves all capabilities for an agent.
func (s *Service) GetByAgentID(ctx context.Context, agentID uuid.UUID) ([]*Capability, error) {
	caps, err := s.repo.GetByAgentID(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get capabilities: %w", err)
	}
	return caps, nil
}

// Update updates a capability.
func (s *Service) Update(ctx context.Context, agentID, capabilityID uuid.UUID, req *UpdateCapabilityRequest) (*Capability, error) {
	// Get existing capability
	cap, err := s.repo.GetByID(ctx, capabilityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get capability: %w", err)
	}
	if cap == nil {
		return nil, ErrCapabilityNotFound
	}

	// Verify ownership
	if cap.AgentID != agentID {
		return nil, ErrUnauthorized
	}

	// Apply updates
	if req.Name != nil {
		cap.Name = *req.Name
	}
	if req.Description != nil {
		cap.Description = *req.Description
	}
	if req.InputSchema != nil {
		if !json.Valid(req.InputSchema) {
			return nil, fmt.Errorf("%w: input_schema is not valid JSON", ErrInvalidSchema)
		}
		cap.InputSchema = req.InputSchema
	}
	if req.OutputSchema != nil {
		if !json.Valid(req.OutputSchema) {
			return nil, fmt.Errorf("%w: output_schema is not valid JSON", ErrInvalidSchema)
		}
		cap.OutputSchema = req.OutputSchema
	}
	if len(req.StatusEvents) > 0 {
		eventsJSON, _ := json.Marshal(req.StatusEvents)
		cap.StatusEvents = eventsJSON
	}
	if req.Geographic != nil {
		switch req.Geographic.Type {
		case "radius":
			cap.GeographicScope = ScopeLocal
			if req.Geographic.Center != nil {
				cap.GeoCenterLat = &req.Geographic.Center.Lat
				cap.GeoCenterLng = &req.Geographic.Center.Lng
			}
			if req.Geographic.RadiusKM > 0 {
				cap.GeoRadiusKM = &req.Geographic.RadiusKM
			}
		default:
			cap.GeographicScope = ScopeInternational
		}
	}
	if req.Temporal != nil {
		cap.AvailableHours = req.Temporal.AvailableHours
		cap.AvailableDays = req.Temporal.AvailableDays
		if req.Temporal.Timezone != "" {
			cap.Timezone = req.Temporal.Timezone
		}
	}
	if req.Pricing != nil {
		cap.PricingModel = req.Pricing.Model
		if req.Pricing.BaseFee > 0 {
			cap.BaseFee = &req.Pricing.BaseFee
		}
		if req.Pricing.PercentageFee > 0 {
			cap.PercentageFee = &req.Pricing.PercentageFee
		}
		if req.Pricing.Currency != "" {
			cap.Currency = req.Pricing.Currency
		}
	}
	if req.SLA != nil {
		cap.ResponseTimeSeconds = &req.SLA.ResponseTimeSeconds
		cap.CompletionTimeP50 = req.SLA.CompletionTimeP50
		cap.CompletionTimeP95 = req.SLA.CompletionTimeP95
	}
	if req.IsActive != nil {
		cap.IsActive = *req.IsActive
	}
	if req.IsAcceptingTasks != nil {
		cap.IsAcceptingTasks = *req.IsAcceptingTasks
	}

	// Save
	if err := s.repo.Update(ctx, cap); err != nil {
		return nil, fmt.Errorf("failed to update capability: %w", err)
	}

	return cap, nil
}

// Delete soft-deletes a capability.
func (s *Service) Delete(ctx context.Context, agentID, capabilityID uuid.UUID) error {
	// Get existing capability
	cap, err := s.repo.GetByID(ctx, capabilityID)
	if err != nil {
		return fmt.Errorf("failed to get capability: %w", err)
	}
	if cap == nil {
		return ErrCapabilityNotFound
	}

	// Verify ownership
	if cap.AgentID != agentID {
		return ErrUnauthorized
	}

	return s.repo.Delete(ctx, capabilityID)
}

// Search searches capabilities with filters.
func (s *Service) Search(ctx context.Context, req *SearchCapabilitiesRequest) (*SearchCapabilitiesResponse, error) {
	return s.repo.Search(ctx, req)
}

// GetDomainTaxonomy retrieves the full domain taxonomy.
func (s *Service) GetDomainTaxonomy(ctx context.Context) ([]DomainTaxonomy, error) {
	return s.repo.GetDomainTaxonomy(ctx)
}

// GetDomainTaxonomyTree retrieves the domain taxonomy as a tree structure.
func (s *Service) GetDomainTaxonomyTree(ctx context.Context) ([]DomainTaxonomy, error) {
	flat, err := s.repo.GetDomainTaxonomy(ctx)
	if err != nil {
		return nil, err
	}

	// Build tree from flat list
	lookup := make(map[string]*DomainTaxonomy)
	var roots []DomainTaxonomy

	for i := range flat {
		lookup[flat[i].Path] = &flat[i]
	}

	for i := range flat {
		if flat[i].ParentPath == nil {
			roots = append(roots, flat[i])
		} else if parent, ok := lookup[*flat[i].ParentPath]; ok {
			parent.Children = append(parent.Children, flat[i])
		}
	}

	return roots, nil
}

// --- Verification methods ---

// RequestVerification initiates a verification process for a capability.
func (s *Service) RequestVerification(ctx context.Context, agentID, capabilityID uuid.UUID, method string) (*Verification, error) {
	// Get capability
	cap, err := s.repo.GetByID(ctx, capabilityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get capability: %w", err)
	}
	if cap == nil {
		return nil, ErrCapabilityNotFound
	}

	// Verify ownership
	if cap.AgentID != agentID {
		return nil, ErrUnauthorized
	}

	// For now, just mark as tested (in real impl, would run tests)
	v := &Verification{
		CapabilityID: capabilityID,
		Level:        VerificationTested,
		Method:       method,
		VerifiedAt:   time.Now(),
		VerifiedBy:   "system",
	}

	if err := s.repo.CreateVerification(ctx, v); err != nil {
		return nil, fmt.Errorf("failed to create verification: %w", err)
	}

	return v, nil
}

// GetVerification gets the current verification for a capability.
func (s *Service) GetVerification(ctx context.Context, capabilityID uuid.UUID) (*Verification, error) {
	return s.repo.GetCurrentVerification(ctx, capabilityID)
}

// --- Stats methods ---

// RecordTaskCompletion records a task completion for a capability.
func (s *Service) RecordTaskCompletion(ctx context.Context, capabilityID uuid.UUID, success bool, rating *float64) error {
	if err := s.repo.IncrementTaskStats(ctx, capabilityID, success); err != nil {
		return fmt.Errorf("failed to update stats: %w", err)
	}

	if rating != nil && *rating > 0 {
		if err := s.repo.UpdateAverageRating(ctx, capabilityID, *rating); err != nil {
			return fmt.Errorf("failed to update rating: %w", err)
		}
	}

	return nil
}
