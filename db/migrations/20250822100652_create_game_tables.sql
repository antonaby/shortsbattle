-- +goose Up
-- +goose StatementBegin
CREATE TYPE game_stage AS ENUM ('lobby', 'submit', 'watch', 'complete');

CREATE TABLE
  games (
    id BIGSERIAL PRIMARY KEY,
    theme_id BIGINT NOT NULL REFERENCES themes (id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now ()
  );

CREATE TABLE
  game_status (
    game_id BIGINT NOT NULL REFERENCES games (id) ON DELETE CASCADE,
    stage game_stage NOT NULL DEFAULT 'lobby',
    round_n INT NOT NULL DEFAULT 0,
    state_changed_at TIMESTAMPTZ NOT NULL,
    next_state_change_at TIMESTAMPTZ NOT NULL,
    next_enqueue_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (game_id)
  );

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