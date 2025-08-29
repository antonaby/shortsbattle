-- +goose Up
-- +goose StatementBegin
CREATE TYPE vote_value AS ENUM ('like', 'dislike', 'skip');

CREATE TABLE
  game_votes (
    game_id BIGINT NOT NULL REFERENCES games (id) ON DELETE CASCADE,
    player_id BIGINT NOT NULL REFERENCES players (tg_id) ON DELETE CASCADE,
    video_id BIGINT NOT NULL REFERENCES videos (id) ON DELETE CASCADE,
    voted_at TIMESTAMPTZ NOT NULL DEFAULT now (),
    value vote_value NOT NULL,
    PRIMARY KEY (game_id, player_id, video_id)
  );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS game_votes;
-- +goose StatementEnd
