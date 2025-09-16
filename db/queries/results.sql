-- name: CreateFinalResult :one
INSERT INTO game_final_results(game_id, result) 
VALUES ($1, $2) 
RETURNING *;
