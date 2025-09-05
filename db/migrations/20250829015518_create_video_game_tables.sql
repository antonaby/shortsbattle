-- +goose Up
-- +goose StatementBegin
CREATE TABLE
  game_videos (
    id BIGSERIAL PRIMARY KEY,
    game_id BIGINT NOT NULL REFERENCES games (id) ON DELETE CASCADE,
    player_id BIGINT NOT NULL REFERENCES players (tg_id) ON DELETE CASCADE,
    video_id BIGINT NOT NULL REFERENCES videos (id) ON DELETE CASCADE,
    round_n INT NOT NULL,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT now (),
    CONSTRAINT unique_player_game_round UNIQUE (game_id, player_id, round_n)
  );

CREATE TYPE vote_value AS ENUM ('like', 'dislike', 'skip');

CREATE TABLE
  game_votes (
    game_video_id BIGINT NOT NULL REFERENCES game_videos (id) ON DELETE CASCADE,
    player_id BIGINT NOT NULL REFERENCES players (tg_id) ON DELETE CASCADE,
    value vote_value NOT NULL,
    voted_at TIMESTAMPTZ NOT NULL DEFAULT now (),
    PRIMARY KEY (game_video_id, player_id)
  );

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS game_videos;
DROP TABLE IF EXISTS game_votes;
-- +goose StatementEnd