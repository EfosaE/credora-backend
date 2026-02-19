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

-- name: UpdateUserFullNameAndEmail :one
UPDATE users
SET full_name = $2, email = $3
WHERE id = $1
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users
SET password = $2
WHERE id = $1;



-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;