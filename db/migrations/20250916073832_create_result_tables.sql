-- +goose Up
-- +goose StatementBegin
CREATE TABLE
  game_video_results (
    game_id BIGINT NOT NULL REFERENCES games (id) ON DELETE CASCADE,
    game_video_id BIGINT NOT NULL REFERENCES game_videos (id) ON DELETE CASCADE,
    result jsonb NOT NULL DEFAULT '{}'::jsonb,
    calculated_at TIMESTAMPTZ NOT NULL DEFAULT now (),
    PRIMARY KEY (game_id, game_video_id),
    CONSTRAINT value_is_object CHECK (jsonb_typeof(result) = 'object')
  );

CREATE TABLE
  player_results (
    player_id BIGINT NOT NULL REFERENCES players (tg_id) ON DELETE CASCADE,
    game_id BIGINT NOT NULL REFERENCES games (id) ON DELETE CASCADE,
    result jsonb NOT NULL DEFAULT '{}'::jsonb,
    points BIGINT NOT NULL DEFAULT 0,
    place INT NOT NULL DEFAULT -1, 
    calculated_at TIMESTAMPTZ NOT NULL DEFAULT now (),
    PRIMARY KEY (player_id, game_id),
    CONSTRAINT value_is_object CHECK (jsonb_typeof(result) = 'object')
  );

CREATE INDEX idx_player_results_game ON player_results (game_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_player_results_game;

DROP TABLE IF EXISTS game_video_results;

DROP TABLE IF EXISTS player_results;
-- +goose StatementEnd
