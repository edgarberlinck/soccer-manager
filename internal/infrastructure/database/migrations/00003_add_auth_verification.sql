-- +goose Up
UPDATE users
SET active = false
WHERE active IS NULL;

ALTER TABLE users
ALTER COLUMN active SET NOT NULL,
ALTER COLUMN active SET DEFAULT false;

ALTER TABLE users
ADD COLUMN IF NOT EXISTS email_verified_at TIMESTAMPTZ,
ADD COLUMN IF NOT EXISTS verification_token TEXT,
ADD COLUMN IF NOT EXISTS verification_token_expires_at TIMESTAMPTZ,
ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

CREATE UNIQUE INDEX IF NOT EXISTS users_verification_token_unique
ON users(verification_token)
WHERE verification_token IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS users_verification_token_unique;

ALTER TABLE users
DROP COLUMN IF EXISTS updated_at,
DROP COLUMN IF EXISTS created_at,
DROP COLUMN IF EXISTS verification_token_expires_at,
DROP COLUMN IF EXISTS verification_token,
DROP COLUMN IF EXISTS email_verified_at;

ALTER TABLE users
ALTER COLUMN active DROP NOT NULL,
ALTER COLUMN active DROP DEFAULT;
