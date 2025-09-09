-- name: AddVideoToPlayer :one  
SELECT * FROM add_video_for_player(sqlc.arg(player_id), sqlc.arg(video_url), sqlc.arg(oembed));

-- name: GetVideosByPlayer :many
SELECT v.* 
FROM videos AS v
JOIN player_videos AS pv ON pv.video_id = v.id
WHERE pv.player_id = $1;

-- name: UpsertGameVideoIfOwned :one
INSERT INTO game_videos (game_id, player_id, video_id, round_n)
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
    JOIN rounds r
      ON r.theme_id = g.theme_id
    WHERE g.id = $1 AND r.round_n = $4
  )
ON CONFLICT ON CONSTRAINT unique_player_game_round
DO UPDATE
SET video_id     = EXCLUDED.video_id,
    submitted_at = now()
RETURNING *;

-- name: GetSubmittedVideosByPlayers :many
SELECT gp.game_id, gp.player_id, gv.id as game_video_id, gv.video_id, gv.round_n, gv.submitted_at
FROM game_players gp
LEFT JOIN game_videos gv
  ON gv.game_id = gp.game_id
  AND gv.player_id = gp.player_id
  AND gv.round_n = $2
WHERE gp.game_id = $1;
    
-- name: GetVideosToWatch :many
SELECT 
    gv.id as game_video_id,
    gv.game_id,
    gv.round_n,
    v.id AS video_id,
    v.video_url,
    v.oembed,
    v.added_at,
    v.updated_at
FROM game_videos gv
JOIN videos v ON gv.video_id = v.id
WHERE gv.game_id = $1
  AND gv.round_n = $2;