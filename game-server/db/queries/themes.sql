-- name: CreateTheme :one
INSERT INTO themes (name, description) VALUES ($1, $2) RETURNING *;

-- name: CreateVideoRequest :one
INSERT INTO video_requests (request, theme_id) VALUES ($1, $2) RETURNING *;

-- name: GetTheme :one
SELECT * FROM themes WHERE id = $1;

-- name: GetVideoRequests :many
SELECT * FROM video_requests WHERE theme_id = $1;

-- name: ListThemes :many
SELECT * FROM themes ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: ListAllThemes :many
SELECT * FROM themes ORDER BY created_at;
