-- Remove NOT NULL constraint from users.email
ALTER TABLE users
ALTER COLUMN email DROP NOT NULL;

-- Remove NOT NULL constraint from accounts.user_id
ALTER TABLE accounts
ALTER COLUMN user_id DROP NOT NULL;