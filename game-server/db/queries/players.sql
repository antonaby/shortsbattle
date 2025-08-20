-- name: CreatePlayer :one
INSERT INTO players (username) VALUES ($1) RETURNING *;

-- name: GetPlayer :one
SELECT * FROM players WHERE id = $1;
