-- +migrate Up
-- Wallet Deposits Table (supports both human users and agents)

CREATE TABLE wallet_deposits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    agent_id UUID REFERENCES agents(id) ON DELETE CASCADE,
    amount DECIMAL(12,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    stripe_payment_intent_id VARCHAR(255),
    stripe_client_secret VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',  -- 'pending', 'processing', 'completed', 'failed', 'cancelled'
    failure_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT wallet_deposits_owner_check CHECK (
        (user_id IS NOT NULL AND agent_id IS NULL) OR
        (user_id IS NULL AND agent_id IS NOT NULL)
    )
);

CREATE INDEX idx_wallet_deposits_user_id ON wallet_deposits(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_wallet_deposits_agent_id ON wallet_deposits(agent_id) WHERE agent_id IS NOT NULL;
CREATE INDEX idx_wallet_deposits_status ON wallet_deposits(status);
CREATE INDEX idx_wallet_deposits_stripe_pi ON wallet_deposits(stripe_payment_intent_id);
CREATE INDEX idx_wallet_deposits_created ON wallet_deposits(created_at DESC);

-- +migrate Down
DROP TABLE IF EXISTS wallet_deposits;
