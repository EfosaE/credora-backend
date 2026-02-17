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
	GetAccountByAccountNumber(ctx context.Context, accountNumber string) (*Account, error)
	GetAccountsForUpdate(ctx context.Context, accountNumbers []string) ([]*Account, error)
	// // Transaction wrapper
	// WithTx(ctx context.Context, fn func(tx AccountTx) error) error
}
