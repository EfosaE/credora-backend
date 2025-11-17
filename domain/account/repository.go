package account

import (
	"context"

	"github.com/EfosaE/credora-backend/internal/db/sqlc"
	"github.com/shopspring/decimal"
)

// AccountRepository can start transactions and do regular queries.
type AccountRepository interface {
	CreateAcct(ctx context.Context, req *CreateAccountRequest) (*Account, error)
	GetUserByAccountNumber(ctx context.Context, accountNumber string) (*sqlc.GetUserByAccountNumberRow, error)
	CreditAccount(ctx context.Context, amount decimal.Decimal, accountNumber string) (*CreditAcctResp, error)
	DebitAccount(ctx context.Context, amount decimal.Decimal, accountNumber string) (*CreditAcctResp, error)

	// Transaction wrapper
	WithTx(ctx context.Context, fn func(tx AccountTx) error) error
}

// AccountTx defines operations allowed within a transaction
type AccountTx interface {
	GetAccountForUpdate(ctx context.Context, accountNumber string) (*Account, error)
	CreditAccount(ctx context.Context, amount decimal.Decimal, accountNumber string) (*CreditAcctResp, error)
	DebitAccount(ctx context.Context, amount decimal.Decimal, accountNumber string) (*CreditAcctResp, error)
}
