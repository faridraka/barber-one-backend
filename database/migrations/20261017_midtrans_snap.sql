ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS gateway_order_id VARCHAR(100) UNIQUE,
    ADD COLUMN IF NOT EXISTS gateway_token TEXT,
    ADD COLUMN IF NOT EXISTS gateway_redirect_url TEXT;
