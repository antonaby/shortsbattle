-- name: GetGameVideosForVote :one
SELECT g.*
  FROM game_videos gv
  JOIN games g ON g.id = gv.game_id
  JOIN game_status gs ON gs.game_id = g.id
  WHERE gv.id = sqlc.arg(game_video_id)
    AND gs.stage = ANY(sqlc.arg(stages)::text[]::game_stage[])
    AND gv.player_id <> sqlc.arg(player_id)
    AND EXISTS (
      SELECT 1
      FROM game_players gp
      WHERE gp.game_id = gv.game_id
        AND gp.player_id = sqlc.arg(player_id)
    ) 
FOR SHARE OF g;

-- name: VoteForVideo :one
INSERT INTO game_votes (game_video_id, player_id, value)
VALUES ($1, $2, $3)
ON CONFLICT (game_video_id, player_id)
DO UPDATE 
SET value = EXCLUDED.value, 
    voted_at = now()
RETURNING *;

-- name: FetchVotesByPlayers :many
SELECT gp.game_id, gp.player_id, gvd.id as game_video_id, gvd.video_id, gvt.value, gvt.voted_at
FROM game_players gp
LEFT JOIN game_votes gvt 
  ON gp.player_id = gvt.player_id
LEFT JOIN game_videos gvd 
  ON gvt.game_video_id = gvd.id 
  AND gvd.round_n = $2
WHERE gp.game_id = $1;
