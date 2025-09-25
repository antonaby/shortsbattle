-- name: CreateGameVideoResult :one
INSERT INTO game_video_results(game_id, game_video_id, result)
VALUES (sqlc.arg(game_id), sqlc.arg(game_video_id), jsonb_build_object('likes', sqlc.arg(likes)::int, 'dislikes', sqlc.arg(dislikes)::int))
ON CONFLICT (game_id, game_video_id) DO UPDATE
  SET result = EXCLUDED.result,
      calculated_at = now()
RETURNING *;

-- name: CreatePlayerResult :one
INSERT INTO player_results(player_id, game_id, result, place, points)
VALUES (
  sqlc.arg(player_id), 
  sqlc.arg(game_id), 
  jsonb_build_object('likes', sqlc.arg(likes)::int, 'dislikes', sqlc.arg(dislikes)::int), 
  sqlc.arg(place),
  sqlc.arg(points)
  )
ON CONFLICT (player_id, game_id) DO UPDATE
  SET result = EXCLUDED.result,
      points = EXCLUDED.points,
      calculated_at = now()
RETURNING *;

-- name: GetPlayerLDResults :many
SELECT 
  p.tg_id,
  p.tg_username,
  ps.place,
  ps.points,
  COALESCE((ps.result -> 'likes')::int, 0)::int as likes,
  COALESCE((ps.result -> 'dislikes')::int, 0)::int as dislikes
FROM player_results ps
JOIN players p ON p.tg_id = ps.player_id
WHERE ps.game_id = $1
ORDER BY ps.place;

-- name: GetPlayerRawResults :many
SELECT ps.* FROM player_results ps
WHERE ps.game_id = $1
  AND EXISTS (
    SELECT 1 FROM game_players gp WHERE gp.game_id = $1 AND gp.player_id = $2
  );
