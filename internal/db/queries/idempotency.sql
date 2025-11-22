-- name: CheckIdempotency :one
SELECT EXISTS (
    SELECT 1 FROM idempotency_keys
    WHERE idem_key = $1
) AS exists;



-- name: InsertIdempotencyKey :exec
INSERT INTO idempotency_keys (idem_key, operation_type, payload, status)
VALUES ($1, $2, $3, $4);



-- -- name: UpsertIdempotencyKey :exec
-- INSERT INTO idempotency_keys (idem_key, operation_type, payload, status)
-- VALUES ($1, $2, $3, $4)
-- ON CONFLICT (idem_key) DO NOTHING;

-- name: UpsertIdempotencyKey :exec
INSERT INTO idempotency_keys (idem_key, operation_type, payload, status)
VALUES ($1, $2, $3, $4)
ON CONFLICT (idem_key)
DO UPDATE SET
    operation_type = EXCLUDED.operation_type,
    payload = EXCLUDED.payload,
    status = EXCLUDED.status;





-- name: GetIdempotencyKey :one
SELECT *
FROM idempotency_keys
WHERE idem_key = $1;



-- name: UpdateIdempotencyStatus :exec
UPDATE idempotency_keys
SET status = $2,
    updated_at = NOW()
WHERE idem_key = $1;



-- name: SaveIdempotencySuccess :exec
UPDATE idempotency_keys
SET status = 'SUCCESS',
    payload = $2,
    updated_at = NOW()
WHERE idem_key = $1;



-- name: SaveIdempotencyFailure :exec
UPDATE idempotency_keys
SET status = 'FAILED',
    payload = $2,
    updated_at = NOW()
WHERE idem_key = $1;



-- name: DeleteIdempotencyKey :exec
DELETE FROM idempotency_keys
WHERE idem_key = $1;
