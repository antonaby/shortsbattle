-- name: CreateVideo :one
INSERT INTO videos (game_id, player_id, video_url)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetVideosByGame :many
SELECT id, game_id, player_id, video_url, submitted_at
FROM videos
WHERE game_id = $1;

-- name: GetVideosByPlayer :one
SELECT id, game_id, player_id, video_url, submitted_at
FROM videos
WHERE game_id = $1 AND player_id = $2;