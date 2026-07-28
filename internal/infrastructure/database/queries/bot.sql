-- name: CreateBotUser :one
INSERT INTO users (
    id,
    username,
    password_hash,
    active,
    is_bot,
    bot_strategy
)
VALUES (
    $1,
    $2,
    $3,
    true,
    true,
    $4
)
RETURNING *;

-- name: GetAllBots :many
SELECT *
FROM users
WHERE is_bot = true;

-- name: GetBotsByStrategy :many
SELECT *
FROM users
WHERE is_bot = true AND bot_strategy = $1;

-- name: GetBotClubs :many
SELECT c.*
FROM clubs c
INNER JOIN users u ON c.user_id = u.id
WHERE u.is_bot = true;

-- name: IsClubOwnedByBot :one
SELECT u.is_bot
FROM clubs c
INNER JOIN users u ON c.user_id = u.id
WHERE c.id = $1
LIMIT 1;
