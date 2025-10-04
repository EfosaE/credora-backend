-- Change allowed status values to uppercase
ALTER TABLE transactions
    DROP CONSTRAINT transactions_status_check,
    ADD CONSTRAINT transactions_status_check
        CHECK (status IN ('PENDING', 'SUCCESS', 'FAILED'));