-- +goose Up
-- +goose StatementBegin
-- games table
CREATE TABLE
    themes (
        id BIGSERIAL PRIMARY KEY,
        name TEXT NOT NULL,
        description TEXT,
        created_at TIMESTAMP
        WITH
            TIME ZONE NOT NULL DEFAULT now ()
    );

CREATE TABLE
    video_requests (
        id BIGSERIAL PRIMARY KEY,
        request TEXT NOT NULL,
        theme_id BIGINT NOT NULL REFERENCES themes (id) ON DELETE CASCADE,
        created_at TIMESTAMP
        WITH
            TIME ZONE NOT NULL DEFAULT now ()
    );

-- players table
CREATE TABLE
    players (
        id BIGSERIAL PRIMARY KEY,
        username TEXT NOT NULL,
        created_at TIMESTAMP
        WITH
            TIME ZONE NOT NULL DEFAULT now ()
    );

-- videos table
CREATE TABLE
    videos (
        id BIGSERIAL PRIMARY KEY,
        player_id BIGINT NOT NULL REFERENCES players (id) ON DELETE CASCADE,
        video_url TEXT NOT NULL,
        added_at TIMESTAMP
        WITH
            TIME ZONE NOT NULL DEFAULT now ()
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS videos;

DROP TABLE IF EXISTS players;

DROP TABLE IF EXISTS video_requests;

DROP TABLE IF EXISTS themes;
-- +goose StatementEnd