SELECT conname
FROM pg_constraint
WHERE conrelid = 'transactions'::regclass
  AND contype = 'c';
ALTER TABLE transactions
DROP CONSTRAINT transactions_direction_check;

ALTER TABLE transactions
ADD CONSTRAINT transactions_direction_check
CHECK (direction IN ('DEBIT', 'CREDIT'));
