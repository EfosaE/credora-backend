package accountsvc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/EfosaE/credora-backend/domain/account"
	"github.com/EfosaE/credora-backend/domain/event"
	"github.com/EfosaE/credora-backend/domain/logger"
	"github.com/EfosaE/credora-backend/domain/user"
	"github.com/EfosaE/credora-backend/internal/eventbus"
	"github.com/shopspring/decimal"
)

type AccountService struct {
	AcctRepo account.AccountRepository
	logger   *logger.Logger
	eventBus eventbus.EventBus
}

func NewAccountService(acctRepo account.AccountRepository, logger *logger.Logger, eventBus eventbus.EventBus) *AccountService {
	return &AccountService{
		AcctRepo: acctRepo,
		logger:   logger,
		eventBus: eventBus,
	}
}

func (a *AccountService) CreateAccount(ctx context.Context, req *account.CreateAccountRequest) (*account.Account, error) {
	acct, err := a.AcctRepo.CreateAcct(ctx, req)
	if err != nil {
		return nil, err
	}
	return acct, nil
}

func (a *AccountService) FindUserByAccount(ctx context.Context, acctNum string) (*user.User, error) {
	acct, err := a.AcctRepo.GetUserByAccountNumber(ctx, acctNum)
	if err != nil {
		return nil, err
	}
	return &user.User{
		ID:    acct.ID,
		Email: acct.Email.String,
		Name:  acct.FullName,
	}, nil
}

func (a *AccountService) FindAccountByAcctNum(ctx context.Context, acctNum string) (*account.Account, error) {
	acct, err := a.AcctRepo.GetAccountByAccountNumber(ctx, acctNum)
	if err != nil {
		return nil, err
	}
	return acct, nil
}

func (a *AccountService) CreditUserBalance(ctx context.Context, amount decimal.Decimal, acctNum string) (*account.CreditAcctResp, error) {
	result, err := a.AcctRepo.CreditAccount(ctx, amount, acctNum)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *AccountService) SubscribeToUserCreatedEvents(ctx context.Context) error {
	return a.eventBus.Subscribe(ctx, "user.created", "account-service-group", "account-service-instance", func(values map[string]any) error {
		raw, ok := values["data"].(string)
		if !ok {
			fmt.Println("❌ invalid event payload: no 'data'")
			return errors.New("❌ invalid event payload: no 'data'")
		}

		var evt event.UserCreatedEvent
		if err := json.Unmarshal([]byte(raw), &evt); err != nil {
			fmt.Println("❌ failed to decode event:", err)
			return fmt.Errorf("❌ failed to decode event:%s", err)
		}

		// Store user ID in accounts table
		_, err := a.AcctRepo.CreateAcct(ctx, &account.CreateAccountRequest{
			UserId:         evt.UserID,
			AccountNumber:  evt.AccountNumber,
			AccountType:    "RESERVED ACCOUNT",
			BankName:       evt.BankName,
			MonnifyCustRef: evt.UserID.String(),
		})
		return err
	})
}
