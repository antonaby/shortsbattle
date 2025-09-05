-- name: VoteForVideo :one
INSERT INTO game_votes (game_video_id, player_id, value)
VALUES ($1, $2, $3)
ON CONFLICT (game_video_id, player_id)
DO UPDATE 
SET value = EXCLUDED.value, 
    voted_at = now()
RETURNING *;


