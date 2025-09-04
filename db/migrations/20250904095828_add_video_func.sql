-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION add_video_for_player(
    p_player_id BIGINT,
    p_video_url TEXT,
    p_oembed JSONB
)
RETURNS videos AS $$
DECLARE
    v_video videos;
BEGIN
    -- Step 1: Insert or update video
    INSERT INTO videos (video_url, oembed)
    VALUES (p_video_url, p_oembed)
    ON CONFLICT (video_url) DO UPDATE
      SET video_url  = EXCLUDED.video_url,
          updated_at = now()
    RETURNING * INTO v_video;

    -- Step 2: Link video to player
    INSERT INTO player_videos (player_id, video_id)
    VALUES (p_player_id, v_video.id)
    ON CONFLICT (player_id, video_id) DO UPDATE
      SET added_at = now();

    -- Step 3: Return video
    RETURN v_video;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION add_video_for_player CASCADE;
-- +goose StatementEnd
