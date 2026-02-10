-- name: RecordNewTransaction :one
INSERT INTO transactions (
    account_id,
    counterparty_account_id,
    amount,
    direction,
    status,
    description,
    reference,
    channel,
    meta
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

