-- name: CreateAccountWithMonnify :one
INSERT INTO accounts (
    user_id,
    username,
    account_number,
    account_type,
    monnify_customer_ref,
    virtual_account_bank
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;



-- name: GetAccountByAccountNumber :one
SELECT a.id, a.account_number, a.balance, a.virtual_account_bank
FROM accounts a
WHERE a.account_number = $1
LIMIT 1;


-- name: GetUserByAccountNumber :one
SELECT u.id, u.password, u.full_name, u.email, u.phone_number,u.is_verified, a.account_number, a.account_type, a.balance, a.currency, a.virtual_account_bank
FROM accounts a
JOIN users u ON a.user_id = u.id
WHERE a.account_number = $1;


-- name: TransferMoneyInternal :one
WITH debit AS (
    UPDATE accounts
    SET balance = accounts.balance - sqlc.arg(amount)
    WHERE accounts.account_number = sqlc.arg(from_account)
      AND accounts.balance >= sqlc.arg(amount)
    RETURNING id AS from_id, balance AS from_balance
),
credit AS (
    UPDATE accounts
    SET balance = accounts.balance + sqlc.arg(amount)
    WHERE accounts.account_number = sqlc.arg(to_account)
      AND EXISTS (SELECT 1 FROM debit)
    RETURNING id AS to_id, balance AS to_balance
)
SELECT *
FROM debit, credit;



-- name: CreditAccountBalance :one
UPDATE accounts
SET balance = balance + @amount
WHERE account_number = @account_number
RETURNING id, balance;


-- name: DebitAccountBalance :one
UPDATE accounts
SET balance = balance - @amount
WHERE account_number = @account_number
RETURNING id, balance;

-- name: GetAccountsForUpdate :many
SELECT *
FROM accounts
WHERE account_number = ANY($1::text[])
ORDER BY account_number ASC
FOR UPDATE NOWAIT;

-- name: GetAccountForUpdate :one
SELECT * FROM accounts 
WHERE account_number = @account_number
FOR UPDATE NOWAIT;