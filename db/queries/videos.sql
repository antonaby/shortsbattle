-- name: CreateVideo :one
INSERT INTO videos (video_url, oembed) 
VALUES ($1, $2) 
ON CONFLICT (video_url) DO UPDATE
  SET video_url  = EXCLUDED.video_url,
      updated_at = now()
RETURNING *;

-- name: AddVideoToPlayer :one  
INSERT INTO player_videos (player_id, video_id) 
VALUES ($1, $2) 
ON CONFLICT (player_id, video_id) DO UPDATE
  SET added_at = now()
RETURNING *;

-- name: GetVideo :one
SELECT * FROM videos WHERE id = $1;

-- name: GetVideosByPlayer :many
SELECT v.* FROM videos AS v
JOIN player_videos AS pv ON pv.video_id = v.id
WHERE pv.player_id = $1;

-- name: GetVideoByPlayerInGame :one
SELECT v.*
FROM videos AS v
JOIN player_videos AS pv
  ON pv.video_id = v.id
JOIN game_players AS gp
  ON gp.player_id = pv.player_id
 AND gp.game_id   = $1
WHERE v.id         = $2
  AND pv.player_id = $3
LIMIT 1;

-- name: AddVideoToGame :exec
INSERT INTO game_videos (game_id, player_id, video_id)
SELECT $1, $2, $3
FROM player_videos pv
WHERE pv.player_id = $2
  AND pv.video_id  = $3
ON CONFLICT (game_id, player_id)
DO UPDATE
SET video_id     = EXCLUDED.video_id,
    submitted_at = now();
    
-- name: GetVideosToWatch :many
SELECT v.*
FROM game_videos gv
JOIN player_videos pv
  ON pv.player_id = gv.player_id
 AND pv.video_id  = gv.video_id
JOIN videos v
  ON v.id = gv.video_id
WHERE gv.game_id = $1
  AND EXISTS (
    SELECT 1
    FROM game_players gp
    WHERE gp.game_id  = $1
      AND gp.player_id = $2
  )
ORDER BY gv.submitted_at;
