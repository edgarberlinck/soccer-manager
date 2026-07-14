-- +goose Up
ALTER TABLE players
    ADD COLUMN IF NOT EXISTS club_id UUID REFERENCES clubs(id) ON DELETE CASCADE,
    ADD COLUMN IF NOT EXISTS position TEXT NOT NULL DEFAULT 'MF',
    ADD COLUMN IF NOT EXISTS overall SMALLINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS potential SMALLINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS tier TEXT NOT NULL DEFAULT 'internal';

CREATE INDEX IF NOT EXISTS idx_players_club_id ON players (club_id);

CREATE TABLE IF NOT EXISTS player_contracts (
    id UUID PRIMARY KEY,
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    salary_cents BIGINT NOT NULL,
    release_clause_cents BIGINT,
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT player_contract_period_check CHECK (ends_at > starts_at),
    CONSTRAINT player_contract_salary_check CHECK (salary_cents >= 0)
);

CREATE INDEX IF NOT EXISTS idx_player_contracts_player_id ON player_contracts (player_id);
CREATE INDEX IF NOT EXISTS idx_player_contracts_period ON player_contracts (starts_at, ends_at);

CREATE TABLE IF NOT EXISTS player_match_stats (
    match_id UUID NOT NULL REFERENCES "match"(id) ON DELETE CASCADE,
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    club_id UUID NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    minutes_played SMALLINT NOT NULL DEFAULT 0,
    goals SMALLINT NOT NULL DEFAULT 0,
    assists SMALLINT NOT NULL DEFAULT 0,
    rating NUMERIC(4,2) NOT NULL DEFAULT 0,
    passes_completed SMALLINT NOT NULL DEFAULT 0,
    shots SMALLINT NOT NULL DEFAULT 0,
    tackles SMALLINT NOT NULL DEFAULT 0,
    saves SMALLINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (match_id, player_id),
    CONSTRAINT player_match_minutes_check CHECK (minutes_played BETWEEN 0 AND 130),
    CONSTRAINT player_match_rating_check CHECK (rating >= 0 AND rating <= 10)
);

CREATE INDEX IF NOT EXISTS idx_player_match_stats_player_id ON player_match_stats (player_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_player_match_stats_club_id ON player_match_stats (club_id);

-- +goose Down
DROP INDEX IF EXISTS idx_player_match_stats_club_id;
DROP INDEX IF EXISTS idx_player_match_stats_player_id;
DROP TABLE IF EXISTS player_match_stats;

DROP INDEX IF EXISTS idx_player_contracts_period;
DROP INDEX IF EXISTS idx_player_contracts_player_id;
DROP TABLE IF EXISTS player_contracts;

DROP INDEX IF EXISTS idx_players_club_id;

ALTER TABLE players
    DROP COLUMN IF EXISTS tier,
    DROP COLUMN IF EXISTS potential,
    DROP COLUMN IF EXISTS overall,
    DROP COLUMN IF EXISTS position,
    DROP COLUMN IF EXISTS club_id;