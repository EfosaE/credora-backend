ALTER TABLE transactions
ADD CONSTRAINT transactions_reference_key
UNIQUE (reference);
