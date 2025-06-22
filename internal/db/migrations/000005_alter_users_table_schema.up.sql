ALTER TABLE users
ADD COLUMN nin VARCHAR(11) NOT NULL,
ADD COLUMN monnify_account_reference VARCHAR(100),
ADD COLUMN virtual_account_number VARCHAR(20),
ADD COLUMN expires_at TIMESTAMPTZ;
