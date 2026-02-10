ALTER TABLE transactions
ADD COLUMN direction VARCHAR(10)
CHECK (direction IN ('debit', 'credit'));
