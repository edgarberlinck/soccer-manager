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