CREATE TABLE IF NOT EXISTS fraud_attempts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reason VARCHAR(100) NOT NULL,
    platform VARCHAR(20) NOT NULL CHECK (platform IN ('android', 'ios')),
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    accuracy DECIMAL(10, 2),
    attempted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_fraud_attempts_user_time
    ON fraud_attempts(user_id, attempted_at DESC);
