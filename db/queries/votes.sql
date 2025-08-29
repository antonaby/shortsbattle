-- name: VoteForVideo :one
INSERT INTO game_votes (game_id, player_id, video_id, value)
SELECT $1, $2, $3, $4
WHERE EXISTS (
  SELECT 1
  FROM game_players gp
  WHERE gp.game_id = $1 AND gp.player_id = $2
)
AND EXISTS (
  SELECT 1 FROM game_videos gv
  WHERE gv.game_id = $1 AND gv.video_id = $3
)
ON CONFLICT (game_id, player_id, video_id)
DO UPDATE 
SET value = EXCLUDED.value, 
    voted_at = now()
RETURNING *;
