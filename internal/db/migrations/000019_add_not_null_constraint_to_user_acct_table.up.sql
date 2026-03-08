-- Add NOT NULL constraint to users.email
ALTER TABLE users
ALTER COLUMN email SET NOT NULL;

-- Add NOT NULL constraint to accounts.user_id
ALTER TABLE accounts
ALTER COLUMN user_id SET NOT NULL;