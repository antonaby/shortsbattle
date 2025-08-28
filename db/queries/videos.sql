-- name: CreateVideo :one
INSERT INTO videos (player_id, video_url, oembed) VALUES ($1, $2, $3) RETURNING *;

-- name: GetVideo :one
SELECT * FROM videos WHERE id = $1;

-- name: GetVideosByPlayer :many
SELECT * FROM videos WHERE player_id = $1;

-- name: GetVideoByPlayerInGame :one
SELECT v.*
FROM videos AS v
JOIN game_players AS gp
  ON gp.player_id = v.player_id 
 AND gp.game_id   = $1 
WHERE v.id         = $2
  AND v.player_id  = $3
LIMIT 1;