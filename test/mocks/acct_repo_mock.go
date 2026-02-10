// File: mocks/mock_account_repo.go

package mocks

import (
	"context"

	"github.com/EfosaE/credora-backend/domain/account"
	"github.com/EfosaE/credora-backend/internal/db/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type MockAcctRepo struct {
	CreateFunc               func(ctx context.Context, req *account.CreateAccountRequest) (*account.Account, error)
	CreditFunc               func(ctx context.Context, amount decimal.Decimal, accountNumber string) (*account.CreditAcctResp, error)
	DebitFunc                func(ctx context.Context, amount decimal.Decimal, accountNumber string) (*account.CreditAcctResp, error)
	GetAccountForUpdateFunc  func(ctx context.Context, accountNumber string) (*account.Account, error)      // NEW
	GetAccountsForUpdateFunc func(ctx context.Context, accountNumbers []string) ([]*account.Account, error) // NEW
	GetAccountByAcctNumFunc  func(ctx context.Context, accountNumber string) (*account.Account, error)
	WithTxFunc               func(ctx context.Context, fn func(tx account.AccountTx) error) error
	Accounts                 map[int]*account.Account
	Txm                      func() pgx.Tx
}

// Tx implements account.AccountTx.
func (m *MockAcctRepo) Tx() pgx.Tx {
	return m.Txm()
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

func (m *MockAcctRepo) DebitAccount(ctx context.Context, amount decimal.Decimal, accountNumber string) (*account.CreditAcctResp, error) {
	return m.DebitFunc(ctx, amount, accountNumber)
}

func (m *MockAcctRepo) GetAccountByAccountNumber(ctx context.Context, accountNumber string) (*account.Account, error) {
	return m.GetAccountByAcctNumFunc(ctx, accountNumber)
}

// GetAccountForUpdate mocks row locking inside a transaction
func (m *MockAcctRepo) GetAccountForUpdate(ctx context.Context, accountNumber string) (*account.Account, error) {
	if m.GetAccountForUpdateFunc != nil {
		return m.GetAccountForUpdateFunc(ctx, accountNumber)
	}
	// Default behavior: return a dummy account or look up from the map
	for _, acct := range m.Accounts {
		if acct.AccountNumber == accountNumber {
			return acct, nil
		}
	}
	return nil, nil
}

func (m *MockAcctRepo) GetAccountsForUpdate(ctx context.Context, accountNumbers []string) ([]*account.Account, error) {
	if m.GetAccountsForUpdateFunc != nil {
		return m.GetAccountsForUpdateFunc(ctx, accountNumbers)
	}
	return nil, nil
}

// WithTx mocks the transactional execution
func (m *MockAcctRepo) WithTx(ctx context.Context, fn func(tx account.AccountTx) error) error {
	if m.WithTxFunc != nil {
		return m.WithTxFunc(ctx, fn)
	}
	// Default: just call the function with the mock itself as the tx
	return fn(m)
}
