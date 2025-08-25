-- name: JoinGameForTheme :one
SELECT join_game_for_theme($1, $2, $3, $4, $5, $6) AS game_id;

-- name: AdvanceGames :many
WITH candidates AS (
    SELECT id, state
    FROM games
    WHERE next_state_change_at <= now()
      AND enqueued = false
      AND state <> 'completed'::game_state     
    ORDER BY next_state_change_at ASC, id
    FOR UPDATE SKIP LOCKED
    LIMIT $1
  ),
  upd AS (
    UPDATE games g
    SET
      enqueued = true,
      enqueued_at = now()
    FROM candidates c
    WHERE g.id = c.id
    RETURNING g.*
  )
  SELECT * FROM upd;

-- name: UpdateGameStatus :exec
UPDATE games
SET 
    state = $1,        
    processed = $2,     
    processed_at = NOW()
WHERE id = $3;
