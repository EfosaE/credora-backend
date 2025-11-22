ALTER TABLE idempotency_keys
ADD COLUMN status TEXT NOT NULL,
ADD CONSTRAINT idempotency_keys_status_check
CHECK (status IN ('PENDING', 'PROCESSING', 'SUCCESS', 'FAILED'));