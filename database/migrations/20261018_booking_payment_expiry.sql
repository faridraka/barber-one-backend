ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ,
    ALTER COLUMN paid_at DROP NOT NULL;

CREATE INDEX IF NOT EXISTS idx_transactions_pending_expiry
    ON transactions(expires_at)
    WHERE payment_status = 'Menunggu' AND expires_at IS NOT NULL;
