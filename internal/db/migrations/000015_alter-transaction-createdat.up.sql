-- Up Migration

ALTER TABLE transactions
ALTER COLUMN created_at
TYPE timestamptz
USING created_at AT TIME ZONE 'UTC';
