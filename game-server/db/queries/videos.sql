-- name: CreateVideo :one
INSERT INTO videos (game_id, player_id, video_url, is_actual)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: InvalidateOtherVideos :exec
UPDATE videos SET is_actual = FALSE WHERE game_id = $1 AND player_id = $2 AND id <> $3;

-- name: GetVideosByGame :many
SELECT id, game_id, player_id, video_url, is_actual, submitted_at
FROM videos
WHERE game_id = $1;

-- name: GetVideosByPlayer :one
SELECT id, game_id, player_id, video_url, submitted_at
FROM videos
WHERE game_id = $1 AND player_id = $2;