-- name: CreateVote :one
INSERT INTO votes (game_id, submission_id, voter_id, voted_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetVotesByGame :many
SELECT id, game_id, submission_id, voter_id, voted_at
FROM votes
WHERE game_id = $1;

-- name: GetVoteCountsBySubmission :many
SELECT submission_id, COUNT(*) AS vote_count
FROM votes
WHERE game_id = $1
GROUP BY submission_id;

-- name: GetWinningSubmission :one
SELECT submission_id, COUNT(*) AS vote_count
FROM votes
WHERE game_id = $1
GROUP BY submission_id
ORDER BY vote_count DESC
LIMIT 1;

-- name: HasPlayerVoted :one
SELECT EXISTS (
    SELECT 1 FROM votes
    WHERE game_id = $1 AND voter_id = $2
) AS exists;