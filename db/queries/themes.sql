-- name: CreateTheme :one
INSERT INTO themes (title, description, mode) VALUES ($1, $2, $3) RETURNING *;

-- name: CreateRound :one
INSERT INTO rounds (round_n, title, description, theme_id) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetTheme :one
SELECT * FROM themes WHERE id = $1;

-- name: GetRounds :many
SELECT * FROM rounds WHERE theme_id = $1 ORDER BY round_n;

-- name: ListThemes :many
SELECT * FROM themes ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: ListAllThemes :many
SELECT * FROM themes ORDER BY created_at;
