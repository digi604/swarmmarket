-- SwarmMarket Capabilities Schema
-- Migration: 002_capabilities

-- Domain taxonomy (hierarchical capability categories)
CREATE TABLE domain_taxonomy (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    path VARCHAR(255) UNIQUE NOT NULL,          -- 'delivery/food/restaurant'
    parent_path VARCHAR(255),                    -- 'delivery/food'
    name VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(50),                            -- emoji or icon name
    schema_template JSONB,                       -- default input/output schema for this domain
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_domain_taxonomy_parent ON domain_taxonomy(parent_path);
CREATE INDEX idx_domain_taxonomy_active ON domain_taxonomy(is_active);

-- Capabilities (what an agent can do)
CREATE TABLE capabilities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    
    -- Taxonomy
    domain VARCHAR(50) NOT NULL,                 -- 'delivery'
    type VARCHAR(50) NOT NULL,                   -- 'food'
    subtype VARCHAR(50),                         -- 'restaurant'
    domain_path VARCHAR(255),                    -- 'delivery/food/restaurant' (denormalized for queries)
    
    -- Metadata
    name VARCHAR(255) NOT NULL,
    description TEXT,
    version VARCHAR(20) DEFAULT '1.0',
    
    -- Schemas (JSON Schema format)
    input_schema JSONB NOT NULL,                 -- what the capability accepts
    output_schema JSONB NOT NULL,                -- what the capability returns
    status_events JSONB,                         -- array of {event, description}
    
    -- Geographic constraints
    geographic_scope VARCHAR(20) DEFAULT 'international',  -- local, regional, national, international
    geo_center_lat DECIMAL(10, 8),
    geo_center_lng DECIMAL(11, 8),
    geo_radius_km INTEGER,
    geo_polygon JSONB,                           -- for complex regions
    
    -- Temporal constraints
    available_hours VARCHAR(20),                 -- '09:00-18:00'
    available_days VARCHAR(50),                  -- 'mon,tue,wed,thu,fri'
    timezone VARCHAR(50) DEFAULT 'UTC',
    
    -- Pricing
    pricing_model VARCHAR(20) DEFAULT 'fixed',   -- fixed, percentage, tiered, custom
    base_fee DECIMAL(20, 8),
    percentage_fee DECIMAL(5, 4),                -- e.g., 0.05 for 5%
    currency VARCHAR(10) DEFAULT 'USD',
    pricing_details JSONB,                       -- for complex pricing
    
    -- SLA
    response_time_seconds INTEGER,               -- expected time to accept/reject
    completion_time_p50 VARCHAR(20),             -- median completion time
    completion_time_p95 VARCHAR(20),             -- 95th percentile
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    is_accepting_tasks BOOLEAN DEFAULT true,
    
    -- Stats (denormalized for performance)
    total_tasks INTEGER DEFAULT 0,
    successful_tasks INTEGER DEFAULT 0,
    failed_tasks INTEGER DEFAULT 0,
    average_rating DECIMAL(3, 2) DEFAULT 0,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for capability search
CREATE INDEX idx_capabilities_agent ON capabilities(agent_id);
CREATE INDEX idx_capabilities_domain ON capabilities(domain, type, subtype);
CREATE INDEX idx_capabilities_domain_path ON capabilities(domain_path);
CREATE INDEX idx_capabilities_active ON capabilities(is_active, is_accepting_tasks);
CREATE INDEX idx_capabilities_geo ON capabilities(geo_center_lat, geo_center_lng) WHERE geo_center_lat IS NOT NULL;
CREATE INDEX idx_capabilities_rating ON capabilities(average_rating DESC);
CREATE INDEX idx_capabilities_pricing ON capabilities(pricing_model, currency);

-- GIN index for full-text search on name/description
CREATE INDEX idx_capabilities_search ON capabilities USING GIN(to_tsvector('english', name || ' ' || COALESCE(description, '')));

-- Capability verifications
CREATE TABLE capability_verifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    capability_id UUID NOT NULL REFERENCES capabilities(id) ON DELETE CASCADE,
    
    -- Verification details
    level VARCHAR(20) NOT NULL DEFAULT 'unverified',  -- unverified, tested, verified, certified
    method VARCHAR(50),                                -- api_test, transaction_history, attestation, manual
    
    -- Proof/evidence
    proof JSONB,                                       -- method-specific proof data
    test_results JSONB,                                -- automated test results
    
    -- Metrics at verification time
    success_rate DECIMAL(5, 4),                        -- e.g., 0.96 for 96%
    total_transactions INTEGER,
    avg_response_time_ms INTEGER,
    
    -- Validity
    verified_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    verified_by VARCHAR(255),                          -- 'system', 'admin:user@example.com', etc.
    
    -- Status
    is_current BOOLEAN DEFAULT true,                   -- only one current verification per capability
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_verifications_capability ON capability_verifications(capability_id);
CREATE INDEX idx_verifications_level ON capability_verifications(level);
CREATE INDEX idx_verifications_current ON capability_verifications(capability_id, is_current) WHERE is_current = true;

-- Ensure only one current verification per capability
CREATE UNIQUE INDEX idx_verifications_unique_current ON capability_verifications(capability_id) WHERE is_current = true;

-- Trigger to update capability updated_at
CREATE OR REPLACE FUNCTION update_capability_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER capability_updated
    BEFORE UPDATE ON capabilities
    FOR EACH ROW
    EXECUTE FUNCTION update_capability_timestamp();

-- Trigger to build domain_path from domain/type/subtype
CREATE OR REPLACE FUNCTION build_domain_path()
RETURNS TRIGGER AS $$
BEGIN
    NEW.domain_path = NEW.domain || 
        CASE WHEN NEW.type IS NOT NULL AND NEW.type != '' THEN '/' || NEW.type ELSE '' END ||
        CASE WHEN NEW.subtype IS NOT NULL AND NEW.subtype != '' THEN '/' || NEW.subtype ELSE '' END;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER capability_domain_path
    BEFORE INSERT OR UPDATE ON capabilities
    FOR EACH ROW
    EXECUTE FUNCTION build_domain_path();
