ALTER TABLE users ADD COLUMN stripe_connect_account_id VARCHAR(255);
ALTER TABLE users ADD COLUMN stripe_connect_charges_enabled BOOLEAN DEFAULT FALSE;
