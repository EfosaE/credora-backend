-- Revert to lowercase if needed
ALTER TABLE transactions
    DROP CONSTRAINT transactions_status_check,
    ADD CONSTRAINT transactions_status_check
        CHECK (status IN ('pending', 'success', 'failed'));