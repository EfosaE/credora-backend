ALTER TABLE users
DROP CONSTRAINT password_hash_not_empty;
-- This migration removes the constraint that requires the password_hash field to be non-empty.