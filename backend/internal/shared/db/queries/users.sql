-- name: GetUser :one
SELECT id, email, first_name, last_name, avatar_url, is_verified, created_at, updated_at, last_login_at, deactivated_at
FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, first_name, last_name, avatar_url, is_verified, created_at, updated_at, last_login_at, deactivated_at
FROM users
WHERE email = $1 LIMIT 1;

-- name: ListUsers :many
SELECT id, email, first_name, last_name, avatar_url, is_verified, created_at, updated_at, last_login_at, deactivated_at
FROM users
ORDER BY created_at
LIMIT $1 OFFSET $2;

-- name: CreateUser :one
INSERT INTO users (
    email, password_hash, first_name, last_name, avatar_url
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING id, email, first_name, last_name, avatar_url, is_verified, created_at, updated_at, last_login_at, deactivated_at;

-- name: UpdateUser :one
UPDATE users SET
    email = $1,
    first_name = $2,
    last_name = $3,
    avatar_url = $4,
    updated_at = NOW()
WHERE id = $5
RETURNING id, email, first_name, last_name, avatar_url, is_verified, created_at, updated_at, last_login_at, deactivated_at;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;