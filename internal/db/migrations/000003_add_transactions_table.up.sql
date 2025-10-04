-- transactions table
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID REFERENCES accounts(id) ON DELETE CASCADE,
    amount DECIMAL(12, 2) NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'success', 'failed')),
    description TEXT,
    reference VARCHAR(100) UNIQUE, -- Monnify or internal reference
    channel VARCHAR(50), -- e.g., 'Monnify', 'manual', 'internal'
    meta JSONB, -- Extra Monnify data
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
