-- name: CreateMatch :one
INSERT INTO "match" (
    id,
    home_club_id,
    away_club_id,
    championship_id,
    random_seed
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING id, home_club_id, away_club_id, championship_id, status, current_tick, random_seed, home_score, away_score, finished_at, created_at, updated_at;

-- name: ListInProgressMatches :many
SELECT id, home_club_id, away_club_id, championship_id, status, current_tick, random_seed, home_score, away_score, finished_at, created_at, updated_at
FROM "match"
WHERE status = 'in_progress'
ORDER BY updated_at ASC
LIMIT $1;

-- name: CreateMatchEvent :one
INSERT INTO match_events (
    match_id,
    tick,
    event_type,
    description,
    home_score,
    away_score,
    payload
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
)
RETURNING id, match_id, tick, event_type, description, home_score, away_score, payload, created_at;

-- name: UpdateMatchScoreAndTick :exec
UPDATE "match"
SET
    home_score = $2,
    away_score = $3,
    current_tick = $4,
    updated_at = NOW()
WHERE id = $1
  AND status = 'in_progress';

-- name: FinishMatch :exec
UPDATE "match"
SET
    status = 'finished',
    finished_at = NOW(),
    updated_at = NOW()
WHERE id = $1
  AND status = 'in_progress';

-- name: UpsertMatchResult :one
INSERT INTO match_results (
    match_id,
    home_team_score,
    away_team_score,
    winner_club_id,
    loser_club_id,
    is_draw
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
ON CONFLICT (match_id)
DO UPDATE SET
    home_team_score = EXCLUDED.home_team_score,
    away_team_score = EXCLUDED.away_team_score,
    winner_club_id = EXCLUDED.winner_club_id,
    loser_club_id = EXCLUDED.loser_club_id,
    is_draw = EXCLUDED.is_draw
RETURNING match_id, home_team_score, away_team_score, winner_club_id, loser_club_id, is_draw, created_at;

-- name: GetRandomBotClub :one
SELECT c.*
FROM clubs c
INNER JOIN users u ON c.user_id = u.id
WHERE u.is_bot = true
ORDER BY RANDOM()
LIMIT 1;

-- name: GetPendingMatches :many
SELECT *
FROM "match"
WHERE status = 'pending'
ORDER BY created_at ASC
LIMIT $1;

-- name: GetClubMatches :many
SELECT *
FROM "match"
WHERE (home_club_id = $1 OR away_club_id = $1)
ORDER BY created_at ASC;

-- name: GetSeasonMatches :many
SELECT *
FROM "match"
WHERE created_at >= $1 AND created_at <= $2
ORDER BY created_at ASC;
