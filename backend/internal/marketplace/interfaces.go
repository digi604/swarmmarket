package marketplace

import (
	"context"

	"github.com/google/uuid"
)

// RepositoryInterface defines the contract for marketplace data persistence.
// This interface enables mock implementations for testing.
type RepositoryInterface interface {
	// Listing Operations
	CreateListing(ctx context.Context, listing *Listing) error
	GetListingByID(ctx context.Context, id uuid.UUID) (*Listing, error)
	GetListingBySlug(ctx context.Context, slug string) (*Listing, error)
	SearchListings(ctx context.Context, params SearchListingsParams) (*ListResult[Listing], error)
	DeleteListing(ctx context.Context, id uuid.UUID, sellerID uuid.UUID) error

	// Request Operations
	CreateRequest(ctx context.Context, req *Request) error
	GetRequestByID(ctx context.Context, id uuid.UUID) (*Request, error)
	GetRequestBySlug(ctx context.Context, slug string) (*Request, error)
	SearchRequests(ctx context.Context, params SearchRequestsParams) (*ListResult[Request], error)

	// Offer Operations
	CreateOffer(ctx context.Context, offer *Offer) error
	GetOfferByID(ctx context.Context, id uuid.UUID) (*Offer, error)
	GetOffersByRequestID(ctx context.Context, requestID uuid.UUID) ([]Offer, error)
	UpdateOfferStatus(ctx context.Context, id uuid.UUID, status OfferStatus) error

	// Category Operations
	GetCategories(ctx context.Context) ([]Category, error)

	// Comment Operations
	CreateComment(ctx context.Context, comment *Comment) error
	GetCommentsByListingID(ctx context.Context, listingID uuid.UUID, limit, offset int) ([]Comment, int, error)
	GetCommentReplies(ctx context.Context, parentID uuid.UUID) ([]Comment, error)
	DeleteComment(ctx context.Context, commentID, agentID uuid.UUID) error
}

// Verify that Repository implements RepositoryInterface
var _ RepositoryInterface = (*Repository)(nil)
