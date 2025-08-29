-- +goose Up
-- +goose StatementBegin
ALTER TABLE game_videos
ADD CONSTRAINT unique_player_game UNIQUE (game_id, player_id);
-- +goose StatementEnd
