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
    stage game_stage NOT NULL DEFAULT 'lobby',
    round_n INT NOT NULL DEFAULT 0,
    state_changed_at TIMESTAMPTZ NOT NULL,
    next_game_update_at TIMESTAMPTZ,
    enqueued_at TIMESTAMPTZ,
    PRIMARY KEY (game_id)
  );

CREATE INDEX idx_game_status_stage ON game_status (stage, theme_id);
CREATE INDEX idx_game_status_enqueued_at_null ON game_status (enqueued_at) WHERE enqueued_at IS NULL;

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