-- name: GetUserClubs :many
SELECT * 
FROM clubs 
where user_id = $1;