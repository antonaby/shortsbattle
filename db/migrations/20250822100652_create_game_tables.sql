-- +goose Up
-- +goose StatementBegin
CREATE TYPE game_state AS ENUM ('lobby', 'submitting', 'wathching', 'completed');

CREATE TABLE
  games (
    id BIGSERIAL PRIMARY KEY,
    theme_id BIGINT NOT NULL REFERENCES themes (id) ON DELETE CASCADE,
    state game_state NOT NULL DEFAULT 'lobby',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    state_changed_at TIMESTAMPTZ,
    next_state_change_at TIMESTAMPTZ,
    enqueued BOOLEAN DEFAULT false,
    enqueued_at TIMESTAMPTZ,
    processed BOOLEAN DEFAULT false,
    processed_at TIMESTAMPTZ
  );

CREATE TABLE
  game_players (
    game_id BIGINT NOT NULL REFERENCES games (id) ON DELETE CASCADE,
    player_id BIGINT NOT NULL REFERENCES players (tg_id) ON DELETE CASCADE,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT now (),
    PRIMARY KEY (game_id, player_id)
  );

-- ---------- Trigger to keep state_changed_at fresh ----------
CREATE OR REPLACE FUNCTION trg_games_touch_state_changed_at()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    -- On insert, ensure state_changed_at is "now" (or keep provided value if set)
    IF NEW.state_changed_at IS NULL THEN
      NEW.state_changed_at := now();
    END IF;
    RETURN NEW;
  ELSIF TG_OP = 'UPDATE' THEN
    -- Only update when status actually changes
    IF NEW.state IS DISTINCT FROM OLD.state THEN
      NEW.state_changed_at := now();
    END IF;
    RETURN NEW;
  END IF;
  RETURN NEW;
END;
$$;

CREATE TRIGGER games_touch_state_changed_at
BEFORE INSERT OR UPDATE ON games
FOR EACH ROW
EXECUTE FUNCTION trg_games_touch_state_changed_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS game_players;

DROP TABLE IF EXISTS games;

DROP TYPE game_state;

-- +goose StatementEnd