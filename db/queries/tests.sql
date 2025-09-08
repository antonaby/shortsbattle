-- name: TestCreateGame :one
INSERT INTO games (
  theme_id,
  state,
  next_state_change_at
) VALUES (
  sqlc.arg(theme_id),
  sqlc.arg(state),
  now() + sqlc.arg(next_change_in)::interval
)
RETURNING *;

-- name: TestAddPlayerToGame :one
INSERT INTO game_players (
  game_id,
  player_id
) VALUES (
  $1, $2
)
RETURNING *;

-- name: TestGetGameById :one
SELECT * from games where id = $1;