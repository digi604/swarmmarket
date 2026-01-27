package capability

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles capability database operations.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new capability repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create inserts a new capability.
func (r *Repository) Create(ctx context.Context, cap *Capability) error {
	query := `
		INSERT INTO capabilities (
			id, agent_id, domain, type, subtype, name, description, version,
			input_schema, output_schema, status_events,
			geographic_scope, geo_center_lat, geo_center_lng, geo_radius_km, geo_polygon,
			available_hours, available_days, timezone,
			pricing_model, base_fee, percentage_fee, currency, pricing_details,
			response_time_seconds, completion_time_p50, completion_time_p95,
			is_active, is_accepting_tasks
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8,
			$9, $10, $11,
			$12, $13, $14, $15, $16,
			$17, $18, $19,
			$20, $21, $22, $23, $24,
			$25, $26, $27,
			$28, $29
		)
		RETURNING domain_path, created_at, updated_at`

	cap.ID = uuid.New()
	if cap.Version == "" {
		cap.Version = "1.0"
	}
	if cap.GeographicScope == "" {
		cap.GeographicScope = ScopeInternational
	}
	if cap.Timezone == "" {
		cap.Timezone = "UTC"
	}
	if cap.PricingModel == "" {
		cap.PricingModel = PricingFixed
	}
	if cap.Currency == "" {
		cap.Currency = "USD"
	}

	return r.pool.QueryRow(ctx, query,
		cap.ID, cap.AgentID, cap.Domain, cap.Type, cap.Subtype, cap.Name, cap.Description, cap.Version,
		cap.InputSchema, cap.OutputSchema, cap.StatusEvents,
		cap.GeographicScope, cap.GeoCenterLat, cap.GeoCenterLng, cap.GeoRadiusKM, cap.GeoPolygon,
		cap.AvailableHours, cap.AvailableDays, cap.Timezone,
		cap.PricingModel, cap.BaseFee, cap.PercentageFee, cap.Currency, cap.PricingDetails,
		cap.ResponseTimeSeconds, cap.CompletionTimeP50, cap.CompletionTimeP95,
		cap.IsActive, cap.IsAcceptingTasks,
	).Scan(&cap.DomainPath, &cap.CreatedAt, &cap.UpdatedAt)
}

// GetByID retrieves a capability by ID.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Capability, error) {
	query := `
		SELECT 
			c.id, c.agent_id, c.domain, c.type, c.subtype, c.domain_path,
			c.name, c.description, c.version,
			c.input_schema, c.output_schema, c.status_events,
			c.geographic_scope, c.geo_center_lat, c.geo_center_lng, c.geo_radius_km, c.geo_polygon,
			c.available_hours, c.available_days, c.timezone,
			c.pricing_model, c.base_fee, c.percentage_fee, c.currency, c.pricing_details,
			c.response_time_seconds, c.completion_time_p50, c.completion_time_p95,
			c.is_active, c.is_accepting_tasks,
			c.total_tasks, c.successful_tasks, c.failed_tasks, c.average_rating,
			c.created_at, c.updated_at,
			a.name as agent_name
		FROM capabilities c
		JOIN agents a ON a.id = c.agent_id
		WHERE c.id = $1`

	cap := &Capability{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&cap.ID, &cap.AgentID, &cap.Domain, &cap.Type, &cap.Subtype, &cap.DomainPath,
		&cap.Name, &cap.Description, &cap.Version,
		&cap.InputSchema, &cap.OutputSchema, &cap.StatusEvents,
		&cap.GeographicScope, &cap.GeoCenterLat, &cap.GeoCenterLng, &cap.GeoRadiusKM, &cap.GeoPolygon,
		&cap.AvailableHours, &cap.AvailableDays, &cap.Timezone,
		&cap.PricingModel, &cap.BaseFee, &cap.PercentageFee, &cap.Currency, &cap.PricingDetails,
		&cap.ResponseTimeSeconds, &cap.CompletionTimeP50, &cap.CompletionTimeP95,
		&cap.IsActive, &cap.IsAcceptingTasks,
		&cap.TotalTasks, &cap.SuccessfulTasks, &cap.FailedTasks, &cap.AverageRating,
		&cap.CreatedAt, &cap.UpdatedAt,
		&cap.AgentName,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get capability: %w", err)
	}

	// Load current verification
	cap.Verification, _ = r.GetCurrentVerification(ctx, id)

	return cap, nil
}

