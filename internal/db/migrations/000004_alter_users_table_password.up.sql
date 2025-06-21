ALTER TABLE users
ADD CONSTRAINT password_not_empty CHECK (length(password) > 0);
