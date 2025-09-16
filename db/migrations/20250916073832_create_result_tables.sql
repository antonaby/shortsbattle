-- +goose Up
-- +goose StatementBegin
CREATE TABLE
  game_final_results (
    game_id BIGINT NOT NULL REFERENCES games (id) ON DELETE CASCADE,
    result jsonb NOT NULL DEFAULT '{}'::jsonb,
    calculated_at TIMESTAMPTZ NOT NULL DEFAULT now (),
    PRIMARY KEY (game_id),
    CONSTRAINT value_is_object CHECK (jsonb_typeof(result) = 'object')
  );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS game_final_results;
-- +goose StatementEnd
