-- +goose Up
-- +goose StatementBegin
ALTER TABLE game_votes
  ADD COLUMN is_err BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN err_msg TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE game_votes
  DROP COLUMN IF EXISTS is_err,
  DROP COLUMN IF EXISTS err_msg;
-- +goose StatementEnd
