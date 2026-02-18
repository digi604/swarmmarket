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

// SpendingChecker checks spending limits for an agent.
type SpendingChecker interface {
	CheckSpendingLimit(ctx context.Context, agentID uuid.UUID, amount float64) error
}

// PaymentCreator interface for creating escrow payments.
type PaymentCreator interface {
	CreateEscrowPayment(ctx context.Context, transactionID, buyerID, sellerID string, amount float64, currency string) (paymentIntentID string, err error)
}

// ListingTransactionCreator interface for creating transactions from listing purchases.
type ListingTransactionCreator interface {
	CreateFromListing(ctx context.Context, buyerID, sellerID uuid.UUID, listingID *uuid.UUID, amount float64, currency string) (uuid.UUID, error)
}

// Service handles marketplace business logic.
type Service struct {
	repo                      RepositoryInterface
	publisher                 EventPublisher
	transactionCreator        TransactionCreator
	spendingChecker           SpendingChecker
	paymentCreator            PaymentCreator
	listingTransactionCreator ListingTransactionCreator
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

// SetSpendingChecker sets the spending checker (to avoid circular dependency).
func (s *Service) SetSpendingChecker(sc SpendingChecker) {
	s.spendingChecker = sc
}

// SetPaymentCreator sets the payment creator (to avoid circular dependency).
func (s *Service) SetPaymentCreator(pc PaymentCreator) {
	s.paymentCreator = pc
}

// SetListingTransactionCreator sets the listing transaction creator (to avoid circular dependency).
func (s *Service) SetListingTransactionCreator(ltc ListingTransactionCreator) {
	s.listingTransactionCreator = ltc
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

// PurchaseListing handles direct purchase of a listing.
func (s *Service) PurchaseListing(ctx context.Context, buyerID uuid.UUID, listingID uuid.UUID, quantity int) (*PurchaseResult, error) {
	// Default quantity to 1
	if quantity <= 0 {
		quantity = 1
	}

	// Get the listing
	listing, err := s.repo.GetListingByID(ctx, listingID)
	if err != nil {
		return nil, err
	}

	// Validate listing is active
	if listing.Status != ListingStatusActive {
		return nil, fmt.Errorf("listing is not available for purchase")
	}

	// Validate buyer is not the seller
	if listing.SellerID == buyerID {
		return nil, fmt.Errorf("cannot purchase your own listing")
	}

	// Validate quantity is available
	if quantity > listing.Quantity {
		return nil, fmt.Errorf("requested quantity (%d) exceeds available quantity (%d)", quantity, listing.Quantity)
	}

	// Validate price is set
	if listing.PriceAmount == nil || *listing.PriceAmount <= 0 {
		return nil, fmt.Errorf("listing price is not set")
	}

	// Calculate total amount
	totalAmount := *listing.PriceAmount * float64(quantity)
	currency := listing.PriceCurrency
	if currency == "" {
		currency = "USD"
	}

	// Check if transaction creator is configured
	if s.listingTransactionCreator == nil {
		return nil, fmt.Errorf("transaction service not configured")
	}

	// Create transaction
	transactionID, err := s.listingTransactionCreator.CreateFromListing(
		ctx,
		buyerID,
		listing.SellerID,
		&listingID,
		totalAmount,
		currency,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Auto-fund escrow via off-session payment
	var paymentIntentID string
	if s.paymentCreator != nil {
		paymentIntentID, err = s.paymentCreator.CreateEscrowPayment(
			ctx,
			transactionID.String(),
			buyerID.String(),
			listing.SellerID.String(),
			totalAmount,
			currency,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create payment: %w", err)
		}
	}

	s.publishEvent(ctx, "listing.purchased", map[string]any{
		"listing_id":     listingID,
		"transaction_id": transactionID,
		"buyer_id":       buyerID,
		"seller_id":      listing.SellerID,
		"quantity":       quantity,
		"amount":         totalAmount,
		"currency":       currency,
	})

	return &PurchaseResult{
		TransactionID:   transactionID,
		PaymentIntentID: paymentIntentID,
		Amount:          totalAmount,
		Currency:        currency,
		Status:          "pending",
	}, nil
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

// UpdateRequest updates a request (only by the requester, only if still open).
func (s *Service) UpdateRequest(ctx context.Context, requesterID uuid.UUID, requestID uuid.UUID, req *UpdateRequestRequest) (*Request, error) {
	// Get existing request to verify ownership and status
	existing, err := s.repo.GetRequestByID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	if existing.RequesterID != requesterID {
		return nil, fmt.Errorf("not authorized to update this request")
	}

	if existing.Status != RequestStatusOpen {
		return nil, fmt.Errorf("can only update open requests")
	}

	updated, err := s.repo.UpdateRequest(ctx, requestID, requesterID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update request: %w", err)
	}

	// Publish event
	s.publishEvent(ctx, "request.updated", map[string]any{
		"request_id":   updated.ID,
		"requester_id": updated.RequesterID,
		"title":        updated.Title,
	})

	return updated, nil
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

	// Check spending limits
	if s.spendingChecker != nil {
		if err := s.spendingChecker.CheckSpendingLimit(ctx, requesterID, offer.PriceAmount); err != nil {
			return nil, fmt.Errorf("spending limit check failed: %w", err)
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
		"offer_id":       offer.ID,
		"request_id":     offer.RequestID,
		"offerer_id":     offer.OffererID,
		"requester_id":   request.RequesterID,
		"price":          offer.PriceAmount,
		"price_amount":   offer.PriceAmount,
		"price_currency": offer.PriceCurrency,
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
		ListingID: &listingID,
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

// --- Request Comments ---

// CreateRequestComment creates a new comment on a request.
func (s *Service) CreateRequestComment(ctx context.Context, agentID, requestID uuid.UUID, req *CreateCommentRequest) (*Comment, error) {
	if req.Content == "" {
		return nil, fmt.Errorf("content is required")
	}

	// Verify request exists
	request, err := s.repo.GetRequestByID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	comment := &Comment{
		ID:        uuid.New(),
		RequestID: &requestID,
		AgentID:   agentID,
		ParentID:  req.ParentID,
		Content:   req.Content,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.CreateRequestComment(ctx, comment); err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	// Notify request owner of new comment
	s.publishEvent(ctx, "request_comment.created", map[string]any{
		"comment_id":   comment.ID,
		"request_id":   requestID,
		"requester_id": request.RequesterID,
		"commenter_id": agentID,
	})

	return comment, nil
}

// GetRequestComments retrieves comments for a request.
func (s *Service) GetRequestComments(ctx context.Context, requestID uuid.UUID, limit, offset int) (*CommentsResponse, error) {
	comments, total, err := s.repo.GetCommentsByRequestID(ctx, requestID, limit, offset)
	if err != nil {
		return nil, err
	}
	return &CommentsResponse{Comments: comments, Total: total}, nil
}

// GetRequestCommentReplies retrieves replies to a request comment.
func (s *Service) GetRequestCommentReplies(ctx context.Context, parentID uuid.UUID) ([]Comment, error) {
	return s.repo.GetRequestCommentReplies(ctx, parentID)
}

// DeleteRequestComment deletes a request comment.
func (s *Service) DeleteRequestComment(ctx context.Context, commentID, agentID uuid.UUID) error {
	return s.repo.DeleteRequestComment(ctx, commentID, agentID)
}

// --- Ownership Checks ---

// IsListingOwner checks if an agent owns a listing.
func (s *Service) IsListingOwner(ctx context.Context, listingID, agentID uuid.UUID) (bool, error) {
	listing, err := s.repo.GetListingByID(ctx, listingID)
	if err != nil {
		return false, err
	}
	return listing.SellerID == agentID, nil
}

// IsRequestOwner checks if an agent owns a request.
func (s *Service) IsRequestOwner(ctx context.Context, requestID, agentID uuid.UUID) (bool, error) {
	request, err := s.repo.GetRequestByID(ctx, requestID)
	if err != nil {
		return false, err
	}
	return request.RequesterID == agentID, nil
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
