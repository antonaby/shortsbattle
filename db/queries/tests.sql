-- name: TestCreateGame :one
INSERT INTO games (
  theme_id
) VALUES (
  sqlc.arg(theme_id)
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