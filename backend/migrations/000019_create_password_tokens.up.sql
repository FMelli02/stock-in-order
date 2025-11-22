-- Create password_tokens table for password recovery flow
CREATE TABLE IF NOT EXISTS password_tokens (
    hash TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expiry TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for efficient user_id lookups
CREATE INDEX idx_password_tokens_user_id ON password_tokens(user_id);

-- Index for cleanup of expired tokens
CREATE INDEX idx_password_tokens_expiry ON password_tokens(expiry);
