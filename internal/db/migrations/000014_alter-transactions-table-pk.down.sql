-- 1️⃣ Drop the bigint id
ALTER TABLE transactions
    DROP COLUMN id;

-- 2️⃣ Recreate UUID id
ALTER TABLE transactions
    ADD COLUMN id UUID PRIMARY KEY DEFAULT gen_random_uuid();
