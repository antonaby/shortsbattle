-- name: AddVideoToPlayer :one  
SELECT * FROM add_video_for_player($1, $2, $3);

-- name: GetVideo :one
SELECT * FROM videos WHERE id = $1;

-- name: GetVideosByPlayer :many
SELECT v.* 
FROM videos AS v
JOIN player_videos AS pv ON pv.video_id = v.id
WHERE pv.player_id = $1;

-- name: UpsertGameVideoIfOwned :one
INSERT INTO game_videos (game_id, player_id, video_id, request_id)
SELECT $1, $2, $3, $4
WHERE
  -- player owns the video
  EXISTS (
    SELECT 1
    FROM player_videos pv
    WHERE pv.player_id = $2
      AND pv.video_id  = $3
  )
  -- request and game share the same theme
  AND EXISTS (
    SELECT 1
    FROM games g
    JOIN video_requests vr
      ON vr.id = $4
    AND vr.theme_id = g.theme_id
    WHERE g.id = $1
  )
ON CONFLICT ON CONSTRAINT unique_player_game_request
DO UPDATE
SET video_id     = EXCLUDED.video_id,
    submitted_at = now()
RETURNING *;
    
-- name: GetVideosToWatch :many
SELECT v.*
FROM game_videos gv
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
