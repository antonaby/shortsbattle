-- +goose Up
-- +goose StatementBegin
CREATE TYPE game_stage AS ENUM ('lobby', 'lobby-full', 'submit', 'submit-complete', 'watch', 'watch-complete', 'complete');

CREATE TABLE
  games (
    id BIGSERIAL PRIMARY KEY,
    theme_id BIGINT NOT NULL REFERENCES themes (id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
    completed_at TIMESTAMPTZ 
  );

CREATE TABLE
  game_status (
    game_id BIGINT NOT NULL REFERENCES games (id) ON DELETE CASCADE,
    theme_id BIGINT NOT NULL REFERENCES themes (id) ON DELETE SET NULL,
    mode game_mode NOT NULL,
    stage game_stage NOT NULL,
    round_n INT NOT NULL DEFAULT 0,
    state_changed_at TIMESTAMPTZ NOT NULL,
    update_key UUID NOT NULL,
    next_game_update_at TIMESTAMPTZ,
    PRIMARY KEY (game_id)
  );

CREATE INDEX idx_game_status_stage ON game_status (stage, theme_id);

CREATE TABLE 
  game_updates (
    game_id BIGINT NOT NULL REFERENCES games (id) ON DELETE CASCADE,
    update_key UUID NOT NULL,
    next_game_update_at TIMESTAMPTZ,
    enqueued_at TIMESTAMPTZ,
    PRIMARY KEY (game_id)
  );

CREATE INDEX idx_game_updates_enqueue ON game_updates (game_id) WHERE enqueued_at is NULL;

CREATE OR REPLACE FUNCTION sync_game_updates()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    IF NEW.stage = 'complete' THEN
      -- ensure no queued update remains if inserted already complete
      DELETE FROM game_updates WHERE game_id = NEW.game_id;
    ELSE
      INSERT INTO game_updates (game_id, update_key, next_game_update_at, enqueued_at)
      VALUES (NEW.game_id, NEW.update_key, NEW.next_game_update_at, NULL)
      ON CONFLICT (game_id) DO UPDATE
        SET next_game_update_at = EXCLUDED.next_game_update_at,
            update_key = EXCLUDED.update_key,
            enqueued_at = NULL;
    END IF;

    RETURN NEW;
  END IF;

  -- TG_OP = 'UPDATE'
  -- If stage is complete, remove any pending update
  IF NEW.stage = 'complete' THEN
    DELETE FROM game_updates WHERE game_id = NEW.game_id;
    RETURN NEW;
  END IF;

  -- Otherwise, keep game_updates in sync when next_game_update_at changes
  IF NEW.next_game_update_at IS DISTINCT FROM OLD.next_game_update_at THEN
    INSERT INTO game_updates (game_id, update_key, next_game_update_at, enqueued_at)
    VALUES (NEW.game_id, NEW.update_key, NEW.next_game_update_at, NULL)
    ON CONFLICT (game_id) DO UPDATE
      SET next_game_update_at = EXCLUDED.next_game_update_at,
          update_key = EXCLUDED.update_key,
          enqueued_at = NULL;
  END IF;

  RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_game_status_updates_sync ON game_status;
CREATE TRIGGER trg_game_status_updates_sync
AFTER INSERT OR UPDATE OF next_game_update_at, stage ON game_status
FOR EACH ROW
EXECUTE FUNCTION sync_game_updates();

CREATE OR REPLACE FUNCTION notify_game_updates()
RETURNS TRIGGER
LANGUAGE plpgsql AS $$
BEGIN
  PERFORM pg_notify(
    'status_updates',
    json_build_object(
      'event', TG_OP, 
      'game_id', NEW.game_id,
      'update_key', NEW.update_key,
      'next_game_update_at', NEW.next_game_update_at
    )::text
  );
  RETURN NEW;
END
$$;

DROP TRIGGER IF EXISTS trg_game_updates_notify ON game_updates;
CREATE TRIGGER trg_game_updates_notify
AFTER INSERT OR UPDATE OF update_key, next_game_update_at ON game_updates
FOR EACH ROW
EXECUTE FUNCTION notify_game_updates();

CREATE TYPE player_game_mode AS ENUM ('submit_and_vote', 'only_vote');

CREATE TABLE
  game_players (
    game_id BIGINT NOT NULL REFERENCES games (id) ON DELETE CASCADE,
    player_id BIGINT NOT NULL REFERENCES players (tg_id) ON DELETE CASCADE,
    mode player_game_mode NOT NULL,
    is_active BOOLEAN DEFAULT FALSE,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT now (),
    PRIMARY KEY (game_id, player_id)
  );

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS game_players;

DROP TABLE IF EXISTS game_stage;

DROP TABLE IF EXISTS games;

DROP TYPE game_state;

DROP TYPE player_game_mode;
-- +goose StatementEnd