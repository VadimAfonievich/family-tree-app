-- relations queries

-- name: CreateRelation :one
INSERT INTO relations (tree_id, person1_id, person2_id, relation_type)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetRelation :one
SELECT * FROM relations WHERE id = $1;

-- name: ListRelationsByTree :many
SELECT * FROM relations WHERE tree_id = $1 ORDER BY created_at;

-- name: DeleteRelation :exec
DELETE FROM relations WHERE id = $1;

-- name: DeleteRelationsByPerson :exec
DELETE FROM relations WHERE person1_id = $1 OR person2_id = $1;
