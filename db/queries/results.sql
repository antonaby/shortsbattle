-- name: CreateFinalResult :one
INSERT INTO game_final_results(game_id, result) 
VALUES ($1, $2) 
ON CONFLICT (game_id) DO UPDATE
SET result = EXCLUDED.result,
    calculated_at = now()
RETURNING *;

-- name: GetFinalResult :one
SELECT f.* FROM game_final_results f
WHERE f.game_id = $1
  AND EXISTS (
    SELECT 1 FROM game_players gp WHERE gp.game_id = $1 AND gp.player_id = $2
  );
