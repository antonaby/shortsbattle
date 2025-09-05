-- +goose Up
-- +goose StatementBegin
-- games table
CREATE TABLE
    themes (
        id BIGSERIAL PRIMARY KEY,
        title TEXT NOT NULL,
        description TEXT,
        created_at TIMESTAMPTZ NOT NULL DEFAULT now()
    );

CREATE TABLE
    rounds (
        round_n INT NOT NULL,
        title TEXT NOT NULL,
        description TEXT,
        theme_id BIGINT NOT NULL REFERENCES themes (id) ON DELETE CASCADE,
        created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
        PRIMARY KEY (theme_id, round_n)
    );

-- players table
CREATE TABLE
    players (
        tg_id BIGINT PRIMARY KEY,
        tg_username TEXT NOT NULL,
        tg_language_code TEXT NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT now()
    );

-- videos table
CREATE TABLE
    videos (
        id BIGSERIAL PRIMARY KEY,
        video_url TEXT NOT NULL UNIQUE,
        oembed jsonb NOT NULL DEFAULT '{}'::jsonb,
        added_at TIMESTAMPTZ NOT NULL DEFAULT now(),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
        CONSTRAINT videos_oembed_is_object CHECK (jsonb_typeof(oembed) = 'object')
    );

CREATE TABLE 
    player_videos (
        player_id BIGINT NOT NULL REFERENCES players (tg_id) ON DELETE CASCADE,
        video_id BIGINT NOT NULL REFERENCES videos (id) ON DELETE CASCADE,
        added_at TIMESTAMPTZ NOT NULL DEFAULT now(),
        PRIMARY KEY (player_id, video_id)
    );

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS player_videos;

DROP TABLE IF EXISTS videos;

DROP TABLE IF EXISTS players;

DROP TABLE IF EXISTS video_requests;

DROP TABLE IF EXISTS themes;
-- +goose StatementEnd