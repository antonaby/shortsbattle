-- name: AddPlayerToGame :exec
INSERT INTO game_players (game_id, player_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: GetPlayersInGame :many
SELECT p.id, p.username, p.created_at
FROM players p
JOIN game_players gp ON p.id = gp.player_id
WHERE gp.game_id = $1;

-- name: IsPlayerInGame :one
SELECT EXISTS (
    SELECT 1 FROM game_players
    WHERE game_id = $1 AND player_id = $2
) AS exists;

-- name: FindLeastCrowdedGameByThemeAndStatuses :one
SELECT g.*, COUNT(gp.player_id) AS player_count
FROM games AS g
LEFT JOIN game_players AS gp
  ON gp.game_id = g.id
WHERE g.theme_id = $1
  AND g.status = ANY(($2::text[])::game_status[])
GROUP BY g.id
ORDER BY player_count
LIMIT 1;