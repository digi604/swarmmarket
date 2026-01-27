-- SwarmMarket Initial Schema
-- Migration: 001_initial_schema

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Agents table
CREATE TABLE agents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    owner_email VARCHAR(255) NOT NULL,
    api_key_hash VARCHAR(64) NOT NULL UNIQUE,
    api_key_prefix VARCHAR(16) NOT NULL,
    verification_level VARCHAR(20) NOT NULL DEFAULT 'basic',
    trust_score DECIMAL(5, 4) NOT NULL DEFAULT 0.5,
    total_transactions INTEGER NOT NULL DEFAULT 0,
    successful_trades INTEGER NOT NULL DEFAULT 0,
    average_rating DECIMAL(3, 2) NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_agents_api_key_hash ON agents(api_key_hash);
CREATE INDEX idx_agents_owner_email ON agents(owner_email);
CREATE INDEX idx_agents_verification_level ON agents(verification_level);
CREATE INDEX idx_agents_is_active ON agents(is_active);
CREATE INDEX idx_agents_trust_score ON agents(trust_score DESC);

-- Categories table (hierarchical taxonomy)
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    parent_id UUID REFERENCES categories(id),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_categories_parent_id ON categories(parent_id);
CREATE INDEX idx_categories_slug ON categories(slug);

-- Listings table (items/services/data for sale)
CREATE TABLE listings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    seller_id UUID NOT NULL REFERENCES agents(id),
    category_id UUID REFERENCES categories(id),
    title VARCHAR(500) NOT NULL,
    description TEXT,
    listing_type VARCHAR(20) NOT NULL, -- 'goods', 'services', 'data'
    price_amount DECIMAL(20, 8),
    price_currency VARCHAR(10) DEFAULT 'USD',
    quantity INTEGER DEFAULT 1,
    geographic_scope VARCHAR(20) DEFAULT 'international', -- 'local', 'regional', 'national', 'international'
    location_lat DECIMAL(10, 8),
    location_lng DECIMAL(11, 8),
    location_radius_km INTEGER,
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- 'draft', 'active', 'paused', 'sold', 'expired'
    expires_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_listings_seller_id ON listings(seller_id);
CREATE INDEX idx_listings_category_id ON listings(category_id);
CREATE INDEX idx_listings_status ON listings(status);
CREATE INDEX idx_listings_listing_type ON listings(listing_type);
CREATE INDEX idx_listings_geographic_scope ON listings(geographic_scope);
CREATE INDEX idx_listings_created_at ON listings(created_at DESC);

-- Requests table (what agents are looking for - reverse auction)
CREATE TABLE requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    requester_id UUID NOT NULL REFERENCES agents(id),
    category_id UUID REFERENCES categories(id),
    title VARCHAR(500) NOT NULL,
    description TEXT,
    request_type VARCHAR(20) NOT NULL, -- 'goods', 'services', 'data'
    budget_min DECIMAL(20, 8),
    budget_max DECIMAL(20, 8),
    budget_currency VARCHAR(10) DEFAULT 'USD',
    quantity INTEGER DEFAULT 1,
    geographic_scope VARCHAR(20) DEFAULT 'international',
    location_lat DECIMAL(10, 8),
    location_lng DECIMAL(11, 8),
    location_radius_km INTEGER,
    status VARCHAR(20) NOT NULL DEFAULT 'open', -- 'open', 'in_progress', 'fulfilled', 'cancelled', 'expired'
    expires_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_requests_requester_id ON requests(requester_id);
CREATE INDEX idx_requests_category_id ON requests(category_id);
CREATE INDEX idx_requests_status ON requests(status);
CREATE INDEX idx_requests_request_type ON requests(request_type);
CREATE INDEX idx_requests_created_at ON requests(created_at DESC);

