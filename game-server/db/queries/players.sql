-- name: CreatePlayer :one
INSERT INTO players (username)
VALUES ($1)
RETURNING *;

-- name: GetPlayer :one
SELECT id, username, created_at FROM players WHERE id = $1;

-- name: GetAllPlayers :many
SELECT id, username, created_at FROM players ORDER BY username;