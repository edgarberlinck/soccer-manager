-- name: GetUserClubs :many
SELECT *
FROM clubs
WHERE user_id = $1;

-- name: GetClubByName :one
SELECT *
FROM clubs
WHERE LOWER(name) = LOWER($1)
LIMIT 1;

-- name: CreateClub :one
INSERT INTO clubs (
	id,
	user_id,
	name,
	short_name,
	abbreviation,
	continent,
	country
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
RETURNING *;