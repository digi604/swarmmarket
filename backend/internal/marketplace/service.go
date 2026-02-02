package marketplace

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// EventPublisher interface for publishing events.
type EventPublisher interface {
	Publish(ctx context.Context, eventType string, payload map[string]any) error
}

// TransactionCreator interface for creating transactions.
type TransactionCreator interface {
	CreateFromOffer(ctx context.Context, buyerID, sellerID uuid.UUID, requestID, offerID *uuid.UUID, amount float64, currency string) (uuid.UUID, error)
}

// WalletChecker interface for checking wallet balance.
type WalletChecker interface {
	GetAgentWalletBalance(ctx context.Context, agentID uuid.UUID) (available float64, err error)
}

// PaymentCreator interface for creating payment intents.
type PaymentCreator interface {
	CreateEscrowPayment(ctx context.Context, transactionID, buyerID, sellerID string, amount float64, currency string) (paymentIntentID, clientSecret string, err error)
}

// ListingTransactionCreator interface for creating transactions from listing purchases.
type ListingTransactionCreator interface {
	CreateFromListing(ctx context.Context, buyerID, sellerID uuid.UUID, listingID *uuid.UUID, amount float64, currency string) (uuid.UUID, error)
}

// Service handles marketplace business logic.
type Service struct {
	repo               RepositoryInterface
	publisher          EventPublisher
	transactionCreator TransactionCreator
	walletChecker      WalletChecker
}

// NewService creates a new marketplace service.
func NewService(repo RepositoryInterface, publisher EventPublisher) *Service {
	return &Service{
		repo:      repo,
		publisher: publisher,
	}
}

// SetTransactionCreator sets the transaction creator (to avoid circular dependency).
func (s *Service) SetTransactionCreator(tc TransactionCreator) {
	s.transactionCreator = tc
}

// SetWalletChecker sets the wallet checker (to avoid circular dependency).
func (s *Service) SetWalletChecker(wc WalletChecker) {
	s.walletChecker = wc
}

// --- Listings ---

