-- +goose Up
CREATE TABLE IF NOT EXISTS "match" (
    id UUID PRIMARY KEY,
    home_club_id UUID NOT NULL REFERENCES clubs(id),
    away_club_id UUID NOT NULL REFERENCES clubs(id),
    championship_id UUID,
    status TEXT NOT NULL DEFAULT 'in_progress',
    current_tick SMALLINT NOT NULL DEFAULT 1,
    random_seed BIGINT NOT NULL DEFAULT 1,
    home_score INTEGER NOT NULL DEFAULT 0,
    away_score INTEGER NOT NULL DEFAULT 0,
    finished_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT match_status_check CHECK (status IN ('in_progress', 'finished', 'cancelled')),
    CONSTRAINT match_tick_check CHECK (current_tick BETWEEN 1 AND 90),
    CONSTRAINT match_distinct_clubs_check CHECK (home_club_id <> away_club_id)
);

CREATE INDEX IF NOT EXISTS idx_match_status_tick ON "match" (status, current_tick);
CREATE INDEX IF NOT EXISTS idx_match_championship ON "match" (championship_id);

CREATE TABLE IF NOT EXISTS match_events (
    id BIGSERIAL PRIMARY KEY,
    match_id UUID NOT NULL REFERENCES "match"(id) ON DELETE CASCADE,
    tick SMALLINT NOT NULL,
    event_type TEXT NOT NULL,
    description TEXT NOT NULL,
    home_score INTEGER NOT NULL,
    away_score INTEGER NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT match_events_tick_check CHECK (tick BETWEEN 1 AND 90)
);

CREATE INDEX IF NOT EXISTS idx_match_events_match_tick ON match_events (match_id, tick);

CREATE TABLE IF NOT EXISTS match_results (
    match_id UUID PRIMARY KEY REFERENCES "match"(id) ON DELETE CASCADE,
    home_team_score INTEGER NOT NULL,
    away_team_score INTEGER NOT NULL,
    winner_club_id UUID REFERENCES clubs(id),
    loser_club_id UUID REFERENCES clubs(id),
    is_draw BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS match_results;
DROP TABLE IF EXISTS match_events;
DROP INDEX IF EXISTS idx_match_championship;
DROP INDEX IF EXISTS idx_match_status_tick;
DROP TABLE IF EXISTS "match";
