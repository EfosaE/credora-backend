package accountsvc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/EfosaE/credora-backend/domain/account"
	domainerr "github.com/EfosaE/credora-backend/domain/domianerrors"
	"github.com/EfosaE/credora-backend/domain/event"
	"github.com/EfosaE/credora-backend/domain/user"
	"github.com/EfosaE/credora-backend/internal/utils"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
)

type AccountService struct {
	AcctRepo account.AccountRepository
	logger   zerolog.Logger
	EventBus event.EventBus
}

func NewAccountService(acctRepo account.AccountRepository, logger zerolog.Logger, eventBus event.EventBus) *AccountService {

	serviceLogger := logger.With().
		Str("service", "account-service").
		Logger()

	return &AccountService{
		AcctRepo: acctRepo,
		logger:   serviceLogger,
		EventBus: eventBus,
	}
}

func (a *AccountService) CreateAccount(ctx context.Context, req *account.CreateAccountRequest) ([]*account.Account, error) {

	logCtx := a.logger.With().
		Str("user_id", req.UserId.String()).
		Str("account_type", req.AccountType).
		Logger()

	logCtx.Info().Msg("creating account")

	accts, err := a.AcctRepo.CreateAcct(ctx, req)
	if err != nil {
		logCtx.Error().Err(err).Msg("failed to create account")
		return nil, err
	}

	logCtx.Info().Msg("account created successfully")
	return accts, nil
}

func (a *AccountService) FindUserByAccountNumber(ctx context.Context, acctNum string) (*user.User, error) {

	logCtx := a.logger.With().Str("account_number", acctNum).Logger()
	logCtx.Info().Msg("fetching user by account number")

	acct, err := a.AcctRepo.GetAccountWithUserInfoByAcctNum(ctx, acctNum)
	if err != nil {
		logCtx.Error().Err(err).Msg("failed to find user by account number")
		return nil, err
	}

	balance, err := decimal.NewFromString(acct.Balance)
	if err != nil {
		return nil, err
	}

	return &user.User{
		ID:       acct.UserId,
		Email:    acct.Email,
		FullName: acct.FullName,
		Balance:  balance,
	}, nil
}

func (a *AccountService) FindAccountByAcctNum(ctx context.Context, acctNum string) (*account.Account, error) {

	logCtx := a.logger.With().Str("account_number", acctNum).Logger()
	logCtx.Info().Msg("fetching account by account number")

	acct, err := a.AcctRepo.GetAccountByAccountNumber(ctx, acctNum)
	if err != nil {
		logCtx.Error().Err(err).Msg("failed to find account")
		return nil, err
	}

	logCtx.Info().Msg("account retrieved successfully")
	return acct, nil
}

func (a *AccountService) FindAccountByOwner(ctx context.Context, userID uuid.UUID, accountNumber string) (*account.Account, error) {
	logCtx := a.logger.With().
		Str("user_id", userID.String()).
		Str("account_number", accountNumber).
		Logger()

	acct, err := a.AcctRepo.FindAccountByOwner(ctx, userID, accountNumber)
	if err != nil {
		if errors.Is(err, domainerr.ErrAccountNotFound) {
			logCtx.Warn().Msg("account not found or does not belong to user")
			return nil, domainerr.ErrAccountNotFound
		}
		logCtx.Error().Err(err).Msg("failed to find account by owner")
		return nil, fmt.Errorf("unable to find account by owner: %w", err)
	}

	logCtx.Debug().Msg("account ownership verified successfully")
	return acct, nil
}

func (a *AccountService) CreditUserBalance(ctx context.Context, amount decimal.Decimal, acctNum string) (*account.CreditAcctResp, error) {

	logCtx := a.logger.With().
		Str("account_number", acctNum).
		Str("amount", amount.String()).
		Logger()

	logCtx.Info().Msg("crediting user balance")

	result, err := a.AcctRepo.CreditAccount(ctx, amount, acctNum)
	if err != nil {
		logCtx.Error().Err(err).Msg("failed to credit user balance")
		return nil, err
	}

	logCtx.Info().Msg("user balance credited successfully")
	return result, nil
}

func (a *AccountService) SubscribeToUserCreatedEvents(ctx context.Context) error {

	consumer := utils.WorkerID("account")

	return a.EventBus.Subscribe(
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
				a.logger.Error().
					Err(err).
					Str("event_type", event.EventUserCreated).
					Msg("failed to decode user.created event")
				return fmt.Errorf("failed to decode %s event: %w", event.EventUserCreated, err)
			}

			logCtx := a.logger.With().
				Str("user_id", evt.UserID.String()).
				Str("user_name", evt.Name).
				Logger()

			_, err := a.AcctRepo.CreateAcct(ctx, &account.CreateAccountRequest{
				UserId:         evt.UserID,
				Username:       evt.Name,
				Accounts:       evt.Accounts,
				AccountType:    "RESERVED ACCOUNT",
				MonnifyCustRef: evt.UserID.String(),
			})

			if err != nil {
				logCtx.Error().
					Err(err).
					Msg("account recording failed")
				return fmt.Errorf("account creation failed: %w", err)
			}

			logCtx.Info().Msg("account recorded successfully")
			return nil
		},
	)
}