// GetByAgentID retrieves all capabilities for an agent.
func (r *Repository) GetByAgentID(ctx context.Context, agentID uuid.UUID) ([]*Capability, error) {
	query := `
		SELECT 
			c.id, c.agent_id, c.domain, c.type, c.subtype, c.domain_path,
			c.name, c.description, c.version,
			c.input_schema, c.output_schema, c.status_events,
			c.geographic_scope, c.geo_center_lat, c.geo_center_lng, c.geo_radius_km, c.geo_polygon,
			c.available_hours, c.available_days, c.timezone,
			c.pricing_model, c.base_fee, c.percentage_fee, c.currency, c.pricing_details,
			c.response_time_seconds, c.completion_time_p50, c.completion_time_p95,
			c.is_active, c.is_accepting_tasks,
			c.total_tasks, c.successful_tasks, c.failed_tasks, c.average_rating,
			c.created_at, c.updated_at
		FROM capabilities c
		WHERE c.agent_id = $1
		ORDER BY c.created_at DESC`

	rows, err := r.pool.Query(ctx, query, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query capabilities: %w", err)
	}
	defer rows.Close()

	var capabilities []*Capability
	for rows.Next() {
		cap := &Capability{}
		err := rows.Scan(
			&cap.ID, &cap.AgentID, &cap.Domain, &cap.Type, &cap.Subtype, &cap.DomainPath,
			&cap.Name, &cap.Description, &cap.Version,
			&cap.InputSchema, &cap.OutputSchema, &cap.StatusEvents,
			&cap.GeographicScope, &cap.GeoCenterLat, &cap.GeoCenterLng, &cap.GeoRadiusKM, &cap.GeoPolygon,
			&cap.AvailableHours, &cap.AvailableDays, &cap.Timezone,
			&cap.PricingModel, &cap.BaseFee, &cap.PercentageFee, &cap.Currency, &cap.PricingDetails,
			&cap.ResponseTimeSeconds, &cap.CompletionTimeP50, &cap.CompletionTimeP95,
			&cap.IsActive, &cap.IsAcceptingTasks,
			&cap.TotalTasks, &cap.SuccessfulTasks, &cap.FailedTasks, &cap.AverageRating,
			&cap.CreatedAt, &cap.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan capability: %w", err)
		}
		capabilities = append(capabilities, cap)
	}

	return capabilities, nil
}

// Update updates a capability.
func (r *Repository) Update(ctx context.Context, cap *Capability) error {
	query := `
		UPDATE capabilities SET
			name = $2, description = $3,
			input_schema = $4, output_schema = $5, status_events = $6,
			geographic_scope = $7, geo_center_lat = $8, geo_center_lng = $9, 
			geo_radius_km = $10, geo_polygon = $11,
			available_hours = $12, available_days = $13, timezone = $14,
			pricing_model = $15, base_fee = $16, percentage_fee = $17, 
			currency = $18, pricing_details = $19,
			response_time_seconds = $20, completion_time_p50 = $21, completion_time_p95 = $22,
			is_active = $23, is_accepting_tasks = $24
		WHERE id = $1
		RETURNING updated_at`

	return r.pool.QueryRow(ctx, query,
		cap.ID, cap.Name, cap.Description,
		cap.InputSchema, cap.OutputSchema, cap.StatusEvents,
		cap.GeographicScope, cap.GeoCenterLat, cap.GeoCenterLng,
		cap.GeoRadiusKM, cap.GeoPolygon,
		cap.AvailableHours, cap.AvailableDays, cap.Timezone,
		cap.PricingModel, cap.BaseFee, cap.PercentageFee,
		cap.Currency, cap.PricingDetails,
		cap.ResponseTimeSeconds, cap.CompletionTimeP50, cap.CompletionTimeP95,
		cap.IsActive, cap.IsAcceptingTasks,
	).Scan(&cap.UpdatedAt)
}

