-- name: RecordNewTransaction :one
INSERT INTO transactions (
    account_id,
    amount,
    status,
    description,
    reference,
    channel,
    meta
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;
