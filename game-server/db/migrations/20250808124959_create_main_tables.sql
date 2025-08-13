-- +goose Up
-- +goose StatementBegin

-- games table
CREATE TABLE themes (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

-- game_status enum
CREATE TYPE game_status AS ENUM ('created', 'lobby', 'submitting', 'voting', 'winner', 'complete');

-- games table
CREATE TABLE games (
    id BIGSERIAL PRIMARY KEY,
    theme_id BIGINT NOT NULL REFERENCES themes(id) ON DELETE CASCADE,
    status game_status NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

-- players table
CREATE TABLE players (
    id BIGSERIAL PRIMARY KEY,
    username TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

-- game_players table (many-to-many between games and players)
CREATE TABLE game_players (
    game_id BIGINT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    player_id BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    PRIMARY KEY (game_id, player_id)
);

-- videos table
CREATE TABLE videos (
    id BIGSERIAL PRIMARY KEY,
    game_id BIGINT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    player_id BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    video_url TEXT NOT NULL,
    submitted_at TIMESTAMP NOT NULL DEFAULT now()
);

-- votes table
CREATE TABLE votes (
    id BIGSERIAL PRIMARY KEY,
    game_id BIGINT NOT NULL REFERENCES games(id)  ON DELETE CASCADE,
    video_id BIGINT NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    voter_id BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    voted_at TIMESTAMP NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS votes;
DROP TABLE IF EXISTS videos;
DROP TABLE IF EXISTS game_players;
DROP TABLE IF EXISTS players;
DROP TABLE IF EXISTS games;
-- +goose StatementEnd
