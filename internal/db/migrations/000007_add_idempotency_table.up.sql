CREATE TABLE idempotency_keys (
    idem_key       TEXT PRIMARY KEY,      -- Unique ID for the operation
    operation_type TEXT NOT NULL,         -- internal_transfer, ext_transfer, webhook, etc.
    payload        JSONB NOT NULL,        -- Original request or webhook data
    created_at     TIMESTAMPTZ DEFAULT NOW(),
    processed_at   TIMESTAMPTZ DEFAULT NOW()
);
