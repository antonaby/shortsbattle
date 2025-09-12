-- name: JoinGame :one
SELECT join_game(
  sqlc.arg(theme_id), 
  sqlc.arg(player_id), 
  sqlc.arg(lobby_stage), 
  sqlc.arg(next_game_update_in),
  sqlc.arg(max_players), 
  sqlc.arg(mode)
) AS game_id;

-- name: EnqueueGames :many
WITH cte AS (
  SELECT game_id
  FROM game_updates
  WHERE enqueued_at IS NULL
  FOR UPDATE SKIP LOCKED
)
UPDATE game_updates g
SET enqueued_at = now()
FROM cte
WHERE g.game_id = cte.game_id
RETURNING g.*;

-- name: EnqueueGame :one
WITH cte AS (
  SELECT gu.game_id
  FROM game_updates gu
  WHERE gu.game_id = $1
  FOR UPDATE SKIP LOCKED
)
UPDATE game_updates g
SET enqueued_at = now()
FROM cte
WHERE g.game_id = cte.game_id
RETURNING g.*;

-- name: GetGameShareLock :one
SELECT gs.*, gp.player_id, gp.mode, gp.is_active, gp.joined_at,
  GREATEST(
    COALESCE((EXTRACT(EPOCH FROM (gs.next_game_update_at - now())) * 1000)::bigint, 0),
    0
  )::bigint AS remaining_ms,
  GREATEST(
    COALESCE((EXTRACT(EPOCH FROM (now() - gs.state_changed_at)) * 1000)::bigint, 0),
    0
  )::bigint AS past_ms
FROM game_players gp
JOIN game_status gs ON gs.game_id = gp.game_id
WHERE gp.game_id  = sqlc.arg(game_id)
  AND gp.player_id = sqlc.arg(player_id)
FOR SHARE OF gs;

-- name: UpdateGameMode :one
UPDATE game_players 
SET mode = $3
WHERE game_id = $1 AND player_id = $2
RETURNING *;

-- name: GetGameLock :one
SELECT 
  gs.*,
  GREATEST(
    COALESCE((EXTRACT(EPOCH FROM (gs.next_game_update_at - now())) * 1000)::bigint, 0),
    0
  )::bigint AS remaining_ms,
  GREATEST(
    COALESCE((EXTRACT(EPOCH FROM (now() - gs.state_changed_at)) * 1000)::bigint, 0),
    0
  )::bigint AS past_ms
FROM game_status gs
WHERE gs.game_id = sqlc.arg(game_id) 
  AND (sqlc.arg(update_key)::UUID IS NULL OR gs.update_key = sqlc.arg(update_key)::UUID)
FOR UPDATE;

-- name: CountPlayersInGame :one
SELECT COUNT(*) AS player_count
FROM game_players
WHERE game_id = $1;

-- name: UpdateGameStatus :one
UPDATE game_status SET 
  stage = sqlc.arg(stage), 
  state_changed_at = now(),
  round_n = sqlc.arg(round_n),
  next_game_update_at = now() + sqlc.arg(next_game_update_in)::interval,
  update_key = uuid_generate_v1mc()
WHERE game_id = sqlc.arg(game_id) 
RETURNING *;

-- name: SetCompletedStatus :one
UPDATE game_status SET 
  stage = sqlc.arg(stage),
  next_state_change_at = NULL,
  next_enqueue_at = NULL,
  round_n = 0
WHERE game_id = sqlc.arg(game_id) 
RETURNING *;

-- name: GetGameRounds :many
SELECT r.* 
FROM rounds r
JOIN games g ON g.theme_id = r.theme_id
WHERE g.id = $1
ORDER BY r.round_n;
