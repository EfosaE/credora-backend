ALTER TABLE users
ADD CONSTRAINT password_hash_not_empty CHECK (length(password_hash) > 0);
