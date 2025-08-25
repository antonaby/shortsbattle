-- name: JoinGameForTheme :one
SELECT join_game_for_theme($1, $2, $3, $4, $5, $6) AS game_id;

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
SELECT g.* FROM games g
JOIN game_players gp ON gp.game_id = g.id
WHERE g.id = $1 AND gp.player_id = $2;

-- name: GetGameAndLock :one
SELECT * from games WHERE id = $1 FOR UPDATE;

-- name: UpdateGameStatus :one
UPDATE games SET 
  state = sqlc.arg(state), 
  next_state_change_at = now() + (sqlc.arg(next_state_in)::interval) 
WHERE id = sqlc.arg(id) 
RETURNING *;

-- name: SetCompletedStatus :one
UPDATE games SET 
  state = sqlc.arg(state),
  next_state_change_at = NULL
WHERE id = sqlc.arg(id) 
RETURNING *;
