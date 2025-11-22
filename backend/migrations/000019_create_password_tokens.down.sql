-- Drop password_tokens table
DROP INDEX IF EXISTS idx_password_tokens_expiry;
DROP INDEX IF EXISTS idx_password_tokens_user_id;
DROP TABLE IF EXISTS password_tokens;
