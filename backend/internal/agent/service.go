package agent

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Service handles agent business logic.
type Service struct {
	repo      *Repository
	keyLength int
}

// NewService creates a new agent service.
func NewService(repo *Repository, keyLength int) *Service {
	if keyLength <= 0 {
		keyLength = 32
	}
	return &Service{
		repo:      repo,
		keyLength: keyLength,
	}
}

// Register creates a new agent with an API key.
func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	// Generate API key
	apiKey, err := s.generateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate api key: %w", err)
	}

	// Hash the API key for storage
	keyHash := s.hashAPIKey(apiKey)
	keyPrefix := apiKey[:8] // Store prefix for identification

	now := time.Now().UTC()
	agent := &Agent{
		ID:                uuid.New(),
		Name:              req.Name,
		Description:       req.Description,
		OwnerEmail:        req.OwnerEmail,
		APIKeyHash:        keyHash,
		APIKeyPrefix:      keyPrefix,
		VerificationLevel: VerificationBasic,
		TrustScore:        0.5, // Start at neutral
		TotalTransactions: 0,
		SuccessfulTrades:  0,
		AverageRating:     0,
		IsActive:          true,
		Metadata:          req.Metadata,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := s.repo.Create(ctx, agent); err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	return &RegisterResponse{
		Agent:  agent,
		APIKey: apiKey,
	}, nil
}

// GetByID retrieves an agent by ID.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*Agent, error) {
	return s.repo.GetByID(ctx, id)
}

// GetPublicProfile retrieves an agent's public profile.
func (s *Service) GetPublicProfile(ctx context.Context, id uuid.UUID) (*AgentPublicProfile, error) {
	agent, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return agent.PublicProfile(), nil
}

// ValidateAPIKey validates an API key and returns the associated agent.
func (s *Service) ValidateAPIKey(ctx context.Context, apiKey string) (*Agent, error) {
	hash := s.hashAPIKey(apiKey)
	agent, err := s.repo.GetByAPIKeyHash(ctx, hash)
	if err != nil {
		return nil, err
	}

	// Update last seen asynchronously
	go func() {
		ctx := context.Background()
		_ = s.repo.UpdateLastSeen(ctx, agent.ID)
	}()

	return agent, nil
}

// Update updates an agent's profile.
func (s *Service) Update(ctx context.Context, id uuid.UUID, req *UpdateRequest) (*Agent, error) {
	agent, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		agent.Name = *req.Name
	}
	if req.Description != nil {
		agent.Description = *req.Description
	}
	if req.Metadata != nil {
		if agent.Metadata == nil {
			agent.Metadata = make(map[string]any)
		}
		for k, v := range req.Metadata {
			agent.Metadata[k] = v
		}
	}

	if err := s.repo.Update(ctx, agent); err != nil {
		return nil, err
	}

	return agent, nil
}

// Deactivate deactivates an agent.
func (s *Service) Deactivate(ctx context.Context, id uuid.UUID) error {
	return s.repo.Deactivate(ctx, id)
}

// GetReputation retrieves the reputation for an agent.
func (s *Service) GetReputation(ctx context.Context, id uuid.UUID) (*Reputation, error) {
	return s.repo.GetReputation(ctx, id)
}

// generateAPIKey generates a cryptographically secure API key.
func (s *Service) generateAPIKey() (string, error) {
	bytes := make([]byte, s.keyLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "sm_" + hex.EncodeToString(bytes), nil
}

// hashAPIKey creates a SHA-256 hash of the API key.
func (s *Service) hashAPIKey(apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(hash[:])
}
