-- +goose Up
-- +goose StatementBegin
CREATE TABLE
  game_final_results (
    game_id BIGINT NOT NULL REFERENCES games (id) ON DELETE CASCADE,
    result jsonb NOT NULL DEFAULT '{}'::jsonb,
    calculated_at TIMESTAMPTZ NOT NULL DEFAULT now (),
    PRIMARY KEY (game_id),
    CONSTRAINT value_is_object CHECK (jsonb_typeof(result) = 'object')
  );

CREATE TABLE
  player_stats (
    player_id BIGINT NOT NULL REFERENCES players (tg_id) ON DELETE CASCADE,
    game_id BIGINT NOT NULL REFERENCES games (id) ON DELETE CASCADE,
    points BIGINT NOT NULL DEFAULT 0, 
    added_at TIMESTAMPTZ NOT NULL DEFAULT now (),
    PRIMARY KEY (player_id, game_id)
  );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS game_final_results;

DROP TABLE IF EXISTS player_stats;
-- +goose StatementEnd
