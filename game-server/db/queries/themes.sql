-- name: CreateTheme :one
INSERT INTO themes (name, description) VALUES ($1, $2) RETURNING *;

-- name: GetThemeByID :one
SELECT * FROM themes WHERE id = $1;

-- name: ListThemes :many
SELECT * FROM themes ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: ListAllThemes :many
SELECT * FROM themes ORDER BY created_at;

-- name: UpdateTheme :one
UPDATE themes SET name = $1, description = $2 WHERE id = $3 RETURNING *;

-- name: DeleteTheme :exec
DELETE FROM themes WHERE id = $1;

-- name: SearchThemesByName :many
SELECT * FROM themes WHERE name ILIKE '%' || $1 || '%' ORDER BY created_at DESC LIMIT $2 OFFSET $3;