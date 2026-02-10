ALTER TABLE transactions
ADD COLUMN counterparty_account_id UUID,
ADD CONSTRAINT fk_counterparty_account
  FOREIGN KEY (counterparty_account_id)
  REFERENCES accounts(id)
  ON DELETE SET NULL;

ALTER TABLE transactions
ADD CONSTRAINT uniq_reference_direction
UNIQUE (reference, direction);
