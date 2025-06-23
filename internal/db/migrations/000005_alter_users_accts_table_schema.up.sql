ALTER TABLE accounts
    ADD COLUMN virtual_account_bank TEXT,
    ADD COLUMN monnify_customer_ref TEXT,
    DROP COLUMN paystack_recipient_code;




ALTER TABLE users
    ADD COLUMN nin VARCHAR(11) UNIQUE NOT NULL,
    ADD COLUMN expires_at TIMESTAMP DEFAULT (NOW() + INTERVAL '48 hours'),
    ADD COLUMN monnify_customer_ref TEXT,
    DROP COLUMN paystack_customer_id;
