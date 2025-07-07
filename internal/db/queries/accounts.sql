-- name: CreateAccountWithMonnify :one
INSERT INTO accounts (
    user_id,
    account_number,
    account_type,
    monnify_customer_ref,
    virtual_account_bank
)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;


-- name: GetUserByAccountNumber :one
SELECT u.id, u.password, u.full_name, u.email, u.phone_number, a.account_number
FROM accounts a
JOIN users u ON a.user_id = u.id
WHERE a.account_number = $1;
