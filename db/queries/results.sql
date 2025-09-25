-- name: CreateGameVideoResult :one
INSERT INTO game_video_results(game_id, game_video_id, result)
VALUES (sqlc.arg(game_id), sqlc.arg(game_video_id), jsonb_build_object('likes', sqlc.arg(likes)::int, 'dislikes', sqlc.arg(dislikes)::int))
ON CONFLICT (game_id, game_video_id) DO UPDATE
  SET result = EXCLUDED.result,
      calculated_at = now()
RETURNING *;

-- name: CreatePlayerResult :one
INSERT INTO player_results(player_id, game_id, result, points)
VALUES ($1, $2, $3, $4)
ON CONFLICT (player_id, game_id) DO UPDATE
  SET result = EXCLUDED.result,
      points = EXCLUDED.points,
      calculated_at = now()
RETURNING *;

-- name: GetPlayersResult :many
SELECT ps.* FROM player_results ps
WHERE ps.game_id = $1
  AND EXISTS (
    SELECT 1 FROM game_players gp WHERE gp.game_id = $1 AND gp.player_id = $2
  );
