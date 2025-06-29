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
