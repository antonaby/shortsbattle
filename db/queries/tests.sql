-- name: TestCreateGame :one
INSERT INTO games (
  theme_id
) VALUES (
  sqlc.arg(theme_id)
)
RETURNING *;

-- name: TestCreateGameStatus :one
INSERT INTO game_status (game_id, stage, state_changed_at, next_state_change_at, next_enqueue_at)
VALUES ($1, $2, now(), now(), now())
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
    (EXTRACT(EPOCH FROM (next_state_change_at - now())) * 1000)::bigint, 
    0
  )::bigint AS remaining_ms
FROM game_status 
WHERE game_id = $1;