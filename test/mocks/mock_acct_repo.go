package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/EfosaE/credora-backend/domain/account"
	"github.com/shopspring/decimal"
)

type MockAccountRepository struct {
	mock.Mock
}

func (m *MockAccountRepository) CreateAcct(
	ctx context.Context,
	req *account.CreateAccountRequest,
) ([]*account.Account, error) {

	args := m.Called(ctx, req)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]*account.Account), args.Error(1)
}

func (m *MockAccountRepository) CreditAccount(
	ctx context.Context,
	amount decimal.Decimal,
	accountNumber string,
) (*account.CreditAcctResp, error) {

	args := m.Called(ctx, amount, accountNumber)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*account.CreditAcctResp), args.Error(1)
}

func (m *MockAccountRepository) DebitAccount(
	ctx context.Context,
	amount decimal.Decimal,
	accountNumber string,
) (*account.CreditAcctResp, error) {

	args := m.Called(ctx, amount, accountNumber)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*account.CreditAcctResp), args.Error(1)
}

func (m *MockAccountRepository) GetAccountsDetails(
	ctx context.Context,
	accountNumbers []string,
) ([]*account.Account, error) {

	args := m.Called(ctx, accountNumbers)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]*account.Account), args.Error(1)
}

func (m *MockAccountRepository) GetAccountByAccountNumber(
	ctx context.Context,
	accountNumber string,
) (*account.Account, error) {

	args := m.Called(ctx, accountNumber)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*account.Account), args.Error(1)
}

func (m *MockAccountRepository) GetAccountsForUpdate(
	ctx context.Context,
	accountNumbers []string,
) ([]*account.Account, error) {

	args := m.Called(ctx, accountNumbers)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]*account.Account), args.Error(1)
}

func (m *MockAccountRepository) GetAccountWithUserInfoByAcctNum(
	ctx context.Context,
	accountNumbers string,
) (*account.GetAccountWithUserInfo, error) {

	args := m.Called(ctx, accountNumbers)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*account.GetAccountWithUserInfo), args.Error(1)
}

func (m *MockAccountRepository) InternalMoneyTransfer(
	ctx context.Context,
	amount decimal.Decimal,
	fromAcctNum,
	toAcctNum string,
) (*account.InternalTransferResp, error) {

	args := m.Called(ctx, amount, fromAcctNum, toAcctNum)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*account.InternalTransferResp), args.Error(1)
}

var _ account.AccountRepository = (*MockAccountRepository)(nil)
