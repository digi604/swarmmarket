package marketplace

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/digi604/swarmmarket/backend/internal/common"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrListingNotFound = errors.New("listing not found")
	ErrRequestNotFound = errors.New("request not found")
	ErrOfferNotFound   = errors.New("offer not found")
)

// Repository handles marketplace data persistence.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new marketplace repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// generateUniqueSlug generates a unique slug from title + short UUID.
func generateUniqueSlug(title string, id uuid.UUID) string {
	baseSlug := common.GenerateSlug(title)
	if baseSlug == "" {
		baseSlug = "item"
	}
	// Use first 8 chars of UUID for uniqueness
	return fmt.Sprintf("%s-%s", baseSlug, id.String()[:8])
}

// --- Listings ---

// CreateListing inserts a new listing.
func (r *Repository) CreateListing(ctx context.Context, listing *Listing) error {
	// Generate unique slug from title + UUID prefix
	listing.Slug = generateUniqueSlug(listing.Title, listing.ID)

	query := `
		INSERT INTO listings (
			id, slug, seller_id, category_id, title, description, listing_type,
			price_amount, price_currency, quantity, geographic_scope,
			location_lat, location_lng, location_radius_km, status,
			expires_at, metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
		)
	`
	_, err := r.pool.Exec(ctx, query,
		listing.ID, listing.Slug, listing.SellerID, listing.CategoryID, listing.Title,
		listing.Description, listing.ListingType, listing.PriceAmount,
		listing.PriceCurrency, listing.Quantity, listing.GeographicScope,
		listing.LocationLat, listing.LocationLng, listing.LocationRadius,
		listing.Status, listing.ExpiresAt, listing.Metadata,
		listing.CreatedAt, listing.UpdatedAt,
	)
	return err
}

// GetListingByID retrieves a listing by ID.
func (r *Repository) GetListingByID(ctx context.Context, id uuid.UUID) (*Listing, error) {
	query := `
		SELECT l.id, COALESCE(l.slug, ''), l.seller_id, l.category_id, l.title, l.description, l.listing_type,
			l.price_amount, l.price_currency, l.quantity, l.geographic_scope,
			l.location_lat, l.location_lng, l.location_radius_km, l.status,
			l.expires_at, l.metadata, l.created_at, l.updated_at,
			COALESCE(a.name, '') as seller_name,
			a.avatar_url as seller_avatar_url,
			COALESCE(a.average_rating, 0) as seller_rating,
			(SELECT COUNT(*) FROM ratings WHERE rated_agent_id = l.seller_id) as seller_rating_count
		FROM listings l
		LEFT JOIN agents a ON l.seller_id = a.id
		WHERE l.id = $1
	`
	listing := &Listing{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&listing.ID, &listing.Slug, &listing.SellerID, &listing.CategoryID, &listing.Title,
		&listing.Description, &listing.ListingType, &listing.PriceAmount,
		&listing.PriceCurrency, &listing.Quantity, &listing.GeographicScope,
		&listing.LocationLat, &listing.LocationLng, &listing.LocationRadius,
		&listing.Status, &listing.ExpiresAt, &listing.Metadata,
		&listing.CreatedAt, &listing.UpdatedAt,
		&listing.SellerName, &listing.SellerAvatarURL, &listing.SellerRating, &listing.SellerRatingCount,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrListingNotFound
	}
	return listing, err
}

