package accountsvc

import (
	"context"

	"github.com/EfosaE/credora-backend/domain/account"
	"github.com/EfosaE/credora-backend/domain/logger"
)

type AccountService struct {
	acctRepo account.AccountRepository
	logger   *logger.Logger
}

func NewAccountService(acctRepo account.AccountRepository, logger *logger.Logger) *AccountService {
	return &AccountService{
		acctRepo: acctRepo,
		logger:   logger,
	}
}

func (a *AccountService) CreateAccount(ctx context.Context, req *account.CreateAccountRequest) (*account.Account, error) {
	acct, err := a.acctRepo.CreateAcct(ctx, req)
	if err != nil {
		return nil, err
	}
	return acct, nil
}
