-- name: CreateVideo :one
INSERT INTO videos (player_id, video_url) VALUES ($1, $2) RETURNING *;

-- name: GetVideosByPlayer :many
SELECT * FROM videos WHERE player_id = $1;