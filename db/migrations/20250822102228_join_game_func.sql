-- +goose Up
-- +goose StatementBegin
-- Join (or create) a game for a theme with capacity left.
-- Returns the game_id the player ended up in.
CREATE OR REPLACE FUNCTION join_game(
  p_theme_id              BIGINT,
  p_player_id             BIGINT,
  p_lobby_stage           game_stage,
  p_lobby_stage_closed    INTERVAL,
  p_max_players           INT,
  p_next_state_change_in  INTERVAL
) RETURNS BIGINT
LANGUAGE plpgsql
AS $$
DECLARE
  v_game_id BIGINT;
BEGIN
  -- 1) Serialize by *player* so they can't join two games concurrently
  PERFORM pg_advisory_xact_lock(1, p_player_id::int);

  -- 2) If the player is already in a game (same theme), return that game_id
  SELECT gp.game_id INTO v_game_id
  FROM game_players gp
  JOIN games g ON g.id = gp.game_id
  JOIN game_status gs ON gs.game_id = g.id
  WHERE gp.player_id = p_player_id
    AND g.theme_id = p_theme_id       
    AND gs.stage = p_lobby_stage    
  ORDER BY gp.created_at DESC            
  LIMIT 1;

  IF v_game_id IS NOT NULL THEN
    RETURN v_game_id;
  END IF;

  -- 3) Serialize by theme using an advisory *transaction* lock
  PERFORM pg_advisory_xact_lock(2, p_theme_id::int);

  -- 4) Find a lobby game with room
  SELECT g.id INTO v_game_id 
  FROM games g
  JOIN game_status gs ON gs.game_id = g.id
  LEFT JOIN game_players gp ON gp.game_id = g.id
  WHERE g.theme_id = p_theme_id
    AND gs.stage = p_lobby_stage
    AND gs.next_state_change_at >= now() + p_lobby_stage_closed
  GROUP BY g.id, g.created_at
  HAVING COUNT(gp.player_id) < p_max_players
  ORDER BY COUNT(gp.player_id) DESC, g.created_at ASC, g.id ASC
  LIMIT 1;

  -- 5) Create a new game if none found
  IF v_game_id IS NULL THEN
    INSERT INTO games (theme_id)
    VALUES (
      p_theme_id
    )
    RETURNING id INTO v_game_id;
    INSERT INTO game_status (game_id, stage, state_changed_at, next_state_change_at)
    VALUES (
      v_game_id, 
      p_lobby_stage, 
      now(), 
      now() + p_next_state_change_in
    );
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
DROP FUNCTION join_game CASCADE;
-- +goose StatementEnd
