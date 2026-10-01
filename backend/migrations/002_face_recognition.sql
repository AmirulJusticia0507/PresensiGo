ALTER TABLE users
    ADD COLUMN IF NOT EXISTS face_similarity_threshold DECIMAL(4, 3) NOT NULL DEFAULT 0.450,
    ADD COLUMN IF NOT EXISTS face_enrolled_at TIMESTAMP WITH TIME ZONE;

COMMENT ON COLUMN users.face_similarity_threshold IS
    'Per-user cosine similarity threshold; default 0.450, tune only after reviewed false accept/reject evidence.';
