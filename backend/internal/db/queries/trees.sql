-- trees queries

-- name: CreateTree :one
INSERT INTO trees (owner_id, title)
VALUES ($1, $2)
RETURNING *;

-- name: ListTrees :many
SELECT * FROM trees WHERE owner_id = $1 ORDER BY created_at DESC;

-- name: GetTree :one
SELECT * FROM trees WHERE id = $1 AND owner_id = $2;

-- name: GetTreeByID :one
SELECT * FROM trees WHERE id = $1;

-- name: DeleteTree :exec
DELETE FROM trees WHERE id = $1 AND owner_id = $2;
