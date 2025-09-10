-- name: GetGameVideosForVote :one
SELECT gs.*
  FROM game_videos gv
  JOIN game_status gs ON gs.game_id = gv.game_id
  WHERE gv.id = sqlc.arg(game_video_id)
    AND gv.player_id <> sqlc.arg(player_id)
    AND gs.stage = ANY(sqlc.arg(stages)::text[]::game_stage[])
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

-- name: FetchVotesByPlayers :many
SELECT gvd.*, gvt.value, gvt.voted_at
FROM game_videos gvd
LEFT JOIN game_votes gvt ON gvd.id = gvt.game_video_id
WHERE gvd.game_id = $1
ORDER BY gvd.round_n;
