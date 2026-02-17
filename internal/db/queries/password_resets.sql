
-- name: CreatePasswordReset :one
INSERT INTO password_resets (user_id, token_hash, expires_at, used_at, created_at)
VALUES ($1, $2, $3, $4, COALESCE($5, NOW()))
RETURNING *;

-- name: DeletePasswordReset :exec
DELETE FROM password_resets
WHERE id = $1;

-- name: UpdatePasswordResetUsedAt :exec
UPDATE password_resets
SET used_at = $2
WHERE id = $1;


-- name: GetActivePasswordReset :one
SELECT *
FROM password_resets
WHERE user_id = $1
  AND used_at IS NULL
  AND expires_at > NOW()
ORDER BY created_at DESC
LIMIT 1;
