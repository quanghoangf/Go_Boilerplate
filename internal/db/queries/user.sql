-- name: GetUserByUserId :one
SELECT * FROM users WHERE user_id = $1;

-- name: CreateUser :exec
INSERT INTO users (user_id, nickname, password, email, created_at, updated_at)
VALUES ($1, $2, $3, $4, NOW(), NOW());
