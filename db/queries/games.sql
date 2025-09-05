-- name: JoinGameForTheme :one
SELECT join_game_for_theme($1, $2, $3, $4, $5, $6) AS game_id;

-- name: FindGameWithPlayer :one
SELECT g.*
FROM game_players gp
JOIN games g ON g.id = gp.game_id
WHERE gp.game_id  = sqlc.arg(game_id)
  AND gp.player_id = sqlc.arg(player_id)
  AND g.state = ANY(sqlc.arg(states)::text[]::game_state[]);

-- name: AdvanceGames :many
WITH candidates AS (
    SELECT id, state
    FROM games
    WHERE next_state_change_at <= now()
      AND enqueued_at IS NULL
      AND state <> 'completed'::game_state     
    ORDER BY next_state_change_at ASC, id
    FOR UPDATE SKIP LOCKED
    LIMIT $1
  ),
  upd AS (
    UPDATE games g
    SET
      enqueued_at = now()
    FROM candidates c
    WHERE g.id = c.id
    RETURNING g.*
  )
  SELECT * FROM upd;

-- name: GetGameForPlayer :one
SELECT g.*, 
  COALESCE((EXTRACT(EPOCH FROM (next_state_change_at - now())) * 1000)::bigint, 0)::bigint AS remaining_ms
FROM games g
JOIN game_players gp ON gp.game_id = g.id
WHERE g.id = $1 AND gp.player_id = $2;

-- name: FetchGameAndLock :one
SELECT 
  g.*, 
  GREATEST(
    (EXTRACT(EPOCH FROM (next_state_change_at - now())) * 1000)::bigint,
    0
  )::bigint AS remaining_ms,
  GREATEST(
    (EXTRACT(EPOCH FROM (now() - state_changed_at)) * 1000)::bigint,
    0
  )::bigint AS past_ms
FROM games g 
WHERE id = $1 
FOR UPDATE;

-- name: CountPlayerInGame :one
SELECT COUNT(*) AS player_count
FROM game_players
WHERE game_id = $1;

-- name: UpdateGameStatus :one
UPDATE games SET 
  state = sqlc.arg(state), 
  next_state_change_at = now() + (sqlc.arg(next_state_in)::interval),
  round_n = sqlc.arg(round_n)
WHERE id = sqlc.arg(id) 
RETURNING *, 
  GREATEST(
    (EXTRACT(EPOCH FROM (next_state_change_at - now())) * 1000)::bigint,
    0
  )::bigint AS remaining_ms;

-- name: SetCompletedStatus :one
UPDATE games SET 
  state = sqlc.arg(state),
  next_state_change_at = NULL,
  round_n = 0
WHERE id = sqlc.arg(id) 
RETURNING *;

-- name: GetGameRounds :many
SELECT r.* 
FROM rounds r
JOIN games g ON g.theme_id = r.theme_id
WHERE g.id = $1
ORDER BY r.round_n;
