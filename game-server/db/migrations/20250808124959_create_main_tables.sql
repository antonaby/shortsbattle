-- +goose Up
-- +goose StatementBegin

-- game_status enum
CREATE TYPE game_status AS ENUM ('lobby', 'submitting', 'voting', 'complete');

-- games table
CREATE TABLE games (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    status game_status NOT NULL
);

-- players table
CREATE TABLE players (
    id BIGSERIAL PRIMARY KEY,
    username TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

-- game_players table (many-to-many between games and players)
CREATE TABLE game_players (
    game_id BIGINT REFERENCES games(id) ON DELETE CASCADE,
    player_id BIGINT REFERENCES players(id) ON DELETE CASCADE,
    PRIMARY KEY (game_id, player_id)
);

-- submissions table
CREATE TABLE submissions (
    id BIGSERIAL PRIMARY KEY,
    game_id BIGINT REFERENCES games(id) ON DELETE CASCADE,
    player_id BIGINT REFERENCES players(id) ON DELETE CASCADE,
    video_url TEXT NOT NULL,
    submitted_at TIMESTAMP NOT NULL DEFAULT now()
);

-- votes table
CREATE TABLE votes (
    id BIGSERIAL PRIMARY KEY,
    game_id BIGINT REFERENCES games(id) ON DELETE CASCADE,
    submission_id BIGINT REFERENCES submissions(id) ON DELETE CASCADE,
    voter_id BIGINT REFERENCES players(id) ON DELETE CASCADE,
    voted_at TIMESTAMP NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS votes;
DROP TABLE IF EXISTS submissions;
DROP TABLE IF EXISTS game_players;
DROP TABLE IF EXISTS players;
DROP TABLE IF EXISTS games;
-- +goose StatementEnd
