-- name: CreateUser :one
INSERT INTO users (email, full_name, phone_number, password, nin)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserByPhone :one
SELECT * FROM users
WHERE phone_number = $1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateUser :one
UPDATE users
SET full_name = $2, email = $3
WHERE id = $1
RETURNING *;


-- -- name: UpdateUser :one
-- This query is commented out because it updates manually but I have associated trigger
-- that automatically updates the `updated_at` field on any update.
-- UPDATE users
-- SET name = $2, email = $3, updated_at = NOW()
-- WHERE id = $1
-- RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;