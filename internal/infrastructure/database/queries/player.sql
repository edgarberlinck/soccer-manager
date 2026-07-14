-- name: CreatePlayer :one
INSERT INTO players (
    id,
    name,
    age,
    pace,
    passing,
    shooting,
    altura,
    peso,
    impulso,
    explosao,
    fisico,
    fisical_status,
    cabeceio,
    cruzamento,
    habilidade,
    finalizacao,
    dominio,
    temperamento
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10,
    $11,
    $12,
    $13,
    $14,
    $15,
    $16,
    $17,
    $18
)
RETURNING *;

-- name: CreateClubPlayer :one
INSERT INTO players (
    id,
    club_id,
    name,
    age,
    position,
    overall,
    potential,
    tier,
    pace,
    passing,
    shooting,
    altura,
    peso,
    impulso,
    explosao,
    fisico,
    fisical_status,
    cabeceio,
    cruzamento,
    habilidade,
    finalizacao,
    dominio,
    temperamento
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10,
    $11,
    $12,
    $13,
    $14,
    $15,
    $16,
    $17,
    $18,
    $19,
    $20,
    $21,
    $22,
    $23
)
RETURNING *;

-- name: GetPlayer :one
SELECT *
FROM players
WHERE id = $1;

-- name: ListPlayers :many
SELECT *
FROM players;

-- name: UpdatePlayer :one
UPDATE players
SET
    name = $2,
    age = $3,
    pace = $4,
    passing = $5,
    shooting = $6,
    altura = $7,
    peso = $8,
    impulso = $9,
    explosao = $10,
    fisico = $11,
    fisical_status = $12,
    cabeceio = $13,
    cruzamento = $14,
    habilidade = $15,
    finalizacao = $16,
    dominio = $17,
    temperamento = $18
WHERE id = $1
RETURNING *;

-- name: IncreasePlayerAge :exec
UPDATE players
SET age = age + 1
WHERE id = $1;

-- name: UpdatePlayerAttributes :one
UPDATE players
SET
    pace = $2,
    passing = $3,
    shooting = $4,
    altura = $5,
    peso = $6,
    impulso = $7,
    explosao = $8,
    fisico = $9,
    fisical_status = $10,
    cabeceio = $11,
    cruzamento = $12,
    habilidade = $13,
    finalizacao = $14,
    dominio = $15,
    temperamento = $16
WHERE id = $1
RETURNING *;

-- name: FindPlayersReadyToRetire :many
SELECT *
FROM players
WHERE age >= $1;

-- name: CountPlayersByClubID :one
SELECT COUNT(*)::BIGINT
FROM players
WHERE club_id = $1;

-- name: ListPlayersByClubID :many
SELECT
        p.id,
        p.club_id,
        p.name,
        p.age,
        p.position,
        p.overall,
        p.potential,
        p.pace,
        p.passing,
        p.shooting,
        c.salary_cents,
        c.ends_at
FROM players p
JOIN LATERAL (
        SELECT salary_cents, ends_at
        FROM player_contracts
        WHERE player_id = p.id
            AND starts_at <= NOW()
            AND ends_at >= NOW()
        ORDER BY starts_at DESC
        LIMIT 1
) c ON TRUE
WHERE p.club_id = $1
ORDER BY p.overall DESC, p.name ASC;

-- name: GetPlayerByClubIDAndID :one
SELECT *
FROM players
WHERE club_id = $1
    AND id = $2
LIMIT 1;

-- name: CreatePlayerContract :one
INSERT INTO player_contracts (
        id,
        player_id,
        salary_cents,
        release_clause_cents,
        starts_at,
        ends_at
)
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6
)
RETURNING *;

-- name: GetActivePlayerContract :one
SELECT *
FROM player_contracts
WHERE player_id = $1
    AND starts_at <= NOW()
    AND ends_at >= NOW()
ORDER BY starts_at DESC
LIMIT 1;

-- name: GetPlayerPerformanceSummary :one
SELECT
        COUNT(*)::BIGINT AS games,
        COALESCE(SUM(goals), 0)::BIGINT AS goals,
        COALESCE(SUM(assists), 0)::BIGINT AS assists,
        COALESCE(AVG(rating), 0)::NUMERIC(4,2) AS avg_rating,
        COALESCE(SUM(minutes_played), 0)::BIGINT AS minutes_played
FROM player_match_stats
WHERE player_id = $1;

-- name: ListPlayerMatchStats :many
SELECT
        pms.match_id,
        pms.player_id,
        pms.club_id,
        pms.minutes_played,
        pms.goals,
        pms.assists,
        pms.rating,
        pms.passes_completed,
        pms.shots,
        pms.tackles,
        pms.saves,
        pms.created_at,
        m.home_club_id,
        m.away_club_id,
        m.home_score,
        m.away_score,
        m.finished_at
FROM player_match_stats pms
JOIN "match" m ON m.id = pms.match_id
WHERE pms.player_id = $1
ORDER BY pms.created_at DESC
LIMIT $2;