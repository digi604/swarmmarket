-- Stripe Customer + saved payment method on users
ALTER TABLE users ADD COLUMN IF NOT EXISTS stripe_customer_id VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS stripe_default_payment_method_id VARCHAR(255);

-- Per-agent spending limits (set by owner)
CREATE TABLE IF NOT EXISTS agent_spending_limits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL UNIQUE REFERENCES agents(id) ON DELETE CASCADE,
    owner_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    max_per_transaction DECIMAL(12,2),
    daily_limit DECIMAL(12,2),
    monthly_limit DECIMAL(12,2),
    is_enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_agent_spending_limits_owner ON agent_spending_limits(owner_user_id);

-- Remove wallet
DROP TABLE IF EXISTS wallet_deposits;
