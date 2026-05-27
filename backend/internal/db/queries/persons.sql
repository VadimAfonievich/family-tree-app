-- persons queries

-- name: CreatePerson :one
INSERT INTO persons (tree_id, first_name, last_name, gender, birth_date, death_date, photo_url)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetPerson :one
SELECT * FROM persons WHERE id = $1;

-- name: GetPersonInTree :one
SELECT * FROM persons WHERE id = $1 AND tree_id = $2;

-- name: ListPersonsByTree :many
SELECT * FROM persons WHERE tree_id = $1 ORDER BY created_at;

-- name: UpdatePerson :one
UPDATE persons
SET first_name = COALESCE($2, first_name),
    last_name = COALESCE($3, last_name),
    gender = COALESCE($4, gender),
    birth_date = COALESCE($5, birth_date),
    death_date = COALESCE($6, death_date),
    photo_url = COALESCE($7, photo_url)
WHERE id = $1
RETURNING *;

-- name: DeletePerson :exec
DELETE FROM persons WHERE id = $1;
