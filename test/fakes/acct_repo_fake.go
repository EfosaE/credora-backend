package fakes

import (
	"context"
	"errors"

	"github.com/EfosaE/credora-backend/domain/account"
	"github.com/google/uuid"

	// "github.com/EfosaE/credora-backend/domain/user"
	// "github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type MockAcctRepo struct {
	CreateAcctFunc                func(ctx context.Context, req *account.CreateAccountRequest) ([]*account.Account, error)
	GetAccountByAccountNumberFunc func(ctx context.Context, accountNumber string) (*account.Account, error)
	CreditFunc                    func(ctx context.Context, amount decimal.Decimal, accountNumber string) (*account.CreditAcctResp, error)
	DebitFunc                     func(ctx context.Context, amount decimal.Decimal, accountNumber string) (*account.CreditAcctResp, error)

	Accounts map[string]*account.Account
}

func (m *MockAcctRepo) CreateAcct(
	ctx context.Context,
	req *account.CreateAccountRequest,
) ([]*account.Account, error) {

	if m.CreateAcctFunc != nil {
		return m.CreateAcctFunc(ctx, req)
	}

	if m.Accounts == nil {
		m.Accounts = make(map[string]*account.Account)
	}

	var accounts []*account.Account
	for _, a := range req.Accounts {
		acct := &account.Account{
			UserId:        req.UserId,
			AccountNumber: a.AccountNumber,
			UserName:      req.Username,
			AccountType:   req.AccountType,
			BankName:      a.BankName,
		}
		m.Accounts[a.AccountNumber] = acct
		accounts = append(accounts, acct)
	}

	return accounts, nil
}

func (m *MockAcctRepo) GetAccountByAccountNumber(
	ctx context.Context,
	accountNumber string,
) (*account.Account, error) {

	if m.GetAccountByAccountNumberFunc != nil {
		return m.GetAccountByAccountNumberFunc(ctx, accountNumber)
	}

	if m.Accounts == nil {
		return nil, errors.New("no accounts in mock")
	}

	acct, ok := m.Accounts[accountNumber]
	if !ok {
		return nil, errors.New("account not found")
	}

	return acct, nil
}

func (m *MockAcctRepo) CreditAccount(
	ctx context.Context,
	amount decimal.Decimal,
	accountNumber string,
) (*account.CreditAcctResp, error) {

	if m.CreditFunc != nil {
		return m.CreditFunc(ctx, amount, accountNumber)
	}

	acct, ok := m.Accounts[accountNumber]
	if !ok {
		return nil, errors.New("account not found")
	}

	newBalance := acct.Balance.Add(amount)
	acct.Balance = newBalance

	return &account.CreditAcctResp{
		AcctId:  acct.ID,
		Balance: newBalance,
	}, nil
}

func (m *MockAcctRepo) DebitAccount(
	ctx context.Context,
	amount decimal.Decimal,
	accountNumber string,
) (*account.CreditAcctResp, error) {

	if m.DebitFunc != nil {
		return m.DebitFunc(ctx, amount, accountNumber)
	}

	acct, ok := m.Accounts[accountNumber]
	if !ok {
		return nil, errors.New("account not found")
	}

	newBalance := acct.Balance.Sub(amount)
	acct.Balance = newBalance

	return &account.CreditAcctResp{
		AcctId:  acct.ID,
		Balance: newBalance,
	}, nil
}

// GetAccountForUpdate fakes row locking inside a transaction
func (m *MockAcctRepo) GetAccountForUpdate(ctx context.Context, accountNumber string) (*account.Account, error) {
	// if m.GetAccountForUpdateFunc != nil {
	// 	return m.GetAccountForUpdateFunc(ctx, accountNumber)
	// }
	// Default behavior: return a dummy account or look up from the map
	for _, acct := range m.Accounts {
		if acct.AccountNumber == accountNumber {
			return acct, nil
		}
	}
	return nil, nil
}

func (m *MockAcctRepo) GetAccountsForUpdate(ctx context.Context, accountNumbers []string) ([]*account.Account, error) {
	// if m.GetAccountsForUpdateFunc != nil {
	// 	return m.GetAccountsForUpdateFunc(ctx, accountNumbers)
	// }
	return nil, nil
}
func (m *MockAcctRepo) GetAccountsDetails(ctx context.Context, accountNumbers []string) ([]*account.Account, error) {
	// if m.GetAccountsForUpdateFunc != nil {
	// 	return m.GetAccountsForUpdateFunc(ctx, accountNumbers)
	// }
	return nil, nil
}
func (m *MockAcctRepo) InternalMoneyTransfer(ctx context.Context, amount decimal.Decimal, fromAcctNum, toAcctNum string) (*account.InternalTransferResp, error) {
	//for now, just return nil
	return nil, nil
}

func (m *MockAcctRepo) GetAccountWithUserInfoByAcctNum(ctx context.Context, accountNumber string) (*account.GetAccountWithUserInfo, error) {
	//for now, just return nil
	return nil, nil
}

func (m *MockAcctRepo) FindAccountByOwner(ctx context.Context, userID uuid.UUID, accountNumber string) (*account.Account, error) {
	//for now, just return nil
	return nil, nil
}
