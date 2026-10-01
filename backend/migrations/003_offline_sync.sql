ALTER TABLE attendances
    ADD COLUMN IF NOT EXISTS check_in_idempotency_key UUID UNIQUE,
    ADD COLUMN IF NOT EXISTS check_out_idempotency_key UUID UNIQUE;

CREATE INDEX IF NOT EXISTS idx_offline_queue_user_created
    ON offline_queue(user_id, created_at);
