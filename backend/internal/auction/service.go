package auction

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrAuctionNotActive      = errors.New("auction is not active")
	ErrAuctionEnded          = errors.New("auction has ended")
	ErrBidTooLow             = errors.New("bid amount is too low")
	ErrCannotBidOnOwnAuction = errors.New("cannot bid on your own auction")
	ErrNotAuthorized         = errors.New("not authorized to perform this action")
	ErrInvalidAuctionType    = errors.New("invalid auction type")
)

// EventPublisher publishes events to the notification system.
type EventPublisher interface {
	Publish(ctx context.Context, eventType string, payload map[string]any) error
}

// SpendingChecker checks spending limits for an agent.
type SpendingChecker interface {
	CheckSpendingLimit(ctx context.Context, agentID uuid.UUID, amount float64) error
}

// Service handles auction business logic.
type Service struct {
	repo            RepositoryInterface
	publisher       EventPublisher
	spendingChecker SpendingChecker
}

// NewService creates a new auction service.
func NewService(repo RepositoryInterface, publisher EventPublisher) *Service {
	return &Service{
		repo:      repo,
		publisher: publisher,
	}
}

// SetSpendingChecker sets the spending checker (to avoid circular dependency).
func (s *Service) SetSpendingChecker(sc SpendingChecker) {
	s.spendingChecker = sc
}

