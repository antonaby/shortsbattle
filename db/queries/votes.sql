-- name: PlayerCanVote :one
SELECT EXISTS (
  SELECT 1
  FROM game_videos gv
  JOIN games g ON g.id = gv.game_id
  WHERE gv.id = sqlc.arg(game_video_id)
    AND g.state = ANY(sqlc.arg(states)::text[]::game_state[])
    AND EXISTS (
      SELECT 1
      FROM game_players gp
      WHERE gp.game_id = gv.game_id
        AND gp.player_id = sqlc.arg(player_id)
    )
) AS in_game_and_in_states;

-- name: VoteForVideo :one
INSERT INTO game_votes (game_video_id, player_id, value)
VALUES ($1, $2, $3)
ON CONFLICT (game_video_id, player_id)
DO UPDATE 
SET value = EXCLUDED.value, 
    voted_at = now()
RETURNING *;
