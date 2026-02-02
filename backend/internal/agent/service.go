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
	repo      RepositoryInterface
	keyLength int
}

// NewService creates a new agent service.
func NewService(repo RepositoryInterface, keyLength int) *Service {
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
	var avatarURL *string
	if req.AvatarURL != "" {
		avatarURL = &req.AvatarURL
	}
	agent := &Agent{
		ID:                uuid.New(),
		Name:              req.Name,
		Description:       req.Description,
		AvatarURL:         avatarURL,
		OwnerEmail:        req.OwnerEmail,
		APIKeyHash:        keyHash,
		APIKeyPrefix:      keyPrefix,
		VerificationLevel: VerificationBasic,
		TrustScore:        0.5, // Start at neutral
		TotalTransactions: 0,
		SuccessfulTrades:  0,
		AverageRating:     0.5, // Start at neutral
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

	profile := agent.PublicProfile()

	// Count active listings
	listingCount, err := s.repo.CountActiveListings(ctx, id)
	if err == nil {
		profile.ActiveListings = listingCount
	}

	return profile, nil
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
	if req.AvatarURL != nil {
		agent.AvatarURL = req.AvatarURL
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

// GenerateOwnershipToken generates a token for claiming agent ownership.
// The token expires after 24 hours.
func (s *Service) GenerateOwnershipToken(ctx context.Context, agentID uuid.UUID) (string, time.Time, error) {
	// Verify agent exists
	_, err := s.repo.GetByID(ctx, agentID)
	if err != nil {
		return "", time.Time{}, err
	}

	// Generate token
	token, err := s.generateOwnershipToken()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to generate ownership token: %w", err)
	}

	// Hash and store
	tokenHash := s.hashToken(token)
	expiresAt := time.Now().UTC().Add(24 * time.Hour)

	if err := s.repo.CreateOwnershipToken(ctx, agentID, tokenHash, expiresAt); err != nil {
		return "", time.Time{}, err
	}

	return token, expiresAt, nil
}

// ClaimOwnership claims ownership of an agent using an ownership token.
func (s *Service) ClaimOwnership(ctx context.Context, userID uuid.UUID, token string) (*Agent, error) {
	tokenHash := s.hashToken(token)

	// Get the token
	ownershipToken, err := s.repo.GetOwnershipTokenByHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	// Check if token is expired
	if time.Now().After(ownershipToken.ExpiresAt) {
		return nil, ErrTokenExpired
	}

	// Check if token is already used
	if ownershipToken.UsedAt != nil {
		return nil, ErrTokenAlreadyUsed
	}

	// Check if agent already has an owner
	existingOwner, err := s.repo.GetAgentOwnerID(ctx, ownershipToken.AgentID)
	if err != nil {
		return nil, err
	}
	if existingOwner != nil {
		return nil, ErrAgentAlreadyOwned
	}

	// Mark token as used
	if err := s.repo.MarkTokenUsed(ctx, ownershipToken.ID, userID); err != nil {
		return nil, err
	}

	// Set agent owner
	if err := s.repo.SetAgentOwner(ctx, ownershipToken.AgentID, userID); err != nil {
		return nil, err
	}

	// Return the claimed agent
	return s.repo.GetByID(ctx, ownershipToken.AgentID)
}

// GetAgentsByOwner retrieves all agents owned by a user.
func (s *Service) GetAgentsByOwner(ctx context.Context, userID uuid.UUID) ([]*Agent, error) {
	return s.repo.GetAgentsByOwner(ctx, userID)
}

// generateOwnershipToken generates a cryptographically secure ownership token.
func (s *Service) generateOwnershipToken() (string, error) {
	bytes := make([]byte, 24) // 24 bytes = 48 hex chars
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "own_" + hex.EncodeToString(bytes), nil
}

// hashToken creates a SHA-256 hash of a token.
func (s *Service) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
