-- name: GetGameVideoShareLock :one
SELECT 
  gs.*,
  gv.id as game_video_id,
  gv.player_id,
  gv.video_id,
  gv.submitted_at
  FROM game_videos gv
  JOIN game_status gs ON gs.game_id = gv.game_id
  WHERE gv.id = sqlc.arg(game_video_id)
    AND gv.player_id <> sqlc.arg(player_id) -- player can't vote for its own video
    AND EXISTS (
      SELECT 1
      FROM game_players gp
      WHERE gp.game_id = gv.game_id
        AND gp.player_id = sqlc.arg(player_id)
    ) 
FOR SHARE OF gs;

-- name: VoteForVideoLD :one
INSERT INTO game_votes (game_video_id, player_id, value, is_err, err_msg)
VALUES (sqlc.arg(game_video_id), sqlc.arg(player_id), jsonb_build_object('value', sqlc.arg(value)::text), false, NULL)
ON CONFLICT (game_video_id, player_id)
DO UPDATE 
SET value = EXCLUDED.value, 
    voted_at = now()
RETURNING *;

-- name: SetVoteErr :one
INSERT INTO game_votes (game_video_id, player_id, is_err, err_msg)
VALUES ($1, $2, true, $3)
ON CONFLICT (game_video_id, player_id)
DO UPDATE 
SET err_msg = EXCLUDED.err_msg, 
    voted_at = now()
RETURNING *;

-- name: GetLDVotes :many
WITH players AS (
  SELECT gp.player_id
  FROM game_players gp
  WHERE gp.game_id = $1
),
videos AS (
  SELECT gv.id AS game_video_id, gv.player_id AS author_id, gv.round_n
  FROM game_videos gv
  WHERE gv.game_id = $1
),
expected AS (
  SELECT p.player_id, v.game_video_id, v.round_n, v.author_id
  FROM players p
  CROSS JOIN videos v
  WHERE p.player_id <> v.author_id
)
SELECT 
  e.player_id,
  e.game_video_id,
  e.round_n,
  e.author_id,
  COALESCE(gvt.value ->> 'value', '')::text as value,
  gvt.voted_at
  FROM expected e
  LEFT JOIN game_votes gvt
    ON gvt.game_video_id = e.game_video_id
   AND gvt.player_id     = e.player_id
  ORDER BY e.round_n, e.game_video_id; 

-- name: GetVotesForRound :many
WITH players AS (
  SELECT gp.player_id
  FROM game_players gp
  WHERE gp.game_id = $1
),
videos AS (
  SELECT gv.id AS game_video_id, gv.player_id AS author_id
  FROM game_videos gv
  WHERE gv.game_id = $1
    AND gv.round_n = $2
),
expected AS (
  SELECT p.player_id, v.game_video_id
  FROM players p
  CROSS JOIN videos v
  WHERE p.player_id <> v.author_id
)
SELECT 
  e.player_id,
  e.game_video_id,
  gvt.value,
  gvt.voted_at
  FROM expected e
  LEFT JOIN game_votes gvt
    ON gvt.game_video_id = e.game_video_id
   AND gvt.player_id     = e.player_id;
