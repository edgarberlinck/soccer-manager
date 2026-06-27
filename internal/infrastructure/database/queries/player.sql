-- name: CreatePlayer :one
INSERT INTO players (
    id,
    name,
    age,
    pace,
    passing,
    shooting
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
    shooting = $6
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
    shooting = $4
WHERE id = $1
RETURNING *;

-- name: FindPlayersReadyToRetire :many
SELECT *
FROM players
WHERE age >= $1;