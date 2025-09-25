-- name: CreatePlayer :one
INSERT INTO players (tg_id, tg_username, tg_language_code) VALUES ($1, $2, $3) RETURNING *;

-- name: GetPlayerByTgId :one
SELECT * FROM players WHERE tg_id = $1;

-- name: SetPlayerOnline :one
UPDATE players 
  SET 
    last_online = now(), 
    is_online = now() 
  WHERE tg_id = $1 RETURNING *;

-- name: SetPlayerOffline :one
UPDATE players 
  SET 
    is_online = NULL 
  WHERE tg_id = $1 RETURNING *;