// GetListingBySlug retrieves a listing by slug (extracts UUID suffix).
func (r *Repository) GetListingBySlug(ctx context.Context, slug string) (*Listing, error) {
	// Extract UUID prefix from end of slug (last 8 chars after final dash)
	parts := strings.Split(slug, "-")
	if len(parts) < 2 {
		return nil, ErrListingNotFound
	}
	uuidPrefix := parts[len(parts)-1]
	if len(uuidPrefix) != 8 {
		return nil, ErrListingNotFound
	}

	query := `
		SELECT l.id, COALESCE(l.slug, ''), l.seller_id, l.category_id, l.title, l.description, l.listing_type,
			l.price_amount, l.price_currency, l.quantity, l.geographic_scope,
			l.location_lat, l.location_lng, l.location_radius_km, l.status,
			l.expires_at, l.metadata, l.created_at, l.updated_at,
			COALESCE(a.name, '') as seller_name,
			a.avatar_url as seller_avatar_url,
			COALESCE(a.average_rating, 0) as seller_rating,
			(SELECT COUNT(*) FROM ratings WHERE rated_agent_id = l.seller_id) as seller_rating_count
		FROM listings l
		LEFT JOIN agents a ON l.seller_id = a.id
		WHERE l.id::text LIKE $1
	`
	listing := &Listing{}
	err := r.pool.QueryRow(ctx, query, uuidPrefix+"%").Scan(
		&listing.ID, &listing.Slug, &listing.SellerID, &listing.CategoryID, &listing.Title,
		&listing.Description, &listing.ListingType, &listing.PriceAmount,
		&listing.PriceCurrency, &listing.Quantity, &listing.GeographicScope,
		&listing.LocationLat, &listing.LocationLng, &listing.LocationRadius,
		&listing.Status, &listing.ExpiresAt, &listing.Metadata,
		&listing.CreatedAt, &listing.UpdatedAt,
		&listing.SellerName, &listing.SellerAvatarURL, &listing.SellerRating, &listing.SellerRatingCount,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrListingNotFound
	}
	return listing, err
}

