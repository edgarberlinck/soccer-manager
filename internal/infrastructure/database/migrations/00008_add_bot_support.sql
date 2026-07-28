-- +goose Up
ALTER TABLE users
ADD COLUMN IF NOT EXISTS is_bot BOOLEAN NOT NULL DEFAULT false,
ADD COLUMN IF NOT EXISTS bot_strategy TEXT CHECK (bot_strategy IN ('conservador', 'equilibrado', 'agressivo'));

CREATE INDEX IF NOT EXISTS users_is_bot_idx ON users(is_bot) WHERE is_bot = true;

-- +goose Down
DROP INDEX IF EXISTS users_is_bot_idx;

ALTER TABLE users
DROP COLUMN IF EXISTS bot_strategy,
DROP COLUMN IF EXISTS is_bot;
