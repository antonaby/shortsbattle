-- name: CreateSubmission :one
INSERT INTO submissions (game_id, player_id, video_url, submitted_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetSubmissionsByGame :many
SELECT id, game_id, player_id, video_url, submitted_at
FROM submissions
WHERE game_id = $1;

-- name: GetSubmissionByPlayer :one
SELECT id, game_id, player_id, video_url, submitted_at
FROM submissions
WHERE game_id = $1 AND player_id = $2;