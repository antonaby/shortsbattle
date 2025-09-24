-- name: CreateGameVideoResult :one
INSERT INTO game_video_results(game_id, game_video_id, result)
VALUES ($1, $2, $3)
ON CONFLICT (game_id, game_video_id) DO UPDATE
  SET result = EXCLUDED.result,
      calculated_at = now()
RETURNING *;

-- name: GetPlayersResult :many
SELECT ps.* FROM player_results ps
WHERE ps.game_id = $1
  AND EXISTS (
    SELECT 1 FROM game_players gp WHERE gp.game_id = $1 AND gp.player_id = $2
  );
