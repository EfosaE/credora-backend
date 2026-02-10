ALTER TABLE transactions
DROP CONSTRAINT uniq_reference_direction;

ALTER TABLE transactions
DROP CONSTRAINT fk_counterparty_account;

ALTER TABLE transactions
DROP COLUMN counterparty_account_id;
