-- name: CreatePlayer :one
INSERT INTO players (name) VALUES ($1)
RETURNING id, name;

-- name: GetPlayer :one
SELECT id, name FROM players WHERE id = $1;