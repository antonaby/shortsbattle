-- name: JoinGameForTheme :one
SELECT join_game_for_theme($1, $2, $3, $4, $5) AS game_id;

-- name: AdvanceGames :many
SELECT * FROM advance_games_batch($1, $2, $3, $4);
