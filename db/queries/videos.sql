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

-- name: AddVideoByPlayerInGame :one
INSERT INTO videos (player_id, video_url, oembed)
SELECT gp.player_id, $3, $4
FROM game_players gp
WHERE gp.game_id  = $1
  AND gp.player_id = $2
RETURNING *;

-- name: AddVideoToGame :exec
INSERT INTO game_videos (game_id, player_id, video_id) 
VALUES ($1, $2, $3) 
ON CONFLICT (game_id, player_id)
DO UPDATE
SET video_id = EXCLUDED.video_id,
    submitted_at = now();
    
-- name: GetVideosToWatch :many
SELECT v.*
FROM game_videos gv
JOIN videos v ON v.id = gv.video_id
WHERE gv.game_id = $1
  AND EXISTS (
    SELECT 1
    FROM game_players gp
    WHERE gp.game_id = $1
      AND gp.player_id = $2
  )
ORDER BY gv.submitted_at;