// Delete soft-deletes a capability.
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE capabilities SET is_active = false WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// Search searches capabilities with filters.
func (r *Repository) Search(ctx context.Context, req *SearchCapabilitiesRequest) (*SearchCapabilitiesResponse, error) {
	var conditions []string
	var args []interface{}
	argNum := 1

	// Base conditions
	conditions = append(conditions, "c.is_active = true")
	conditions = append(conditions, "c.is_accepting_tasks = true")

	// Domain filtering
	if req.DomainPath != "" {
		conditions = append(conditions, fmt.Sprintf("c.domain_path LIKE $%d", argNum))
		args = append(args, req.DomainPath+"%")
		argNum++
	} else {
		if req.Domain != "" {
			conditions = append(conditions, fmt.Sprintf("c.domain = $%d", argNum))
			args = append(args, req.Domain)
			argNum++
		}
		if req.Type != "" {
			conditions = append(conditions, fmt.Sprintf("c.type = $%d", argNum))
			args = append(args, req.Type)
			argNum++
		}
		if req.Subtype != "" {
			conditions = append(conditions, fmt.Sprintf("c.subtype = $%d", argNum))
			args = append(args, req.Subtype)
			argNum++
		}
	}

	// Full-text search
	if req.Query != "" {
		conditions = append(conditions, fmt.Sprintf(
			"to_tsvector('english', c.name || ' ' || COALESCE(c.description, '')) @@ plainto_tsquery('english', $%d)", argNum))
		args = append(args, req.Query)
		argNum++
	}

	// Location filtering
	var distanceSelect string
	if req.Lat != nil && req.Lng != nil {
		distanceSelect = fmt.Sprintf(`
			, CASE WHEN c.geo_center_lat IS NOT NULL THEN
				6371 * acos(cos(radians($%d)) * cos(radians(c.geo_center_lat)) * 
				cos(radians(c.geo_center_lng) - radians($%d)) + 
				sin(radians($%d)) * sin(radians(c.geo_center_lat)))
			ELSE NULL END as distance_km`, argNum, argNum+1, argNum+2)
		args = append(args, *req.Lat, *req.Lng, *req.Lat)
		argNum += 3

		if req.RadiusKM != nil {
			conditions = append(conditions, fmt.Sprintf(`(
				c.geographic_scope = 'international' OR
				(c.geo_center_lat IS NOT NULL AND 
				 6371 * acos(cos(radians($%d)) * cos(radians(c.geo_center_lat)) * 
				 cos(radians(c.geo_center_lng) - radians($%d)) + 
				 sin(radians($%d)) * sin(radians(c.geo_center_lat))) <= $%d)
			)`, argNum, argNum+1, argNum+2, argNum+3))
			args = append(args, *req.Lat, *req.Lng, *req.Lat, *req.RadiusKM)
			argNum += 4
		}
	}

	// Verification filter
	if req.VerifiedOnly {
		conditions = append(conditions, `EXISTS (
			SELECT 1 FROM capability_verifications cv 
			WHERE cv.capability_id = c.id 
			AND cv.is_current = true 
			AND cv.level IN ('tested', 'verified', 'certified')
		)`)
	}

	// Rating filter
	if req.MinRating != nil {
		conditions = append(conditions, fmt.Sprintf("c.average_rating >= $%d", argNum))
		args = append(args, *req.MinRating)
		argNum++
	}

	// Build query
	whereClause := strings.Join(conditions, " AND ")

	// Sorting
	orderBy := "c.average_rating DESC, c.total_tasks DESC"
	if req.SortBy != "" {
		switch req.SortBy {
		case "reputation":
			orderBy = "c.average_rating DESC"
		case "price":
			orderBy = "c.base_fee ASC NULLS LAST"
		case "response_time":
			orderBy = "c.response_time_seconds ASC NULLS LAST"
		case "distance":
			if distanceSelect != "" {
				orderBy = "distance_km ASC NULLS LAST"
			}
		}
		if req.SortOrder == "asc" {
			orderBy = strings.Replace(orderBy, "DESC", "ASC", 1)
		}
	}

	// Pagination
	limit := 20
	if req.Limit > 0 && req.Limit <= 100 {
		limit = req.Limit
	}
	offset := req.Offset

	// Count query
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM capabilities c WHERE %s`, whereClause)
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count capabilities: %w", err)
	}

	// Main query
	query := fmt.Sprintf(`
		SELECT 
			c.id, c.agent_id, c.domain, c.type, c.subtype, c.domain_path,
			c.name, c.description, c.version,
			c.input_schema, c.output_schema, c.status_events,
			c.geographic_scope, c.geo_center_lat, c.geo_center_lng, c.geo_radius_km, c.geo_polygon,
			c.available_hours, c.available_days, c.timezone,
			c.pricing_model, c.base_fee, c.percentage_fee, c.currency, c.pricing_details,
			c.response_time_seconds, c.completion_time_p50, c.completion_time_p95,
			c.is_active, c.is_accepting_tasks,
			c.total_tasks, c.successful_tasks, c.failed_tasks, c.average_rating,
			c.created_at, c.updated_at,
			a.name as agent_name
			%s
		FROM capabilities c
		JOIN agents a ON a.id = c.agent_id
		WHERE %s
		ORDER BY %s
		LIMIT %d OFFSET %d`,
		distanceSelect, whereClause, orderBy, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search capabilities: %w", err)
	}
	defer rows.Close()

	var results []CapabilityMatch
	for rows.Next() {
		var cap CapabilityMatch
		var distanceKM *float64

		scanArgs := []interface{}{
			&cap.ID, &cap.AgentID, &cap.Domain, &cap.Type, &cap.Subtype, &cap.DomainPath,
			&cap.Name, &cap.Description, &cap.Version,
			&cap.InputSchema, &cap.OutputSchema, &cap.StatusEvents,
			&cap.GeographicScope, &cap.GeoCenterLat, &cap.GeoCenterLng, &cap.GeoRadiusKM, &cap.GeoPolygon,
			&cap.AvailableHours, &cap.AvailableDays, &cap.Timezone,
			&cap.PricingModel, &cap.BaseFee, &cap.PercentageFee, &cap.Currency, &cap.PricingDetails,
			&cap.ResponseTimeSeconds, &cap.CompletionTimeP50, &cap.CompletionTimeP95,
			&cap.IsActive, &cap.IsAcceptingTasks,
			&cap.TotalTasks, &cap.SuccessfulTasks, &cap.FailedTasks, &cap.AverageRating,
			&cap.CreatedAt, &cap.UpdatedAt,
			&cap.AgentName,
		}

		if distanceSelect != "" {
			scanArgs = append(scanArgs, &distanceKM)
		}

		if err := rows.Scan(scanArgs...); err != nil {
			return nil, fmt.Errorf("failed to scan capability: %w", err)
		}

		cap.DistanceKM = distanceKM

		// Load verification
		cap.Verification, _ = r.GetCurrentVerification(ctx, cap.ID)

		results = append(results, cap)
	}

	return &SearchCapabilitiesResponse{
		Capabilities: results,
		Total:        total,
		Limit:        limit,
		Offset:       offset,
	}, nil
}

// --- Verification methods ---

// CreateVerification creates a new verification record.
func (r *Repository) CreateVerification(ctx context.Context, v *Verification) error {
	// First, mark any existing current verification as not current
	_, err := r.pool.Exec(ctx, 
		`UPDATE capability_verifications SET is_current = false WHERE capability_id = $1 AND is_current = true`,
		v.CapabilityID)
	if err != nil {
		return fmt.Errorf("failed to update existing verification: %w", err)
	}

	query := `
		INSERT INTO capability_verifications (
			id, capability_id, level, method, proof, test_results,
			success_rate, total_transactions, avg_response_time_ms,
			verified_at, expires_at, verified_by, is_current
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING created_at`

	v.ID = uuid.New()
	v.IsCurrent = true

	return r.pool.QueryRow(ctx, query,
		v.ID, v.CapabilityID, v.Level, v.Method, v.Proof, v.TestResults,
		v.SuccessRate, v.TotalTxns, v.AvgResponseTime,
		v.VerifiedAt, v.ExpiresAt, v.VerifiedBy, v.IsCurrent,
	).Scan(&v.CreatedAt)
}

// GetCurrentVerification gets the current verification for a capability.
func (r *Repository) GetCurrentVerification(ctx context.Context, capabilityID uuid.UUID) (*Verification, error) {
	query := `
		SELECT id, capability_id, level, method, proof, test_results,
			success_rate, total_transactions, avg_response_time_ms,
			verified_at, expires_at, verified_by, is_current, created_at
		FROM capability_verifications
		WHERE capability_id = $1 AND is_current = true`

	v := &Verification{}
	err := r.pool.QueryRow(ctx, query, capabilityID).Scan(
		&v.ID, &v.CapabilityID, &v.Level, &v.Method, &v.Proof, &v.TestResults,
		&v.SuccessRate, &v.TotalTxns, &v.AvgResponseTime,
		&v.VerifiedAt, &v.ExpiresAt, &v.VerifiedBy, &v.IsCurrent, &v.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get verification: %w", err)
	}
	return v, nil
}

// --- Domain Taxonomy methods ---

// GetDomainTaxonomy retrieves the full domain taxonomy tree.
func (r *Repository) GetDomainTaxonomy(ctx context.Context) ([]DomainTaxonomy, error) {
	query := `
		SELECT id, path, parent_path, name, description, icon, schema_template, is_active, created_at
		FROM domain_taxonomy
		WHERE is_active = true
		ORDER BY path`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query taxonomy: %w", err)
	}
	defer rows.Close()

	var domains []DomainTaxonomy
	for rows.Next() {
		var d DomainTaxonomy
		err := rows.Scan(&d.ID, &d.Path, &d.ParentPath, &d.Name, &d.Description, &d.Icon, &d.SchemaTemplate, &d.IsActive, &d.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan taxonomy: %w", err)
		}
		domains = append(domains, d)
	}

	return domains, nil
}

// GetDomainByPath retrieves a single domain by path.
func (r *Repository) GetDomainByPath(ctx context.Context, path string) (*DomainTaxonomy, error) {
	query := `
		SELECT id, path, parent_path, name, description, icon, schema_template, is_active, created_at
		FROM domain_taxonomy
		WHERE path = $1`

	d := &DomainTaxonomy{}
	err := r.pool.QueryRow(ctx, query, path).Scan(
		&d.ID, &d.Path, &d.ParentPath, &d.Name, &d.Description, &d.Icon, &d.SchemaTemplate, &d.IsActive, &d.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get domain: %w", err)
	}
	return d, nil
}

// --- Stats methods ---

// IncrementTaskStats updates task statistics for a capability.
func (r *Repository) IncrementTaskStats(ctx context.Context, capabilityID uuid.UUID, success bool) error {
	var query string
	if success {
		query = `UPDATE capabilities SET total_tasks = total_tasks + 1, successful_tasks = successful_tasks + 1 WHERE id = $1`
	} else {
		query = `UPDATE capabilities SET total_tasks = total_tasks + 1, failed_tasks = failed_tasks + 1 WHERE id = $1`
	}
	_, err := r.pool.Exec(ctx, query, capabilityID)
	return err
}

// UpdateAverageRating updates the average rating for a capability.
func (r *Repository) UpdateAverageRating(ctx context.Context, capabilityID uuid.UUID, newRating float64) error {
	// Simple approach: weighted average based on total transactions
	query := `
		UPDATE capabilities 
		SET average_rating = (average_rating * total_tasks + $2) / (total_tasks + 1)
		WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, capabilityID, newRating)
	return err
}
