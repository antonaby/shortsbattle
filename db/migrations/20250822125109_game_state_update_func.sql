-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION advance_games_batch(
  p_lobby_to_submitting     INTERVAL,
  p_submitting_to_wathching INTERVAL,
  p_wathching_to_completed  INTERVAL,
  p_batch_limit             INT DEFAULT 100
)
RETURNS SETOF games
LANGUAGE plpgsql
AS $$
BEGIN
  RETURN QUERY
  WITH candidates AS (
    SELECT id, state
    FROM games
    WHERE next_state_change_at <= now()
      AND state_change_published = false
      AND state <> 'completed'::game_state     
    ORDER BY next_state_change_at, id
    FOR UPDATE SKIP LOCKED
    LIMIT p_batch_limit
  ),
  upd AS (
    UPDATE games g
    SET
      state = CASE c.state
        WHEN 'lobby'::game_state THEN 'submitting'::game_state
        WHEN 'submitting'::game_state THEN 'wathching'::game_state
        WHEN 'wathching'::game_state THEN 'completed'::game_state
        ELSE g.state
      END,
      next_state_change_at = CASE c.state
        WHEN 'lobby'::game_state THEN now() + p_lobby_to_submitting
        WHEN 'submitting'::game_state THEN now() + p_submitting_to_wathching
        WHEN 'wathching'::game_state THEN now() + p_wathching_to_completed
        WHEN 'completed'::game_state THEN now()
        ELSE NULL
      END
    FROM candidates c
    WHERE g.id = c.id
    RETURNING g.*
  )
  SELECT * FROM upd;
END;
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION advance_games_batch CASCADE;
-- +goose StatementEnd
