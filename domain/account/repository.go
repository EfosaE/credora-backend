package account

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// AccountRepository can start transactions and do regular queries.
type AccountRepository interface {
	FindAccountByOwner(ctx context.Context, userID uuid.UUID, accountNumber string) (*Account, error)
	CreateAcct(ctx context.Context, req *CreateAccountRequest) ([]*Account, error)
	GetAccountWithUserInfoByAcctNum(ctx context.Context, accountNumber string) (*GetAccountWithUserInfo, error)
	CreditAccount(ctx context.Context, amount decimal.Decimal, accountNumber string) (*CreditAcctResp, error)
	DebitAccount(ctx context.Context, amount decimal.Decimal, accountNumber string) (*CreditAcctResp, error)
	GetAccountsDetails(ctx context.Context, accountNumbers []string) ([]*Account, error)
	GetAccountByAccountNumber(ctx context.Context, accountNumber string) (*Account, error)
	GetAccountsForUpdate(ctx context.Context, accountNumbers []string) ([]*Account, error)
	InternalMoneyTransfer(ctx context.Context, amount decimal.Decimal, fromAcctNum, toAcctNum string) (*InternalTransferResp, error)
}
