-- Migration 004: Add profile fields to users table
-- This migration adds optional profile fields to support enhanced user management
-- All fields are nullable to maintain backward compatibility

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS phone VARCHAR(20),
    ADD COLUMN IF NOT EXISTS emergency_contact_name VARCHAR(100),
    ADD COLUMN IF NOT EXISTS emergency_contact_phone VARCHAR(20),
    ADD COLUMN IF NOT EXISTS address TEXT,
    ADD COLUMN IF NOT EXISTS profile_picture_url VARCHAR(500),
    ADD COLUMN IF NOT EXISTS terms_accepted_at TIMESTAMP WITH TIME ZONE;

-- Add comments explaining each new field
COMMENT ON COLUMN users.phone IS
    'User phone number. Optional field for contact purposes.';

COMMENT ON COLUMN users.emergency_contact_name IS
    'Full name of emergency contact person. Optional field.';

COMMENT ON COLUMN users.emergency_contact_phone IS
    'Phone number of emergency contact. Optional field.';

COMMENT ON COLUMN users.address IS
    'User address or residence. Optional field for administrative purposes.';

COMMENT ON COLUMN users.profile_picture_url IS
    'URL to user profile picture hosted on cloud storage. Optional field.';

COMMENT ON COLUMN users.terms_accepted_at IS
    'Timestamp when user accepted terms and conditions. Null until explicitly accepted.';

-- Create email index if it does not exist
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Create index on email for faster lookups (if unique index already exists, this is redundant)
-- but kept for clarity in query optimization
CREATE INDEX IF NOT EXISTS idx_users_email_lookup ON users(email) WHERE email IS NOT NULL;
