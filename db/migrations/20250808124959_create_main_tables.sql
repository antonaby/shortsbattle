-- +goose Up
-- +goose StatementBegin
-- games table
CREATE TABLE
    themes (
        id BIGSERIAL PRIMARY KEY,
        name TEXT NOT NULL,
        description TEXT,
        created_at TIMESTAMPTZ NOT NULL DEFAULT now()
    );

CREATE TABLE
    video_requests (
        id BIGSERIAL PRIMARY KEY,
        request TEXT NOT NULL,
        theme_id BIGINT NOT NULL REFERENCES themes (id) ON DELETE CASCADE,
        created_at TIMESTAMPTZ NOT NULL DEFAULT now()
    );

-- players table
CREATE TABLE
    players (
        id BIGSERIAL PRIMARY KEY,
        tg_id BIGINT NOT NULL,
        tg_username TEXT NOT NULL,
        tg_language_code TEXT NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT now()
    );

-- videos table
CREATE TABLE
    videos (
        id BIGSERIAL PRIMARY KEY,
        player_id BIGINT NOT NULL REFERENCES players (id) ON DELETE CASCADE,
        video_url TEXT NOT NULL,
        oembed jsonb NOT NULL DEFAULT '{}'::jsonb,
        added_at TIMESTAMPTZ NOT NULL DEFAULT now(),
        CONSTRAINT videos_oembed_is_object CHECK (jsonb_typeof(oembed) = 'object')
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS videos;

DROP TABLE IF EXISTS players;

DROP TABLE IF EXISTS video_requests;

DROP TABLE IF EXISTS themes;
-- +goose StatementEnd