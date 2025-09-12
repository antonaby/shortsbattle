-- +goose Up
-- +goose StatementBegin
-- Join (or create) a game for a theme with capacity left.
-- Returns the game_id the player ended up in.
CREATE OR REPLACE FUNCTION join_game(
  p_theme_id              BIGINT,
  p_player_id             BIGINT,
  p_lobby_stage           game_stage,
  p_next_game_update_in   INTERVAL,
  p_max_players           INT,
  p_mode                  player_game_mode
) RETURNS BIGINT
LANGUAGE plpgsql
AS $$
DECLARE
  v_game_id BIGINT;
BEGIN
  -- 1) Serialize by *player* so they can't join two games concurrently
  PERFORM pg_advisory_xact_lock(p_player_id);

  -- 2) If the player is already in a game (same theme), return that game_id
  SELECT gp.game_id INTO v_game_id
  FROM game_status gs
  JOIN game_players gp ON gp.game_id = gs.game_id AND gp.player_id = p_player_id
  WHERE gs.stage = p_lobby_stage  
    AND gs.theme_id = p_theme_id       
  LIMIT 1
  FOR SHARE OF gs;

  IF v_game_id IS NOT NULL THEN
    RETURN v_game_id;
  END IF;

  -- 3) Serialize by theme
  PERFORM pg_advisory_xact_lock(1, p_theme_id::int);

  -- 4) Find a lobby game with room
  SELECT gs.game_id INTO v_game_id 
  FROM game_status gs
  LEFT JOIN game_players gp ON gp.game_id = gs.game_id
  WHERE gs.stage = p_lobby_stage
    AND gs.theme_id = p_theme_id
  GROUP BY gs.game_id, gs.state_changed_at
  HAVING COUNT(gp.player_id) < p_max_players
  ORDER BY COUNT(gp.player_id) DESC, gs.state_changed_at ASC
  LIMIT 1;

  -- 5) Create a new game if none found
  IF v_game_id IS NULL THEN
    INSERT INTO games (theme_id)
    VALUES (
      p_theme_id
    )
    RETURNING id INTO v_game_id;
    
    INSERT INTO game_status (game_id, theme_id, stage, state_changed_at, next_game_update_at, update_key)
    VALUES (
      v_game_id, 
      p_theme_id,
      p_lobby_stage, 
      now(),
      now() + p_next_game_update_in,
      uuid_generate_v1mc()
    );
  END IF;

  -- Add player to the chosen/created game.
  -- If they're already in, do nothing (idempotent).
  INSERT INTO game_players (game_id, player_id, mode)
  VALUES (v_game_id, p_player_id, p_mode)
  ON CONFLICT DO NOTHING;

  RETURN v_game_id;
END;
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION join_game CASCADE;
-- +goose StatementEnd
