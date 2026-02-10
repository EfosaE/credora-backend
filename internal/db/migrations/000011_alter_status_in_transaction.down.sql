ALTER TABLE transactions
DROP CONSTRAINT transactions_direction_check;

ALTER TABLE transactions
ADD CONSTRAINT transactions_direction_check
CHECK (direction IN ('debit', 'credit'));