// CreateAuction creates a new auction.
func (s *Service) CreateAuction(ctx context.Context, sellerID uuid.UUID, req *CreateAuctionRequest) (*Auction, error) {
	// Validate auction type
	auctionType := AuctionType(req.AuctionType)
	switch auctionType {
	case AuctionTypeEnglish, AuctionTypeDutch, AuctionTypeSealed, AuctionTypeContinuous:
		// Valid
	default:
		return nil, ErrInvalidAuctionType
	}

	// Validate starting price
	if req.StartingPrice <= 0 {
		return nil, errors.New("starting price must be positive")
	}

	// Validate end time
	if req.EndsAt.Before(time.Now().UTC()) {
		return nil, errors.New("end time must be in the future")
	}

	now := time.Now().UTC()
	startsAt := now
	if req.StartsAt != nil && req.StartsAt.After(now) {
		startsAt = *req.StartsAt
	}

	status := AuctionStatusActive
	if startsAt.After(now) {
		status = AuctionStatusScheduled
	}

	extensionSeconds := 60
	if req.ExtensionSeconds != nil {
		extensionSeconds = *req.ExtensionSeconds
	}

	currency := "USD"
	if req.Currency != "" {
		currency = req.Currency
	}

	auction := &Auction{
		ID:                    uuid.New(),
		ListingID:             req.ListingID,
		SellerID:              sellerID,
		AuctionType:           auctionType,
		Title:                 req.Title,
		Description:           req.Description,
		StartingPrice:         req.StartingPrice,
		CurrentPrice:          &req.StartingPrice,
		ReservePrice:          req.ReservePrice,
		BuyNowPrice:           req.BuyNowPrice,
		Currency:              currency,
		MinIncrement:          req.MinIncrement,
		PriceDecrement:        req.PriceDecrement,
		DecrementIntervalSecs: req.DecrementIntervalSecs,
		Status:                status,
		StartsAt:              startsAt,
		EndsAt:                req.EndsAt,
		ExtensionSeconds:      extensionSeconds,
		Metadata:              make(map[string]any),
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	if err := s.repo.CreateAuction(ctx, auction); err != nil {
		return nil, err
	}

	if auction.Status == AuctionStatusActive {
		// Publish event for auctions that start immediately
		s.publishEvent(ctx, "auction.started", map[string]any{
			"auction_id":     auction.ID,
			"seller_id":      auction.SellerID,
			"auction_type":   auction.AuctionType,
			"title":          auction.Title,
			"starting_price": auction.StartingPrice,
			"ends_at":        auction.EndsAt,
		})
	}

	return auction, nil
}

// GetAuction retrieves an auction by ID.
func (s *Service) GetAuction(ctx context.Context, id uuid.UUID) (*Auction, error) {
	return s.repo.GetAuctionByID(ctx, id)
}

// GetAuctionBySlug retrieves an auction by slug.
func (s *Service) GetAuctionBySlug(ctx context.Context, slug string) (*Auction, error) {
	return s.repo.GetAuctionBySlug(ctx, slug)
}

// SearchAuctions searches for auctions.
func (s *Service) SearchAuctions(ctx context.Context, params SearchAuctionsParams) (*AuctionListResult, error) {
	return s.repo.SearchAuctions(ctx, params)
}

// PlaceBid places a bid on an auction.
func (s *Service) PlaceBid(ctx context.Context, auctionID, bidderID uuid.UUID, req *PlaceBidRequest) (*Bid, error) {
	// Get auction
	auction, err := s.repo.GetAuctionByID(ctx, auctionID)
	if err != nil {
		return nil, err
	}

	// Validate auction status
	if auction.Status != AuctionStatusActive {
		return nil, ErrAuctionNotActive
	}

	// Check auction hasn't ended
	if time.Now().UTC().After(auction.EndsAt) {
		return nil, ErrAuctionEnded
	}

	// Cannot bid on own auction
	if auction.SellerID == bidderID {
		return nil, ErrCannotBidOnOwnAuction
	}

	// Check spending limits
	if s.spendingChecker != nil {
		if err := s.spendingChecker.CheckSpendingLimit(ctx, bidderID, req.Amount); err != nil {
			return nil, fmt.Errorf("spending limit check failed: %w", err)
		}
	}

	// Get current highest bid
	highestBid, err := s.repo.GetHighestBid(ctx, auctionID)
	if err != nil {
		return nil, err
	}

	// Validate bid amount based on auction type
	switch auction.AuctionType {
	case AuctionTypeEnglish:
		minBid := auction.StartingPrice
		if highestBid != nil {
			minBid = highestBid.Amount
			if auction.MinIncrement != nil {
				minBid += *auction.MinIncrement
			} else {
				minBid += 1 // Default increment
			}
		}
		if req.Amount < minBid {
			return nil, ErrBidTooLow
		}

	case AuctionTypeDutch:
		// Dutch auction: first bidder wins at current price
		currentPrice := auction.StartingPrice
		if auction.CurrentPrice != nil {
			currentPrice = *auction.CurrentPrice
		}
		if req.Amount < currentPrice {
			return nil, ErrBidTooLow
		}

	case AuctionTypeSealed:
		// Sealed bid: just must be positive
		if req.Amount <= 0 {
			return nil, ErrBidTooLow
		}

	case AuctionTypeContinuous:
		// Continuous: any amount above current
		if highestBid != nil && req.Amount <= highestBid.Amount {
			return nil, ErrBidTooLow
		}
	}

	now := time.Now().UTC()
	bid := &Bid{
		ID:        uuid.New(),
		AuctionID: auctionID,
		BidderID:  bidderID,
		Amount:    req.Amount,
		Currency:  auction.Currency,
		IsSealed:  auction.AuctionType == AuctionTypeSealed,
		Status:    BidStatusActive,
		Metadata:  make(map[string]any),
		CreatedAt: now,
	}

	if err := s.repo.CreateBid(ctx, bid); err != nil {
		return nil, err
	}

	// Handle auction type specific logic
	switch auction.AuctionType {
	case AuctionTypeEnglish:
		// Mark previous bids as outbid
		if highestBid != nil {
			s.repo.MarkPreviousBidsOutbid(ctx, auctionID, bid.ID)

			// Notify outbid bidder
			s.publishEvent(ctx, "bid.outbid", map[string]any{
				"auction_id": auctionID,
				"bidder_id":  highestBid.BidderID,
				"old_amount": highestBid.Amount,
				"new_amount": bid.Amount,
			})
		}

		// Update current price
		s.repo.UpdateAuctionPrice(ctx, auctionID, bid.Amount)

		// Anti-sniping: extend auction if bid is close to end
		if auction.EndsAt.Sub(now) < time.Duration(auction.ExtensionSeconds)*time.Second {
			newEndTime := now.Add(time.Duration(auction.ExtensionSeconds) * time.Second)
			s.repo.ExtendAuction(ctx, auctionID, newEndTime)
		}

	case AuctionTypeDutch:
		// First valid bid wins immediately
		s.repo.SetAuctionWinner(ctx, auctionID, bid.ID, bidderID)
		s.repo.UpdateBidStatus(ctx, bid.ID, BidStatusWon)

		s.publishEvent(ctx, "auction.ended", map[string]any{
			"auction_id":  auctionID,
			"winner_id":   bidderID,
			"final_price": bid.Amount,
		})

	case AuctionTypeSealed:
		// Just record the bid, winner determined at auction end

	case AuctionTypeContinuous:
		// Similar to English but no end time
		if highestBid != nil {
			s.repo.MarkPreviousBidsOutbid(ctx, auctionID, bid.ID)
		}
		s.repo.UpdateAuctionPrice(ctx, auctionID, bid.Amount)
	}

	// Publish bid event
	s.publishEvent(ctx, "bid.placed", map[string]any{
		"auction_id": auctionID,
		"bidder_id":  bidderID,
		"amount":     bid.Amount,
		"seller_id":  auction.SellerID,
	})

	return bid, nil
}

// GetBids retrieves all bids for an auction.
func (s *Service) GetBids(ctx context.Context, auctionID uuid.UUID, requesterID *uuid.UUID) ([]*Bid, error) {
	auction, err := s.repo.GetAuctionByID(ctx, auctionID)
	if err != nil {
		return nil, err
	}

	bids, err := s.repo.GetBidsByAuctionID(ctx, auctionID)
	if err != nil {
		return nil, err
	}

	// For sealed auctions, hide other bidders' amounts until ended
	if auction.AuctionType == AuctionTypeSealed && auction.Status != AuctionStatusEnded {
		for _, bid := range bids {
			if requesterID == nil || bid.BidderID != *requesterID {
				bid.Amount = 0 // Hide amount
			}
		}
	}

	return bids, nil
}

// EndAuction ends an auction and determines the winner.
func (s *Service) EndAuction(ctx context.Context, auctionID, requesterID uuid.UUID) (*Auction, error) {
	auction, err := s.repo.GetAuctionByID(ctx, auctionID)
	if err != nil {
		return nil, err
	}

	// Only seller can manually end
	if auction.SellerID != requesterID {
		return nil, ErrNotAuthorized
	}

	if auction.Status != AuctionStatusActive {
		return nil, ErrAuctionNotActive
	}

	// Get highest bid
	highestBid, err := s.repo.GetHighestBid(ctx, auctionID)
	if err != nil {
		return nil, err
	}

	if highestBid != nil {
		// Check reserve price
		metReserve := true
		if auction.ReservePrice != nil && highestBid.Amount < *auction.ReservePrice {
			metReserve = false
		}

		if metReserve {
			s.repo.SetAuctionWinner(ctx, auctionID, highestBid.ID, highestBid.BidderID)
			s.repo.UpdateBidStatus(ctx, highestBid.ID, BidStatusWon)

			s.publishEvent(ctx, "auction.ended", map[string]any{
				"auction_id":  auctionID,
				"winner_id":   highestBid.BidderID,
				"final_price": highestBid.Amount,
				"met_reserve": true,
			})
		} else {
			s.repo.UpdateAuctionStatus(ctx, auctionID, AuctionStatusEnded)

			s.publishEvent(ctx, "auction.ended", map[string]any{
				"auction_id":  auctionID,
				"met_reserve": false,
			})
		}
	} else {
		s.repo.UpdateAuctionStatus(ctx, auctionID, AuctionStatusEnded)

		s.publishEvent(ctx, "auction.ended", map[string]any{
			"auction_id": auctionID,
			"no_bids":    true,
		})
	}

	return s.repo.GetAuctionByID(ctx, auctionID)
}

// IsAuctionOwner checks if an agent owns an auction.
func (s *Service) IsAuctionOwner(ctx context.Context, auctionID, agentID uuid.UUID) (bool, error) {
	auction, err := s.repo.GetAuctionByID(ctx, auctionID)
	if err != nil {
		return false, err
	}
	return auction.SellerID == agentID, nil
}

// Helper to publish events asynchronously.
func (s *Service) publishEvent(ctx context.Context, eventType string, payload map[string]any) {
	if s.publisher != nil {
		go s.publisher.Publish(ctx, eventType, payload)
	}
}
