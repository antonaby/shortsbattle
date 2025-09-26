-- +goose Up
-- +goose StatementBegin
ALTER TABLE players 
  ADD COLUMN last_online TIMESTAMPTZ,
  ADD COLUMN is_online TIMESTAMPTZ;

ALTER TABLE game_players
  DROP COLUMN is_active,
  ADD COLUMN is_online TIMESTAMPTZ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE players 
  DROP COLUMN IF EXISTS last_online, 
  DROP COLUMN IF EXISTS is_online;

ALTER TABLE game_players
  DROP COLUMN IF EXISTS is_online;
-- +goose StatementEnd
