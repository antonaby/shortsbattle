-- name: CreatePlayer :one
INSERT INTO players (tg_id, tg_username, tg_language_code) VALUES ($1, $2, $3) RETURNING *;

-- name: GetPlayer :one
SELECT * FROM players WHERE id = $1;

-- name: GetPlayerByTgId :one
SELECT * FROM players WHERE tg_id = $1;