// CreateListing creates a new listing.
func (s *Service) CreateListing(ctx context.Context, sellerID uuid.UUID, req *CreateListingRequest) (*Listing, error) {
	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if req.ListingType == "" {
		return nil, fmt.Errorf("listing_type is required")
	}

	now := time.Now().UTC()
	listing := &Listing{
		ID:              uuid.New(),
		SellerID:        sellerID,
		CategoryID:      req.CategoryID,
		Title:           req.Title,
		Description:     req.Description,
		ListingType:     req.ListingType,
		PriceAmount:     req.PriceAmount,
		PriceCurrency:   defaultString(req.PriceCurrency, "USD"),
		Quantity:        defaultInt(req.Quantity, 1),
		GeographicScope: defaultScope(req.GeographicScope, ScopeInternational),
		LocationLat:     req.LocationLat,
		LocationLng:     req.LocationLng,
		LocationRadius:  req.LocationRadius,
		Status:          ListingStatusActive,
		ExpiresAt:       req.ExpiresAt,
		Metadata:        req.Metadata,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.repo.CreateListing(ctx, listing); err != nil {
		return nil, fmt.Errorf("failed to create listing: %w", err)
	}

	// Publish event for interested agents
	// This triggers notifications to agents who have matching search criteria
	s.publishEvent(ctx, "listing.created", map[string]any{
		"listing_id":       listing.ID,
		"seller_id":        listing.SellerID,
		"title":            listing.Title,
		"listing_type":     listing.ListingType,
		"geographic_scope": listing.GeographicScope,
		"category_id":      listing.CategoryID,
	})

	return listing, nil
}

// GetListing retrieves a listing by ID.
func (s *Service) GetListing(ctx context.Context, id uuid.UUID) (*Listing, error) {
	return s.repo.GetListingByID(ctx, id)
}

// GetListingBySlug retrieves a listing by slug.
func (s *Service) GetListingBySlug(ctx context.Context, slug string) (*Listing, error) {
	return s.repo.GetListingBySlug(ctx, slug)
}

// SearchListings searches for listings.
func (s *Service) SearchListings(ctx context.Context, params SearchListingsParams) (*ListResult[Listing], error) {
	return s.repo.SearchListings(ctx, params)
}

// DeleteListing deletes a listing (seller only).
func (s *Service) DeleteListing(ctx context.Context, id uuid.UUID, sellerID uuid.UUID) error {
	return s.repo.DeleteListing(ctx, id, sellerID)
}

// --- Requests ---

// CreateRequest creates a new request (what an agent needs).
func (s *Service) CreateRequest(ctx context.Context, requesterID uuid.UUID, req *CreateRequestRequest) (*Request, error) {
	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if req.RequestType == "" {
		return nil, fmt.Errorf("request_type is required")
	}

	now := time.Now().UTC()
	request := &Request{
		ID:              uuid.New(),
		RequesterID:     requesterID,
		CategoryID:      req.CategoryID,
		Title:           req.Title,
		Description:     req.Description,
		RequestType:     req.RequestType,
		BudgetMin:       req.BudgetMin,
		BudgetMax:       req.BudgetMax,
		BudgetCurrency:  defaultString(req.BudgetCurrency, "USD"),
		Quantity:        defaultInt(req.Quantity, 1),
		GeographicScope: defaultScope(req.GeographicScope, ScopeInternational),
		LocationLat:     req.LocationLat,
		LocationLng:     req.LocationLng,
		LocationRadius:  req.LocationRadius,
		Status:          RequestStatusOpen,
		ExpiresAt:       req.ExpiresAt,
		Metadata:        req.Metadata,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.repo.CreateRequest(ctx, request); err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Publish event - this notifies agents who can fulfill this request
	// The notification service will:
	// 1. Find agents with matching listings (e.g., delivery services)
	// 2. Find agents subscribed to this category
	// 3. Send WebSocket events to connected agents
	// 4. Send webhook calls to agents with registered endpoints
	s.publishEvent(ctx, "request.created", map[string]any{
		"request_id":       request.ID,
		"requester_id":     request.RequesterID,
		"title":            request.Title,
		"request_type":     request.RequestType,
		"budget_min":       request.BudgetMin,
		"budget_max":       request.BudgetMax,
		"geographic_scope": request.GeographicScope,
		"category_id":      request.CategoryID,
	})

	return request, nil
}

// GetRequest retrieves a request by ID.
func (s *Service) GetRequest(ctx context.Context, id uuid.UUID) (*Request, error) {
	return s.repo.GetRequestByID(ctx, id)
}

// GetRequestBySlug retrieves a request by slug.
func (s *Service) GetRequestBySlug(ctx context.Context, slug string) (*Request, error) {
	return s.repo.GetRequestBySlug(ctx, slug)
}

// SearchRequests searches for open requests.
func (s *Service) SearchRequests(ctx context.Context, params SearchRequestsParams) (*ListResult[Request], error) {
	return s.repo.SearchRequests(ctx, params)
}

// --- Offers ---

// SubmitOffer submits an offer to a request.
func (s *Service) SubmitOffer(ctx context.Context, offererID uuid.UUID, requestID uuid.UUID, req *CreateOfferRequest) (*Offer, error) {
	// Verify request exists and is open
	request, err := s.repo.GetRequestByID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if request.Status != RequestStatusOpen {
		return nil, fmt.Errorf("request is not open for offers")
	}

	// Prevent self-offers
	if request.RequesterID == offererID {
		return nil, fmt.Errorf("cannot submit offer to your own request")
	}

	if req.PriceAmount <= 0 {
		return nil, fmt.Errorf("price_amount must be positive")
	}

	now := time.Now().UTC()
	offer := &Offer{
		ID:            uuid.New(),
		RequestID:     requestID,
		OffererID:     offererID,
		ListingID:     req.ListingID,
		PriceAmount:   req.PriceAmount,
		PriceCurrency: defaultString(req.PriceCurrency, "USD"),
		Description:   req.Description,
		DeliveryTerms: req.DeliveryTerms,
		ValidUntil:    req.ValidUntil,
		Status:        OfferStatusPending,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.repo.CreateOffer(ctx, offer); err != nil {
		return nil, fmt.Errorf("failed to create offer: %w", err)
	}

	// Notify the requester that they received an offer
	// This sends a real-time notification via WebSocket or webhook
	s.publishEvent(ctx, "offer.received", map[string]any{
		"offer_id":      offer.ID,
		"request_id":    offer.RequestID,
		"requester_id":  request.RequesterID, // Who to notify
		"offerer_id":    offer.OffererID,
		"price_amount":  offer.PriceAmount,
		"price_currency": offer.PriceCurrency,
	})

	return offer, nil
}

// GetOffersByRequest retrieves all offers for a request.
func (s *Service) GetOffersByRequest(ctx context.Context, requestID uuid.UUID) ([]Offer, error) {
	return s.repo.GetOffersByRequestID(ctx, requestID)
}

// AcceptOffer accepts an offer, creating a transaction.
func (s *Service) AcceptOffer(ctx context.Context, requesterID uuid.UUID, offerID uuid.UUID) (*Offer, error) {
	offer, err := s.repo.GetOfferByID(ctx, offerID)
	if err != nil {
		return nil, err
	}

	// Verify the requester owns the request
	request, err := s.repo.GetRequestByID(ctx, offer.RequestID)
	if err != nil {
		return nil, err
	}
	if request.RequesterID != requesterID {
		return nil, fmt.Errorf("not authorized to accept this offer")
	}
	if offer.Status != OfferStatusPending {
		return nil, fmt.Errorf("offer is not pending")
	}

	// Check wallet balance if wallet checker is available
	if s.walletChecker != nil {
		balance, err := s.walletChecker.GetAgentWalletBalance(ctx, requesterID)
		if err != nil {
			return nil, fmt.Errorf("failed to check wallet balance: %w", err)
		}
		if balance < offer.PriceAmount {
			return nil, fmt.Errorf("insufficient funds: need %.2f %s, have %.2f", offer.PriceAmount, offer.PriceCurrency, balance)
		}
	}

	// Update offer status
	if err := s.repo.UpdateOfferStatus(ctx, offerID, OfferStatusAccepted); err != nil {
		return nil, err
	}
	offer.Status = OfferStatusAccepted

	// Create transaction if transaction creator is set
	var transactionID *uuid.UUID
	if s.transactionCreator != nil {
		txID, err := s.transactionCreator.CreateFromOffer(
			ctx,
			request.RequesterID, // buyer
			offer.OffererID,     // seller
			&offer.RequestID,
			&offer.ID,
			offer.PriceAmount,
			offer.PriceCurrency,
		)
		if err == nil {
			transactionID = &txID
		}
	}

	// Notify the offerer that their offer was accepted
	eventPayload := map[string]any{
		"offer_id":     offer.ID,
		"request_id":   offer.RequestID,
		"offerer_id":   offer.OffererID,
		"requester_id": request.RequesterID,
	}
	if transactionID != nil {
		eventPayload["transaction_id"] = *transactionID
	}
	s.publishEvent(ctx, "offer.accepted", eventPayload)

	return offer, nil
}

// --- Categories ---

// GetCategories retrieves all categories.
func (s *Service) GetCategories(ctx context.Context) ([]Category, error) {
	return s.repo.GetCategories(ctx)
}

// --- Comments ---

// CreateComment creates a new comment on a listing.
func (s *Service) CreateComment(ctx context.Context, agentID, listingID uuid.UUID, req *CreateCommentRequest) (*Comment, error) {
	if req.Content == "" {
		return nil, fmt.Errorf("content is required")
	}

	// Verify listing exists
	listing, err := s.repo.GetListingByID(ctx, listingID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	comment := &Comment{
		ID:        uuid.New(),
		ListingID: listingID,
		AgentID:   agentID,
		ParentID:  req.ParentID,
		Content:   req.Content,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.CreateComment(ctx, comment); err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	// Notify listing seller of new comment
	s.publishEvent(ctx, "comment.created", map[string]any{
		"comment_id": comment.ID,
		"listing_id": listingID,
		"seller_id":  listing.SellerID,
		"agent_id":   agentID,
	})

	return comment, nil
}

// GetListingComments retrieves comments for a listing.
func (s *Service) GetListingComments(ctx context.Context, listingID uuid.UUID, limit, offset int) (*CommentsResponse, error) {
	comments, total, err := s.repo.GetCommentsByListingID(ctx, listingID, limit, offset)
	if err != nil {
		return nil, err
	}
	return &CommentsResponse{
		Comments: comments,
		Total:    total,
	}, nil
}

// GetCommentReplies retrieves replies to a comment.
func (s *Service) GetCommentReplies(ctx context.Context, parentID uuid.UUID) ([]Comment, error) {
	return s.repo.GetCommentReplies(ctx, parentID)
}

// DeleteComment deletes a comment.
func (s *Service) DeleteComment(ctx context.Context, commentID, agentID uuid.UUID) error {
	return s.repo.DeleteComment(ctx, commentID, agentID)
}

// --- Helpers ---

func (s *Service) publishEvent(ctx context.Context, eventType string, payload map[string]any) {
	if s.publisher != nil {
		go func() {
			_ = s.publisher.Publish(context.Background(), eventType, payload)
		}()
	}
}

func defaultString(val, def string) string {
	if val == "" {
		return def
	}
	return val
}

func defaultInt(val, def int) int {
	if val == 0 {
		return def
	}
	return val
}

func defaultScope(val, def GeographicScope) GeographicScope {
	if val == "" {
		return def
	}
	return val
}
