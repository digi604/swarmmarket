package agent

import (
	"strings"
	"testing"
)

func TestGenerateAPIKey(t *testing.T) {
	s := NewService(nil, 32)

	key, err := s.generateAPIKey()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check prefix
	if !strings.HasPrefix(key, "sm_") {
		t.Errorf("expected key to have prefix 'sm_', got %s", key)
	}

	// Check length: "sm_" (3) + 64 hex chars (32 bytes * 2)
	expectedLen := 3 + 64
	if len(key) != expectedLen {
		t.Errorf("expected key length %d, got %d", expectedLen, len(key))
	}
}

func TestHashAPIKey(t *testing.T) {
	s := NewService(nil, 32)

	key := "sm_test_key_12345"
	hash1 := s.hashAPIKey(key)
	hash2 := s.hashAPIKey(key)

	// Same input should produce same hash
	if hash1 != hash2 {
		t.Error("expected same hash for same input")
	}

	// Different input should produce different hash
	hash3 := s.hashAPIKey("sm_different_key")
	if hash1 == hash3 {
		t.Error("expected different hash for different input")
	}

	// Hash should be 64 chars (SHA-256 = 32 bytes = 64 hex chars)
	if len(hash1) != 64 {
		t.Errorf("expected hash length 64, got %d", len(hash1))
	}
}

func TestNewServiceDefaultKeyLength(t *testing.T) {
	s := NewService(nil, 0)

	if s.keyLength != 32 {
		t.Errorf("expected default key length 32, got %d", s.keyLength)
	}
}

func TestAgentPublicProfile(t *testing.T) {
	agent := &Agent{
		Name:              "Test Agent",
		Description:       "A test agent",
		OwnerEmail:        "secret@example.com",
		APIKeyHash:        "secrethash",
		VerificationLevel: VerificationBasic,
		TrustScore:        0.8,
		TotalTransactions: 10,
		SuccessfulTrades:  9,
		AverageRating:     4.5,
	}

	profile := agent.PublicProfile()

	// Should include public fields
	if profile.Name != agent.Name {
		t.Errorf("expected name %s, got %s", agent.Name, profile.Name)
	}
	if profile.TrustScore != agent.TrustScore {
		t.Errorf("expected trust score %f, got %f", agent.TrustScore, profile.TrustScore)
	}

	// Should not include sensitive fields (checked by type system)
	// OwnerEmail and APIKeyHash are not in AgentPublicProfile
}
