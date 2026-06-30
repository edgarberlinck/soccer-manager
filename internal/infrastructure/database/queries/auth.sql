-- name: CreateUser :one
INSERT INTO users (
    id,
    username,
    password_hash,
    active,
    verification_token,
    verification_token_expires_at
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING id, username, password_hash, active, email_verified_at, verification_token, verification_token_expires_at, created_at, updated_at;

-- name: GetUserByEmail :one
SELECT id, username, password_hash, active, email_verified_at, verification_token, verification_token_expires_at, created_at, updated_at
FROM users
WHERE username = $1;

-- name: GetUserByID :one
SELECT id, username, password_hash, active, email_verified_at, verification_token, verification_token_expires_at, created_at, updated_at
FROM users
WHERE id = $1;

-- name: VerifyUserByToken :one
UPDATE users
SET
    active = true,
    email_verified_at = NOW(),
    verification_token = NULL,
    verification_token_expires_at = NULL,
    updated_at = NOW()
WHERE verification_token = $1
  AND verification_token_expires_at > NOW()
RETURNING id, username, password_hash, active, email_verified_at, verification_token, verification_token_expires_at, created_at, updated_at;
