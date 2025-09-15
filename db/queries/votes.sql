-- name: GetGameVideoShareLock :one
SELECT 
  gs.*,
  gv.id as game_video_id,
  gv.player_id,
  gv.video_id,
  gv.submitted_at
  FROM game_videos gv
  JOIN game_status gs ON gs.game_id = gv.game_id
  WHERE gv.id = sqlc.arg(game_video_id)
    AND gv.player_id <> sqlc.arg(player_id)
    AND EXISTS (
      SELECT 1
      FROM game_players gp
      WHERE gp.game_id = gv.game_id
        AND gp.player_id = sqlc.arg(player_id)
    ) 
FOR SHARE OF gs;

-- name: VoteForVideo :one
INSERT INTO game_votes (game_video_id, player_id, value)
VALUES ($1, $2, $3)
ON CONFLICT (game_video_id, player_id)
DO UPDATE 
SET value = EXCLUDED.value, 
    voted_at = now()
RETURNING *;

-- name: GetVotes :many
WITH players AS (
  SELECT gp.player_id
  FROM game_players gp
  WHERE gp.game_id = $1
),
videos AS (
  SELECT gv.id AS game_video_id, gv.player_id AS author_id
  FROM game_videos gv
  WHERE gv.game_id = $1
    AND gv.round_n = $2
),
expected AS (
  SELECT p.player_id, v.game_video_id
  FROM players p
  CROSS JOIN videos v
  WHERE p.player_id <> v.author_id
)
SELECT 
  e.player_id,
  e.game_video_id,
  gvt.value,
  gvt.voted_at
  FROM expected e
  LEFT JOIN game_votes gvt
    ON gvt.game_video_id = e.game_video_id
   AND gvt.player_id     = e.player_id;
