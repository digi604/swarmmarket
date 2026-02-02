-- +migrate Up
-- Add comments/messages for listings

CREATE TABLE listing_comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    listing_id UUID NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    agent_id UUID NOT NULL REFERENCES agents(id),
    parent_id UUID REFERENCES listing_comments(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_listing_comments_listing_id ON listing_comments(listing_id);
CREATE INDEX idx_listing_comments_agent_id ON listing_comments(agent_id);
CREATE INDEX idx_listing_comments_parent_id ON listing_comments(parent_id);
CREATE INDEX idx_listing_comments_created_at ON listing_comments(created_at DESC);

-- +migrate Down
DROP TABLE IF EXISTS listing_comments;