// SearchListings searches for listings with filters.
func (r *Repository) SearchListings(ctx context.Context, params SearchListingsParams) (*ListResult[Listing], error) {
	var conditions []string
	var args []interface{}
	argNum := 1

	// Only show active listings by default
	if params.Status != nil {
		conditions = append(conditions, fmt.Sprintf("l.status = $%d", argNum))
		args = append(args, *params.Status)
		argNum++
	} else {
		conditions = append(conditions, fmt.Sprintf("l.status = $%d", argNum))
		args = append(args, ListingStatusActive)
		argNum++
	}

	if params.CategoryID != nil {
		conditions = append(conditions, fmt.Sprintf("l.category_id = $%d", argNum))
		args = append(args, *params.CategoryID)
		argNum++
	}
	if params.ListingType != nil {
		conditions = append(conditions, fmt.Sprintf("l.listing_type = $%d", argNum))
		args = append(args, *params.ListingType)
		argNum++
	}
	if params.MinPrice != nil {
		conditions = append(conditions, fmt.Sprintf("l.price_amount >= $%d", argNum))
		args = append(args, *params.MinPrice)
		argNum++
	}
	if params.MaxPrice != nil {
		conditions = append(conditions, fmt.Sprintf("l.price_amount <= $%d", argNum))
		args = append(args, *params.MaxPrice)
		argNum++
	}
	if params.GeographicScope != nil {
		conditions = append(conditions, fmt.Sprintf("l.geographic_scope = $%d", argNum))
		args = append(args, *params.GeographicScope)
		argNum++
	}
	if params.SellerID != nil {
		conditions = append(conditions, fmt.Sprintf("l.seller_id = $%d", argNum))
		args = append(args, *params.SellerID)
		argNum++
	}
	if params.Query != "" {
		conditions = append(conditions, fmt.Sprintf("(l.title ILIKE $%d OR l.description ILIKE $%d)", argNum, argNum))
		args = append(args, "%"+params.Query+"%")
		argNum++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM listings l %s", whereClause)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	// Apply pagination
	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := params.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
		SELECT l.id, COALESCE(l.slug, ''), l.seller_id, l.category_id, l.title, l.description, l.listing_type,
			l.price_amount, l.price_currency, l.quantity, l.geographic_scope,
			l.location_lat, l.location_lng, l.location_radius_km, l.status,
			l.expires_at, l.metadata, l.created_at, l.updated_at,
			COALESCE(a.name, '') as seller_name,
			a.avatar_url as seller_avatar_url,
			COALESCE(a.average_rating, 0) as seller_rating,
			(SELECT COUNT(*) FROM ratings WHERE rated_agent_id = l.seller_id) as seller_rating_count
		FROM listings l
		LEFT JOIN agents a ON l.seller_id = a.id
		%s
		ORDER BY l.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argNum, argNum+1)

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var listings []Listing
	for rows.Next() {
		var l Listing
		if err := rows.Scan(
			&l.ID, &l.Slug, &l.SellerID, &l.CategoryID, &l.Title, &l.Description,
			&l.ListingType, &l.PriceAmount, &l.PriceCurrency, &l.Quantity,
			&l.GeographicScope, &l.LocationLat, &l.LocationLng, &l.LocationRadius,
			&l.Status, &l.ExpiresAt, &l.Metadata, &l.CreatedAt, &l.UpdatedAt,
			&l.SellerName, &l.SellerAvatarURL, &l.SellerRating, &l.SellerRatingCount,
		); err != nil {
			return nil, err
		}
		listings = append(listings, l)
	}

	return &ListResult[Listing]{
		Items:  listings,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

// DeleteListing soft-deletes a listing by setting status to expired.
func (r *Repository) DeleteListing(ctx context.Context, id uuid.UUID, sellerID uuid.UUID) error {
	query := `UPDATE listings SET status = $1, updated_at = $2 WHERE id = $3 AND seller_id = $4`
	result, err := r.pool.Exec(ctx, query, ListingStatusExpired, time.Now().UTC(), id, sellerID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrListingNotFound
	}
	return nil
}

// --- Requests ---

// CreateRequest inserts a new request.
func (r *Repository) CreateRequest(ctx context.Context, req *Request) error {
	// Generate unique slug from title + UUID prefix
	req.Slug = generateUniqueSlug(req.Title, req.ID)

	query := `
		INSERT INTO requests (
			id, slug, requester_id, category_id, title, description, request_type,
			budget_min, budget_max, budget_currency, quantity, geographic_scope,
			location_lat, location_lng, location_radius_km, status,
			expires_at, metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
		)
	`
	_, err := r.pool.Exec(ctx, query,
		req.ID, req.Slug, req.RequesterID, req.CategoryID, req.Title, req.Description,
		req.RequestType, req.BudgetMin, req.BudgetMax, req.BudgetCurrency,
		req.Quantity, req.GeographicScope, req.LocationLat, req.LocationLng,
		req.LocationRadius, req.Status, req.ExpiresAt, req.Metadata,
		req.CreatedAt, req.UpdatedAt,
	)
	return err
}

// GetRequestByID retrieves a request by ID.
func (r *Repository) GetRequestByID(ctx context.Context, id uuid.UUID) (*Request, error) {
	query := `
		SELECT r.id, COALESCE(r.slug, ''), r.requester_id, r.category_id, r.title, r.description,
			r.request_type, r.budget_min, r.budget_max, r.budget_currency,
			r.quantity, r.geographic_scope, r.location_lat, r.location_lng,
			r.location_radius_km, r.status, r.expires_at, r.metadata,
			r.created_at, r.updated_at,
			(SELECT COUNT(*) FROM offers WHERE request_id = r.id) as offer_count,
			COALESCE(a.name, '') as requester_name,
			a.avatar_url as requester_avatar_url,
			COALESCE(a.average_rating, 0) as requester_rating
		FROM requests r
		LEFT JOIN agents a ON r.requester_id = a.id
		WHERE r.id = $1
	`
	req := &Request{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&req.ID, &req.Slug, &req.RequesterID, &req.CategoryID, &req.Title, &req.Description,
		&req.RequestType, &req.BudgetMin, &req.BudgetMax, &req.BudgetCurrency,
		&req.Quantity, &req.GeographicScope, &req.LocationLat, &req.LocationLng,
		&req.LocationRadius, &req.Status, &req.ExpiresAt, &req.Metadata,
		&req.CreatedAt, &req.UpdatedAt, &req.OfferCount,
		&req.RequesterName, &req.RequesterAvatarURL, &req.RequesterRating,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRequestNotFound
	}
	return req, err
}

// GetRequestBySlug retrieves a request by slug (extracts UUID suffix).
func (r *Repository) GetRequestBySlug(ctx context.Context, slug string) (*Request, error) {
	parts := strings.Split(slug, "-")
	if len(parts) < 2 {
		return nil, ErrRequestNotFound
	}
	uuidPrefix := parts[len(parts)-1]
	if len(uuidPrefix) != 8 {
		return nil, ErrRequestNotFound
	}

	query := `
		SELECT r.id, COALESCE(r.slug, ''), r.requester_id, r.category_id, r.title, r.description,
			r.request_type, r.budget_min, r.budget_max, r.budget_currency,
			r.quantity, r.geographic_scope, r.location_lat, r.location_lng,
			r.location_radius_km, r.status, r.expires_at, r.metadata,
			r.created_at, r.updated_at,
			(SELECT COUNT(*) FROM offers WHERE request_id = r.id) as offer_count,
			COALESCE(a.name, '') as requester_name,
			a.avatar_url as requester_avatar_url,
			COALESCE(a.average_rating, 0) as requester_rating
		FROM requests r
		LEFT JOIN agents a ON r.requester_id = a.id
		WHERE r.id::text LIKE $1
	`
	req := &Request{}
	err := r.pool.QueryRow(ctx, query, uuidPrefix+"%").Scan(
		&req.ID, &req.Slug, &req.RequesterID, &req.CategoryID, &req.Title, &req.Description,
		&req.RequestType, &req.BudgetMin, &req.BudgetMax, &req.BudgetCurrency,
		&req.Quantity, &req.GeographicScope, &req.LocationLat, &req.LocationLng,
		&req.LocationRadius, &req.Status, &req.ExpiresAt, &req.Metadata,
		&req.CreatedAt, &req.UpdatedAt, &req.OfferCount,
		&req.RequesterName, &req.RequesterAvatarURL, &req.RequesterRating,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRequestNotFound
	}
	return req, err
}

// SearchRequests searches for requests with filters.
func (r *Repository) SearchRequests(ctx context.Context, params SearchRequestsParams) (*ListResult[Request], error) {
	var conditions []string
	var args []interface{}
	argNum := 1

	// Only show open requests by default
	if params.Status != nil {
		conditions = append(conditions, fmt.Sprintf("r.status = $%d", argNum))
		args = append(args, *params.Status)
		argNum++
	} else {
		conditions = append(conditions, fmt.Sprintf("r.status = $%d", argNum))
		args = append(args, RequestStatusOpen)
		argNum++
	}

	if params.CategoryID != nil {
		conditions = append(conditions, fmt.Sprintf("r.category_id = $%d", argNum))
		args = append(args, *params.CategoryID)
		argNum++
	}
	if params.RequestType != nil {
		conditions = append(conditions, fmt.Sprintf("r.request_type = $%d", argNum))
		args = append(args, *params.RequestType)
		argNum++
	}
	if params.MinBudget != nil {
		conditions = append(conditions, fmt.Sprintf("r.budget_max >= $%d", argNum))
		args = append(args, *params.MinBudget)
		argNum++
	}
	if params.MaxBudget != nil {
		conditions = append(conditions, fmt.Sprintf("r.budget_min <= $%d", argNum))
		args = append(args, *params.MaxBudget)
		argNum++
	}
	if params.GeographicScope != nil {
		conditions = append(conditions, fmt.Sprintf("r.geographic_scope = $%d", argNum))
		args = append(args, *params.GeographicScope)
		argNum++
	}
	if params.RequesterID != nil {
		conditions = append(conditions, fmt.Sprintf("r.requester_id = $%d", argNum))
		args = append(args, *params.RequesterID)
		argNum++
	}
	if params.Query != "" {
		conditions = append(conditions, fmt.Sprintf("(r.title ILIKE $%d OR r.description ILIKE $%d)", argNum, argNum))
		args = append(args, "%"+params.Query+"%")
		argNum++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM requests r %s", whereClause)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := params.Offset
	if offset < 0 {
		offset = 0
	}

	// Determine sort order
	orderBy := "r.created_at DESC" // default: newest
	switch params.SortBy {
	case "budget_high":
		orderBy = "COALESCE(r.budget_max, r.budget_min, 0) DESC, r.created_at DESC"
	case "budget_low":
		orderBy = "COALESCE(r.budget_min, r.budget_max, 0) ASC, r.created_at DESC"
	case "ending_soon":
		orderBy = "r.expires_at ASC NULLS LAST, r.created_at DESC"
	case "newest":
		orderBy = "r.created_at DESC"
	}

	query := fmt.Sprintf(`
		SELECT r.id, COALESCE(r.slug, ''), r.requester_id, r.category_id, r.title, r.description,
			r.request_type, r.budget_min, r.budget_max, r.budget_currency,
			r.quantity, r.geographic_scope, r.location_lat, r.location_lng,
			r.location_radius_km, r.status, r.expires_at, r.metadata,
			r.created_at, r.updated_at,
			(SELECT COUNT(*) FROM offers WHERE request_id = r.id) as offer_count,
			COALESCE(a.name, '') as requester_name,
			a.avatar_url as requester_avatar_url,
			COALESCE(a.average_rating, 0) as requester_rating
		FROM requests r
		LEFT JOIN agents a ON r.requester_id = a.id
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argNum, argNum+1)

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []Request
	for rows.Next() {
		var req Request
		if err := rows.Scan(
			&req.ID, &req.Slug, &req.RequesterID, &req.CategoryID, &req.Title, &req.Description,
			&req.RequestType, &req.BudgetMin, &req.BudgetMax, &req.BudgetCurrency,
			&req.Quantity, &req.GeographicScope, &req.LocationLat, &req.LocationLng,
			&req.LocationRadius, &req.Status, &req.ExpiresAt, &req.Metadata,
			&req.CreatedAt, &req.UpdatedAt, &req.OfferCount,
			&req.RequesterName, &req.RequesterAvatarURL, &req.RequesterRating,
		); err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}

	return &ListResult[Request]{
		Items:  requests,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

// --- Offers ---

// CreateOffer inserts a new offer.
func (r *Repository) CreateOffer(ctx context.Context, offer *Offer) error {
	query := `
		INSERT INTO offers (
			id, request_id, offerer_id, listing_id, price_amount, price_currency,
			description, delivery_terms, valid_until, status, metadata,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
	`
	_, err := r.pool.Exec(ctx, query,
		offer.ID, offer.RequestID, offer.OffererID, offer.ListingID,
		offer.PriceAmount, offer.PriceCurrency, offer.Description,
		offer.DeliveryTerms, offer.ValidUntil, offer.Status, offer.Metadata,
		offer.CreatedAt, offer.UpdatedAt,
	)
	return err
}

// GetOfferByID retrieves an offer by ID.
func (r *Repository) GetOfferByID(ctx context.Context, id uuid.UUID) (*Offer, error) {
	query := `
		SELECT id, request_id, offerer_id, listing_id, price_amount, price_currency,
			description, delivery_terms, valid_until, status, metadata,
			created_at, updated_at
		FROM offers WHERE id = $1
	`
	offer := &Offer{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&offer.ID, &offer.RequestID, &offer.OffererID, &offer.ListingID,
		&offer.PriceAmount, &offer.PriceCurrency, &offer.Description,
		&offer.DeliveryTerms, &offer.ValidUntil, &offer.Status, &offer.Metadata,
		&offer.CreatedAt, &offer.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrOfferNotFound
	}
	return offer, err
}

// GetOffersByRequestID retrieves all offers for a request.
func (r *Repository) GetOffersByRequestID(ctx context.Context, requestID uuid.UUID) ([]Offer, error) {
	query := `
		SELECT o.id, o.request_id, o.offerer_id, o.listing_id, o.price_amount, o.price_currency,
			o.description, o.delivery_terms, o.valid_until, o.status, o.metadata,
			o.created_at, o.updated_at,
			COALESCE(a.name, '') AS offerer_name
		FROM offers o
		LEFT JOIN agents a ON a.id = o.offerer_id
		WHERE o.request_id = $1
		ORDER BY o.created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, requestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var offers []Offer
	for rows.Next() {
		var o Offer
		if err := rows.Scan(
			&o.ID, &o.RequestID, &o.OffererID, &o.ListingID, &o.PriceAmount,
			&o.PriceCurrency, &o.Description, &o.DeliveryTerms, &o.ValidUntil,
			&o.Status, &o.Metadata, &o.CreatedAt, &o.UpdatedAt,
			&o.OffererName,
		); err != nil {
			return nil, err
		}
		offers = append(offers, o)
	}
	return offers, nil
}

// UpdateOfferStatus updates the status of an offer.
func (r *Repository) UpdateOfferStatus(ctx context.Context, id uuid.UUID, status OfferStatus) error {
	query := `UPDATE offers SET status = $1, updated_at = $2 WHERE id = $3`
	result, err := r.pool.Exec(ctx, query, status, time.Now().UTC(), id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrOfferNotFound
	}
	return nil
}

// --- Categories ---

// GetCategories retrieves all categories.
func (r *Repository) GetCategories(ctx context.Context) ([]Category, error) {
	query := `
		SELECT id, parent_id, name, slug, description, metadata, created_at
		FROM categories
		ORDER BY name
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(
			&c.ID, &c.ParentID, &c.Name, &c.Slug, &c.Description,
			&c.Metadata, &c.CreatedAt,
		); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}

// --- Comments ---

// CreateComment creates a new comment on a listing.
func (r *Repository) CreateComment(ctx context.Context, comment *Comment) error {
	query := `
		INSERT INTO listing_comments (id, listing_id, agent_id, parent_id, content, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.pool.Exec(ctx, query,
		comment.ID, comment.ListingID, comment.AgentID, comment.ParentID,
		comment.Content, comment.CreatedAt, comment.UpdatedAt,
	)
	return err
}

// GetCommentsByListingID retrieves all top-level comments for a listing.
func (r *Repository) GetCommentsByListingID(ctx context.Context, listingID uuid.UUID, limit, offset int) ([]Comment, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM listing_comments WHERE listing_id = $1 AND parent_id IS NULL`
	if err := r.pool.QueryRow(ctx, countQuery, listingID).Scan(&total); err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 20
	}

	query := `
		SELECT c.id, c.listing_id, c.agent_id, c.parent_id, c.content, c.created_at, c.updated_at,
			COALESCE(a.name, '') as agent_name,
			a.avatar_url as agent_avatar_url,
			(SELECT COUNT(*) FROM listing_comments WHERE parent_id = c.id) as reply_count
		FROM listing_comments c
		LEFT JOIN agents a ON c.agent_id = a.id
		WHERE c.listing_id = $1 AND c.parent_id IS NULL
		ORDER BY c.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, listingID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var c Comment
		if err := rows.Scan(
			&c.ID, &c.ListingID, &c.AgentID, &c.ParentID, &c.Content,
			&c.CreatedAt, &c.UpdatedAt, &c.AgentName, &c.AgentAvatarURL, &c.ReplyCount,
		); err != nil {
			return nil, 0, err
		}
		comments = append(comments, c)
	}
	return comments, total, nil
}

// GetCommentReplies retrieves replies to a comment.
func (r *Repository) GetCommentReplies(ctx context.Context, parentID uuid.UUID) ([]Comment, error) {
	query := `
		SELECT c.id, c.listing_id, c.agent_id, c.parent_id, c.content, c.created_at, c.updated_at,
			COALESCE(a.name, '') as agent_name,
			a.avatar_url as agent_avatar_url,
			0 as reply_count
		FROM listing_comments c
		LEFT JOIN agents a ON c.agent_id = a.id
		WHERE c.parent_id = $1
		ORDER BY c.created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var c Comment
		if err := rows.Scan(
			&c.ID, &c.ListingID, &c.AgentID, &c.ParentID, &c.Content,
			&c.CreatedAt, &c.UpdatedAt, &c.AgentName, &c.AgentAvatarURL, &c.ReplyCount,
		); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, nil
}

// DeleteComment deletes a comment (only by the author).
func (r *Repository) DeleteComment(ctx context.Context, commentID, agentID uuid.UUID) error {
	query := `DELETE FROM listing_comments WHERE id = $1 AND agent_id = $2`
	result, err := r.pool.Exec(ctx, query, commentID, agentID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("comment not found or not authorized")
	}
	return nil
}
