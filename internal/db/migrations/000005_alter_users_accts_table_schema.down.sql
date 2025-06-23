ALTER TABLE accounts
    DROP COLUMN virtual_account_bank,
    DROP COLUMN monnify_customer_ref,
    ADD COLUMN paystack_recipient_code VARCHAR(50);
ALTER TABLE users
    DROP COLUMN nin,
    DROP COLUMN expires_at,
    DROP COLUMN monnify_customer_ref,
    ADD COLUMN paystack_customer_id VARCHAR(100);
