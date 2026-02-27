-- name: GetDeviceTokensByUserID :many
SELECT * FROM device_tokens WHERE user_id = $1;

-- name: CreateDeviceToken :one
INSERT INTO device_tokens (user_id, token, platform)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateDeviceToken :one
UPDATE device_tokens
SET token = $1, platform = $2
WHERE id = $3
RETURNING *;

-- name: DeleteDeviceToken :exec
DELETE FROM device_tokens WHERE id = $1;

-- name: DeleteDeviceTokensByUserID :exec
DELETE FROM device_tokens WHERE user_id = $1;