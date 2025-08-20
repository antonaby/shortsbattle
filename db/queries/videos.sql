-- name: CreateVideo :one
INSERT INTO videos (player_id, video_url, oembed) VALUES ($1, $2, $3) RETURNING *;

-- name: GetVideo :one
SELECT * FROM videos WHERE id = $1;

-- name: GetVideosByPlayer :many
SELECT * FROM videos WHERE player_id = $1;