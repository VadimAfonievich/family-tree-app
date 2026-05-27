-- users queries

-- name: CreateUser :one
INSERT INTO users (telegram_id, username)
VALUES ($1, $2)
ON CONFLICT (telegram_id) DO UPDATE SET username = EXCLUDED.username
RETURNING *;

-- name: GetUserByTelegramID :one
SELECT * FROM users WHERE telegram_id = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;
