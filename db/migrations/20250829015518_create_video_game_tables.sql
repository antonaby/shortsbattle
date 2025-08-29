-- +goose Up
-- +goose StatementBegin
CREATE TABLE
  game_videos (
    game_id BIGINT NOT NULL REFERENCES games (id) ON DELETE CASCADE,
    player_id BIGINT NOT NULL REFERENCES players (tg_id) ON DELETE CASCADE,
    video_id BIGINT NOT NULL REFERENCES videos (id) ON DELETE CASCADE,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT now (),
    PRIMARY KEY (game_id, player_id, video_id)
  );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS game_videos;
-- +goose StatementEnd