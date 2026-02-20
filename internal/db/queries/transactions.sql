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

-- name: GetUserTransactionHistory :many
SELECT
  t.account_id,
  t.amount,
  t.status,
  t.description,
  t.reference,
  t.channel,
  t.meta,
  t.created_at,
  t.direction,
  t.counterparty_account_id,
  t.id
FROM transactions t
JOIN accounts a ON t.account_id = a.id
WHERE a.user_id = sqlc.arg(user_id)
  AND (
    sqlc.arg(cursor_created_at)::timestamptz IS NULL
    OR
    (t.created_at, t.id) <
    (
      sqlc.arg(cursor_created_at)::timestamptz,
      sqlc.arg(cursor_id)::bigint
    )
  )
ORDER BY t.created_at DESC, t.id DESC
LIMIT sqlc.arg(page_limit);

