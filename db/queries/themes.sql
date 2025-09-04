-- name: CreateTheme :one
INSERT INTO themes (name, description) VALUES ($1, $2) RETURNING *;

-- name: CreateVideoRequest :one
INSERT INTO video_requests (request, theme_id) VALUES ($1, $2) RETURNING *;

-- name: GetTheme :one
SELECT
  t.id          AS theme_id,
  t.name,
  t.description,
  t.created_at,
  COALESCE(
    jsonb_agg(
      jsonb_build_object(
        'id', vr.id,
        'request', vr.request,
        'created_at', vr.created_at
      )
      ORDER BY vr.created_at
    ) FILTER (WHERE vr.id IS NOT NULL),
    '[]'::jsonb
  ) AS requests
FROM themes t
LEFT JOIN video_requests vr
  ON vr.theme_id = t.id
WHERE t.id = $1
GROUP BY t.id, t.name, t.description, t.created_at
ORDER BY t.created_at;

-- name: GetVideoRequests :many
SELECT * FROM video_requests WHERE theme_id = $1;

-- name: ListThemes :many
SELECT * FROM themes ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: ListAllThemes :many
SELECT * FROM themes ORDER BY created_at;
