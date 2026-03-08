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

-- name: VerifyUser :exec
UPDATE users
SET is_verified = TRUE
WHERE id = $1;

-- name: GetUserWithAccountsByEmail :many
SELECT 
    u.id,
    u.password,
    u.full_name,
    u.email,
    u.phone_number,
    u.is_verified,
    a.account_number,
    a.account_type,
    a.balance,
    a.currency,
    a.virtual_account_bank
FROM users u
LEFT JOIN accounts a ON a.user_id = u.id
WHERE u.email = $1;

-- name: GetUserWithAccountsByAccountNumber :many
SELECT 
    u.id,
    u.password,
    u.full_name,
    u.email,
    u.phone_number,
    u.is_verified,
    a.id AS account_id,
    a.account_number,
    a.account_type,
    a.balance,
    a.currency,
    a.virtual_account_bank
FROM accounts acc
JOIN users u ON u.id = acc.user_id
JOIN accounts a ON a.user_id = u.id
WHERE acc.account_number = $1
ORDER BY a.created_at;


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