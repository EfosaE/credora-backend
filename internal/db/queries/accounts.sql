-- name: CreateAccountWithMonnify :one
INSERT INTO accounts (
    user_id,
    account_number,
    account_type,
    monnify_customer_ref,
    virtual_account_bank
)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;



-- name: GetAccountByAccountNumber :one
SELECT a.id, a.account_number, a.balance, a.virtual_account_bank
FROM accounts a
WHERE a.account_number = $1
LIMIT 1;


-- name: GetUserByAccountNumber :one
SELECT u.id, u.password, u.full_name, u.email, u.phone_number, a.account_number, a.balance, a.virtual_account_bank
FROM accounts a
JOIN users u ON a.user_id = u.id
WHERE a.account_number = $1;


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

-- name: GetAccountForUpdate :one
SELECT *
FROM accounts
WHERE account_number = @account_number
FOR UPDATE;
