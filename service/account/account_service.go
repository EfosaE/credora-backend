package accountsvc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/EfosaE/credora-backend/domain/account"
	"github.com/EfosaE/credora-backend/domain/event"
	"github.com/EfosaE/credora-backend/domain/user"
	"github.com/EfosaE/credora-backend/internal/utils"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
)

type AccountService struct {
	AcctRepo account.AccountRepository
	logger   zerolog.Logger
	eventBus event.EventBus
}

func NewAccountService(acctRepo account.AccountRepository, logger zerolog.Logger, eventBus event.EventBus) *AccountService {
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
		ID:      acct.UserId,
		Email:   acct.Email,
		Name:    acct.FullName,
		Balance: acct.Balance,
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

	consumer := utils.WorkerID("account")

	return a.eventBus.Subscribe(
		ctx,
		event.StreamUserEvents,
		"account-service-group",
		consumer,
		func(ctx context.Context, msg event.EventMessage) error {

			if msg.EventType != event.EventUserCreated {
				return nil
			}

			var evt event.UserCreatedEvent
			if err := json.Unmarshal([]byte(msg.Data), &evt); err != nil {
				return fmt.Errorf("failed to decode %s event: %w",
					event.EventUserCreated,
					err,
				)
			}

			_, err := a.AcctRepo.CreateAcct(ctx, &account.CreateAccountRequest{
				UserId:         evt.UserID,
				AccountNumber:  evt.AccountNumber,
				AccountType:    "RESERVED ACCOUNT",
				BankName:       evt.BankName,
				MonnifyCustRef: evt.UserID.String(),
			})

			if err != nil {
				return fmt.Errorf("account creation failed: %w", err)
			}

			return nil
		},
	)
}