-- Offers table (responses to requests)
CREATE TABLE offers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    request_id UUID NOT NULL REFERENCES requests(id),
    offerer_id UUID NOT NULL REFERENCES agents(id),
    listing_id UUID REFERENCES listings(id), -- Optional link to existing listing
    price_amount DECIMAL(20, 8) NOT NULL,
    price_currency VARCHAR(10) DEFAULT 'USD',
    description TEXT,
    delivery_terms TEXT,
    valid_until TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'accepted', 'rejected', 'withdrawn', 'expired'
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_offers_request_id ON offers(request_id);
CREATE INDEX idx_offers_offerer_id ON offers(offerer_id);
CREATE INDEX idx_offers_status ON offers(status);

-- Auctions table
CREATE TABLE auctions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    listing_id UUID REFERENCES listings(id),
    seller_id UUID NOT NULL REFERENCES agents(id),
    auction_type VARCHAR(20) NOT NULL, -- 'english', 'dutch', 'sealed', 'continuous'
    title VARCHAR(500) NOT NULL,
    description TEXT,
    starting_price DECIMAL(20, 8) NOT NULL,
    current_price DECIMAL(20, 8),
    reserve_price DECIMAL(20, 8),
    buy_now_price DECIMAL(20, 8),
    price_currency VARCHAR(10) DEFAULT 'USD',
    min_increment DECIMAL(20, 8), -- For english auctions
    price_decrement DECIMAL(20, 8), -- For dutch auctions
    decrement_interval_seconds INTEGER, -- For dutch auctions
    status VARCHAR(20) NOT NULL DEFAULT 'scheduled', -- 'scheduled', 'active', 'ended', 'cancelled'
    starts_at TIMESTAMP WITH TIME ZONE NOT NULL,
    ends_at TIMESTAMP WITH TIME ZONE NOT NULL,
    extension_seconds INTEGER DEFAULT 60, -- Anti-sniping extension
    winning_bid_id UUID,
    winner_id UUID REFERENCES agents(id),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_auctions_seller_id ON auctions(seller_id);
CREATE INDEX idx_auctions_auction_type ON auctions(auction_type);
CREATE INDEX idx_auctions_status ON auctions(status);
CREATE INDEX idx_auctions_starts_at ON auctions(starts_at);
CREATE INDEX idx_auctions_ends_at ON auctions(ends_at);

-- Bids table
CREATE TABLE bids (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    auction_id UUID NOT NULL REFERENCES auctions(id),
    bidder_id UUID NOT NULL REFERENCES agents(id),
    amount DECIMAL(20, 8) NOT NULL,
    currency VARCHAR(10) DEFAULT 'USD',
    is_sealed BOOLEAN DEFAULT false, -- For sealed-bid auctions
    is_winning BOOLEAN DEFAULT false,
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- 'active', 'outbid', 'winning', 'cancelled'
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bids_auction_id ON bids(auction_id);
CREATE INDEX idx_bids_bidder_id ON bids(bidder_id);
CREATE INDEX idx_bids_amount ON bids(amount DESC);
CREATE INDEX idx_bids_status ON bids(status);

-- Add foreign key for winning bid
ALTER TABLE auctions ADD CONSTRAINT fk_winning_bid FOREIGN KEY (winning_bid_id) REFERENCES bids(id);

-- Transactions table (completed trades)
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    buyer_id UUID NOT NULL REFERENCES agents(id),
    seller_id UUID NOT NULL REFERENCES agents(id),
    listing_id UUID REFERENCES listings(id),
    request_id UUID REFERENCES requests(id),
    offer_id UUID REFERENCES offers(id),
    auction_id UUID REFERENCES auctions(id),
    amount DECIMAL(20, 8) NOT NULL,
    currency VARCHAR(10) DEFAULT 'USD',
    platform_fee DECIMAL(20, 8) DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'escrow_funded', 'delivered', 'completed', 'disputed', 'refunded', 'cancelled'
    delivery_confirmed_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_buyer_id ON transactions(buyer_id);
CREATE INDEX idx_transactions_seller_id ON transactions(seller_id);
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_transactions_created_at ON transactions(created_at DESC);

-- Escrow accounts table
CREATE TABLE escrow_accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    transaction_id UUID NOT NULL REFERENCES transactions(id),
    amount DECIMAL(20, 8) NOT NULL,
    currency VARCHAR(10) DEFAULT 'USD',
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'funded', 'released', 'refunded', 'disputed'
    funded_at TIMESTAMP WITH TIME ZONE,
    released_at TIMESTAMP WITH TIME ZONE,
    stripe_payment_intent_id VARCHAR(255),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_escrow_transaction_id ON escrow_accounts(transaction_id);
CREATE INDEX idx_escrow_status ON escrow_accounts(status);

-- Webhooks table
CREATE TABLE webhooks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id UUID NOT NULL REFERENCES agents(id),
    url VARCHAR(2048) NOT NULL,
    secret VARCHAR(64) NOT NULL,
    events TEXT[] NOT NULL, -- Array of event types to subscribe to
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_triggered_at TIMESTAMP WITH TIME ZONE,
    failure_count INTEGER DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webhooks_agent_id ON webhooks(agent_id);
CREATE INDEX idx_webhooks_is_active ON webhooks(is_active);

-- Ratings table
CREATE TABLE ratings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    transaction_id UUID NOT NULL REFERENCES transactions(id),
    rater_id UUID NOT NULL REFERENCES agents(id),
    rated_agent_id UUID NOT NULL REFERENCES agents(id),
    score INTEGER NOT NULL CHECK (score >= 1 AND score <= 5),
    comment TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(transaction_id, rater_id)
);

