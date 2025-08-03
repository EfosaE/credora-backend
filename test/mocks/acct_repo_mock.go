package mocks

import (
	"context"

	"github.com/EfosaE/credora-backend/domain/account"
	"github.com/EfosaE/credora-backend/internal/db/sqlc"
	"github.com/shopspring/decimal"
)

// type AccountRepository interface {
// 	CreateAcct(ctx context.Context, req *CreateAccountRequest) (*Account, error)
// 	GetUserByAccountNumber(ctx context.Context, accountNumber string) (*sqlc.GetUserByAccountNumberRow, error)
// 	CreditAccount(ctx context.Context, amount decimal.Decimal, accountNumber string) (*CreditAcctResp, error)
// }

type MockAcctRepo struct {
	CreateFunc func(ctx context.Context, req *account.CreateAccountRequest) (*account.Account, error)
	CreditFunc func(ctx context.Context, amount decimal.Decimal, accountNumber string) (*account.CreditAcctResp, error)
	Accounts   map[int]*account.Account
}

func (m *MockAcctRepo) CreateAcct(ctx context.Context, req *account.CreateAccountRequest) (*account.Account, error) {
	return m.CreateFunc(ctx, req)
}

func (m *MockAcctRepo) GetUserByAccountNumber(ctx context.Context, accountNumber string) (*sqlc.GetUserByAccountNumberRow, error) {
	return nil, nil // implement if needed later
}

func (m *MockAcctRepo) CreditAccount(ctx context.Context, amount decimal.Decimal, accountNumber string) (*account.CreditAcctResp, error) {
	return m.CreditFunc(ctx, amount, accountNumber)
}
