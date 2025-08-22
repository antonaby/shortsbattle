-- +goose Up
-- +goose StatementBegin
-- Join (or create) a game for a theme with capacity left.
-- Returns the game_id the player ended up in.
CREATE OR REPLACE FUNCTION join_game_for_theme(
  p_theme_id              BIGINT,
  p_player_id             BIGINT,
  p_state                 game_state,
  p_max_players           INT,
  p_next_state_change_in  INTERVAL
) RETURNS BIGINT
LANGUAGE plpgsql
AS $$
DECLARE
  v_game_id BIGINT;
BEGIN
  -- 1) Serialize by theme using an advisory *transaction* lock
  PERFORM pg_advisory_xact_lock(1, p_theme_id::int);

  -- 2) Find a lobby game with room, preferring the one with the fewest players
  SELECT g.id INTO v_game_id 
  FROM games g
  LEFT JOIN game_players gp ON gp.game_id = g.id
  WHERE g.theme_id = p_theme_id
    AND g.state = p_state
  GROUP BY g.id, g.created_at
  HAVING COUNT(gp.player_id) < p_max_players
  ORDER BY COUNT(gp.player_id) ASC, g.created_at ASC, g.id ASC
  LIMIT 1;

  -- 3) Create a new game if none found
  IF v_game_id IS NULL THEN
    INSERT INTO games (theme_id, state, next_state_change_at)
    VALUES (
      p_theme_id, 
      p_state, 
      now() + p_next_state_change_in
    )
    RETURNING id INTO v_game_id;
  END IF;

  -- Add player to the chosen/created game.
  -- If they're already in, do nothing (idempotent).
  INSERT INTO game_players (game_id, player_id)
  VALUES (v_game_id, p_player_id)
  ON CONFLICT DO NOTHING;

  RETURN v_game_id;
END;
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION join_game_for_theme CASCADE;
-- +goose StatementEnd