CREATE INDEX idx_ratings_transaction_id ON ratings(transaction_id);
CREATE INDEX idx_ratings_rated_agent_id ON ratings(rated_agent_id);
CREATE INDEX idx_ratings_score ON ratings(score);

-- Events table (audit log)
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_type VARCHAR(100) NOT NULL,
    agent_id UUID REFERENCES agents(id),
    resource_type VARCHAR(50),
    resource_id UUID,
    payload JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_events_event_type ON events(event_type);
CREATE INDEX idx_events_agent_id ON events(agent_id);
CREATE INDEX idx_events_resource ON events(resource_type, resource_id);
CREATE INDEX idx_events_created_at ON events(created_at DESC);

-- Seed default categories
INSERT INTO categories (name, slug, description) VALUES
    ('Goods', 'goods', 'Physical or virtual goods'),
    ('Services', 'services', 'Services provided by agents'),
    ('Data', 'data', 'Data sets, feeds, and information');

-- Add subcategories for Goods
INSERT INTO categories (parent_id, name, slug, description) VALUES
    ((SELECT id FROM categories WHERE slug = 'goods'), 'Digital Products', 'goods-digital', 'Digital goods and software'),
    ((SELECT id FROM categories WHERE slug = 'goods'), 'Physical Products', 'goods-physical', 'Physical goods requiring shipping');

-- Add subcategories for Services
INSERT INTO categories (parent_id, name, slug, description) VALUES
    ((SELECT id FROM categories WHERE slug = 'services'), 'Computation', 'services-computation', 'Computing and processing services'),
    ((SELECT id FROM categories WHERE slug = 'services'), 'Analysis', 'services-analysis', 'Data analysis and insights'),
    ((SELECT id FROM categories WHERE slug = 'services'), 'Creation', 'services-creation', 'Content and asset creation');

-- Add subcategories for Data
INSERT INTO categories (parent_id, name, slug, description) VALUES
    ((SELECT id FROM categories WHERE slug = 'data'), 'Real-time Feeds', 'data-realtime', 'Live data streams'),
    ((SELECT id FROM categories WHERE slug = 'data'), 'Datasets', 'data-datasets', 'Static or periodically updated datasets'),
    ((SELECT id FROM categories WHERE slug = 'data'), 'APIs', 'data-apis', 'API access to data sources');
