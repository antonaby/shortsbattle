-- name: JoinGameForTheme :one
SELECT join_game_for_theme($1, $2, $3, $4, $5) AS game_id;
