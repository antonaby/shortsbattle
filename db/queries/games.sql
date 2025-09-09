-- name: JoinGame :one
SELECT join_game(
  sqlc.arg(theme_id), 
  sqlc.arg(player_id), 
  sqlc.arg(lobby_stage), 
  sqlc.arg(lobby_stage_closed), 
  sqlc.arg(max_players), 
  sqlc.arg(next_stage_change_in),
  sqlc.arg(mode)
) AS game_id;

-- name: AdvanceGames :many
WITH candidates AS (
    SELECT game_id, stage
    FROM game_status
    WHERE (next_state_change_at <= now() OR next_enqueue_at <= now() OR next_enqueue_at is NULL)
      AND stage <> 'completed'::game_stage     
    ORDER BY next_state_change_at ASC, game_id
    FOR UPDATE SKIP LOCKED
    LIMIT sqlc.arg(batch_size)
  ),
  upd AS (
    UPDATE game_status gs
    SET
      next_enqueue_at = now() + (sqlc.arg(enqueue_interval)::interval)
    FROM candidates c
    WHERE gs.game_id = c.game_id
    RETURNING gs.*
  )
  SELECT * FROM upd;

-- name: GetGameByPlayerAndStage :one
SELECT g.*
FROM game_players gp
JOIN games g ON g.id = gp.game_id
JOIN game_status gs ON gs.game_id = g.id
WHERE gp.game_id  = sqlc.arg(game_id)
  AND gp.player_id = sqlc.arg(player_id)
  AND gs.stage = ANY(sqlc.arg(stages)::text[]::game_stage[])
FOR SHARE OF g;

-- name: GetGameByPlayer :one
SELECT g.*, 
  COALESCE((EXTRACT(EPOCH FROM (gs.next_state_change_at - now())) * 1000)::bigint, 0)::bigint AS remaining_ms
FROM game_players gp
JOIN games g ON g.id = gp.game_id
JOIN game_status gs ON gs.game_id = g.id
WHERE gp.game_id  = sqlc.arg(game_id)
  AND gp.player_id = sqlc.arg(player_id)
FOR SHARE OF g;

-- name: GetGameAndLock :one
SELECT 
  g.*, 
  gs.*,
  GREATEST(
    (EXTRACT(EPOCH FROM (gs.next_state_change_at - now())) * 1000)::bigint,
    0
  )::bigint AS remaining_ms,
  GREATEST(
    (EXTRACT(EPOCH FROM (now() - gs.state_changed_at)) * 1000)::bigint,
    0
  )::bigint AS past_ms
FROM games g 
JOIN game_status gs ON gs.game_id = g.id
WHERE g.id = $1 
FOR UPDATE;

-- name: CountPlayerInGame :one
SELECT COUNT(*) AS player_count
FROM game_players
WHERE game_id = $1;

-- name: UpdateGameStatus :one
UPDATE game_status SET 
  stage = sqlc.arg(stage), 
  next_state_change_at = now() + (sqlc.arg(next_state_in)::interval),
  round_n = sqlc.arg(round_n)
WHERE game_id = sqlc.arg(game_id) 
RETURNING *, 
  GREATEST(
    (EXTRACT(EPOCH FROM (next_state_change_at - now())) * 1000)::bigint,
    0
  )::bigint AS remaining_ms;

-- name: UpdateRemainingTime :one
UPDATE game_status SET 
  next_state_change_at = now() + (sqlc.arg(next_state_in)::interval)
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
