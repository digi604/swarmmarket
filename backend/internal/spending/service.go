package spending

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrSpendingLimitExceeded = errors.New("spending limit exceeded")
	ErrNotAgentOwner         = errors.New("not the owner of this agent")
)

// OwnershipChecker verifies agent ownership.
type OwnershipChecker interface {
	IsAgentOwner(ctx context.Context, userID, agentID uuid.UUID) (bool, error)
}

type Service struct {
	repo             *Repository
	ownershipChecker OwnershipChecker
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) SetOwnershipChecker(oc OwnershipChecker) {
	s.ownershipChecker = oc
}

// CheckSpendingLimit checks if the agent can spend the given amount.
func (s *Service) CheckSpendingLimit(ctx context.Context, agentID uuid.UUID, amount float64) error {
	sl, err := s.repo.GetByAgentID(ctx, agentID)
	if err != nil {
		return err
	}
	// No limits configured = allowed
	if sl == nil || !sl.IsEnabled {
		return nil
	}

	// Per-transaction limit
	if sl.MaxPerTransaction != nil && amount > *sl.MaxPerTransaction {
		return fmt.Errorf("%w: amount %.2f exceeds per-transaction limit %.2f", ErrSpendingLimitExceeded, amount, *sl.MaxPerTransaction)
	}

	// Daily limit
	if sl.DailyLimit != nil {
		startOfDay := time.Now().UTC().Truncate(24 * time.Hour)
		spent, err := s.repo.GetAgentSpendSince(ctx, agentID, startOfDay)
		if err != nil {
			return err
		}
		if spent+amount > *sl.DailyLimit {
			return fmt.Errorf("%w: daily spend %.2f + %.2f would exceed limit %.2f", ErrSpendingLimitExceeded, spent, amount, *sl.DailyLimit)
		}
	}

	// Monthly limit
	if sl.MonthlyLimit != nil {
		now := time.Now().UTC()
		startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		spent, err := s.repo.GetAgentSpendSince(ctx, agentID, startOfMonth)
		if err != nil {
			return err
		}
		if spent+amount > *sl.MonthlyLimit {
			return fmt.Errorf("%w: monthly spend %.2f + %.2f would exceed limit %.2f", ErrSpendingLimitExceeded, spent, amount, *sl.MonthlyLimit)
		}
	}

	return nil
}

func (s *Service) GetLimits(ctx context.Context, agentID uuid.UUID) (*SpendingLimit, error) {
	return s.repo.GetByAgentID(ctx, agentID)
}

func (s *Service) SetLimits(ctx context.Context, ownerUserID, agentID uuid.UUID, req *SetSpendingLimitRequest) error {
	// Verify ownership
	if s.ownershipChecker != nil {
		isOwner, err := s.ownershipChecker.IsAgentOwner(ctx, ownerUserID, agentID)
		if err != nil {
			return err
		}
		if !isOwner {
			return ErrNotAgentOwner
		}
	}

	isEnabled := true
	if req.IsEnabled != nil {
		isEnabled = *req.IsEnabled
	}

	sl := &SpendingLimit{
		AgentID:           agentID,
		OwnerUserID:       ownerUserID,
		MaxPerTransaction: req.MaxPerTransaction,
		DailyLimit:        req.DailyLimit,
		MonthlyLimit:      req.MonthlyLimit,
		IsEnabled:         isEnabled,
	}

	return s.repo.Upsert(ctx, sl)
}
