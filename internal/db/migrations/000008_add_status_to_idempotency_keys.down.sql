ALTER TABLE idempotency_keys
DROP CONSTRAINT IF EXISTS idempotency_keys_status_check;

ALTER TABLE idempotency_keys
DROP COLUMN IF EXISTS status;
