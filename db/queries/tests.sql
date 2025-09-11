-- name: TestCreateGame :one
INSERT INTO games (
  theme_id
) VALUES (
  sqlc.arg(theme_id)
)
RETURNING *;

-- name: TestCreateGameStatus :one
INSERT INTO game_status (game_id, theme_id, stage, state_changed_at)
VALUES ($1, $2, $3, now())
RETURNING *;

-- name: TestUpdateGameStage :one
UPDATE game_status
SET stage = $1
WHERE game_id = $2
RETURNING *;

-- name: TestAddPlayerToGame :one
INSERT INTO game_players (
  game_id,
  player_id,
  mode
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: TestGetGameById :one
SELECT * FROM games WHERE id = $1;

-- name: TestGetGameStatusById :one
SELECT *,  
  GREATEST(
    COALESCE((EXTRACT(EPOCH FROM (now() - gs.state_changed_at)) * 1000)::bigint, 0),
    0
  )::bigint AS past_ms
FROM game_status 
WHERE game_id = $1;