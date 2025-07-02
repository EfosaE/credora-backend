package accountsvc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/EfosaE/credora-backend/domain/account"
	"github.com/EfosaE/credora-backend/domain/event"
	"github.com/EfosaE/credora-backend/domain/logger"
	"github.com/EfosaE/credora-backend/internal/eventbus"
)

type AccountService struct {
	acctRepo account.AccountRepository
	logger   *logger.Logger
	eventBus eventbus.EventBus
}

func NewAccountService(acctRepo account.AccountRepository, logger *logger.Logger, eventBus eventbus.EventBus) *AccountService {
	return &AccountService{
		acctRepo: acctRepo,
		logger:   logger,
		eventBus: eventBus,
	}
}

func (a *AccountService) CreateAccount(ctx context.Context, req *account.CreateAccountRequest) (*account.Account, error) {
	acct, err := a.acctRepo.CreateAcct(ctx, req)
	if err != nil {
		return nil, err
	}
	return acct, nil
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
		_, err := a.acctRepo.CreateAcct(ctx, &account.CreateAccountRequest{
			UserId:         evt.UserID,
			AccountNumber:  evt.AccountNumber,
			AccountType:    "RESERVED ACCOUNT",
			BankName:       evt.BankName,
			MonnifyCustRef: evt.UserID.String(),
		})
		return err
	})
}
