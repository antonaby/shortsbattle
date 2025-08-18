-- name: CreateGame :one
INSERT INTO games (theme_id, status)
VALUES ($1, $2)
RETURNING *;

-- name: GetGame :one
SELECT * FROM games WHERE id = $1;

-- name: GetAllGames :many
SELECT * FROM games ORDER BY created_at DESC;

-- name: UpdateGameStatus :one
UPDATE games SET status = $1 WHERE id = $2 RETURNING *;

-- name: DeleteGame :exec
DELETE FROM games WHERE id = $1;