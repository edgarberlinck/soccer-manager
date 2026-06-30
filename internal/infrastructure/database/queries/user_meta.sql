-- name: GetUserMetaByUserID :one
SELECT *
FROM user_meta
WHERE user_id = $1;

-- name: UpsertUserMeta :one
INSERT INTO user_meta (
    user_id,
    full_name,
    country,
    social_links,
    metadata
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
ON CONFLICT (user_id)
DO UPDATE SET
    full_name = EXCLUDED.full_name,
    country = EXCLUDED.country,
    social_links = EXCLUDED.social_links,
    metadata = EXCLUDED.metadata,
    updated_at = NOW()
RETURNING *;
