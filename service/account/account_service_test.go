package accountsvc_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/EfosaE/credora-backend/domain/account"
	"github.com/EfosaE/credora-backend/domain/event"
	"github.com/EfosaE/credora-backend/domain/monnify"
	accountsvc "github.com/EfosaE/credora-backend/service/account"
	"github.com/EfosaE/credora-backend/test/fakes"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestAccountService_CoreMethods(t *testing.T) {
	ctx := context.Background()
	acctID := uuid.New()
	// userID := uuid.New()
	acctNum := "1234567890"
	amount := decimal.NewFromInt(1000)

	tests := []struct {
		name     string
		setup    func(repo *fakes.MockAcctRepo)
		exec     func(svc *accountsvc.AccountService) (any, error)
		validate func(t *testing.T, result any, err error)
	}{
		{
			name: "CreateAccount success",
			setup: func(repo *fakes.MockAcctRepo) {
				repo.CreateAcctFunc = func(ctx context.Context, req *account.CreateAccountRequest) ([]*account.Account, error) {
					return []*account.Account{{ID: acctID}}, nil
				}
			},
			exec: func(svc *accountsvc.AccountService) (any, error) {
				return svc.CreateAccount(ctx, &account.CreateAccountRequest{})
			},
			validate: func(t *testing.T, result any, err error) {
				require.NoError(t, err)
				require.NotNil(t, result)
			},
		},
		{
			name: "CreateAccount failure",
			setup: func(repo *fakes.MockAcctRepo) {
				repo.CreateAcctFunc = func(ctx context.Context, req *account.CreateAccountRequest) ([]*account.Account, error) {
					return nil, errors.New("db error")
				}
			},
			exec: func(svc *accountsvc.AccountService) (any, error) {
				return svc.CreateAccount(ctx, &account.CreateAccountRequest{})
			},
			validate: func(t *testing.T, result any, err error) {
				require.Error(t, err)
				require.Nil(t, result)
			},
		},
		// {
		// 	name: "FindUserByAccount success (mapping verified)",
		// 	setup: func(repo *fakes.MockAcctRepo) {
		// 		repo.GetUserByAccountNumberFunc = func(ctx context.Context, acc string) (*account.GetUserDetailsWithAccountRow, error) {
		// 			return &account.GetUserDetailsWithAccountRow{
		// 				UserId:   userID,
		// 				Email:    "test@email.com",
		// 				FullName: "John Doe",
		// 				Balance:  amount.String(),
		// 			}, nil
		// 		}
		// 	},
		// 	exec: func(svc *accountsvc.AccountService) (any, error) {
		// 		return svc.FindUserByAccount(ctx, acctNum)
		// 	},
		// 	validate: func(t *testing.T, result any, err error) {
		// 		require.NoError(t, err)
		// 		u := result.(*user.User)
		// 		require.Equal(t, userID, u.ID)
		// 		require.Equal(t, "test@email.com", u.Email)
		// 		require.Equal(t, "John Doe", u.Name)
		// 		require.True(t, amount.Equal(utils.MustStringToDecimal(u.Balance)))
		// 	},
		// },
		{
			name: "FindAccountByAcctNum success",
			setup: func(repo *fakes.MockAcctRepo) {
				repo.GetAccountByAccountNumberFunc = func(ctx context.Context, acc string) (*account.Account, error) {
					return &account.Account{ID: acctID}, nil
				}
			},
			exec: func(svc *accountsvc.AccountService) (any, error) {
				return svc.FindAccountByAcctNum(ctx, acctNum)
			},
			validate: func(t *testing.T, result any, err error) {
				require.NoError(t, err)
				require.NotNil(t, result)
			},
		},
		{
			name: "FindAccountByAcctNum failure",
			setup: func(repo *fakes.MockAcctRepo) {
				repo.GetAccountByAccountNumberFunc = func(ctx context.Context, acc string) (*account.Account, error) {
					return nil, errors.New("not found")
				}
			},
			exec: func(svc *accountsvc.AccountService) (any, error) {
				return svc.FindAccountByAcctNum(ctx, acctNum)
			},
			validate: func(t *testing.T, result any, err error) {
				require.Error(t, err)
				require.Nil(t, result)
			},
		},
		{
			name: "CreditUserBalance success",
			setup: func(repo *fakes.MockAcctRepo) {
				repo.CreditFunc = func(ctx context.Context, amt decimal.Decimal, acc string) (*account.CreditAcctResp, error) {
					return &account.CreditAcctResp{
						AcctId:  acctID,
						Balance: amt,
					}, nil
				}
			},
			exec: func(svc *accountsvc.AccountService) (any, error) {
				return svc.CreditUserBalance(ctx, amount, acctNum)
			},
			validate: func(t *testing.T, result any, err error) {
				require.NoError(t, err)
				require.NotNil(t, result)
			},
		},
		{
			name: "CreditUserBalance failure",
			setup: func(repo *fakes.MockAcctRepo) {
				repo.CreditFunc = func(ctx context.Context, amt decimal.Decimal, acc string) (*account.CreditAcctResp, error) {
					return nil, errors.New("credit failed")
				}
			},
			exec: func(svc *accountsvc.AccountService) (any, error) {
				return svc.CreditUserBalance(ctx, amount, acctNum)
			},
			validate: func(t *testing.T, result any, err error) {
				require.Error(t, err)
				require.Nil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &fakes.MockAcctRepo{}
			tt.setup(mockRepo)

			service := &accountsvc.AccountService{
				AcctRepo: mockRepo,
			}

			result, err := tt.exec(service)
			tt.validate(t, result, err)
		})
	}
}

func TestAccountService_SubscribeToUserCreatedEvents(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name        string
		eventType   string
		payload     any
		setupRepo   func(repo *fakes.MockAcctRepo)
		expectError bool
	}{
		{
			name:        "ignores non USER_CREATED event",
			eventType:   "SOME_OTHER_EVENT",
			payload:     map[string]any{},
			setupRepo:   func(repo *fakes.MockAcctRepo) {},
			expectError: false,
		},
		{
			name:        "fails on invalid JSON",
			eventType:   event.EventUserCreated,
			payload:     "{bad-json",
			setupRepo:   func(repo *fakes.MockAcctRepo) {},
			expectError: true,
		},
		// Update the UserCreatedEvent struct usage
		{
			name:      "successfully creates account",
			eventType: event.EventUserCreated,
			payload: event.UserCreatedEvent{
				UserID: userID,
				Name:   "John Doe",
				Accounts: []monnify.ReservedAccount{
					{
						AccountNumber: "12345",
						BankName:      "Test Bank",
						AccountName:   "John Doe",
						BankCode:      "057",
					},
				},
			},
			setupRepo: func(repo *fakes.MockAcctRepo) {
				repo.CreateAcctFunc = func(ctx context.Context, req *account.CreateAccountRequest) ([]*account.Account, error) {
					require.Equal(t, userID, req.UserId)
					require.Equal(t, "John Doe", req.Username)
					require.Equal(t, "RESERVED ACCOUNT", req.AccountType)
					require.Equal(t, userID.String(), req.MonnifyCustRef)
					// require.Len(t, req.Accounts, 1)
					require.Equal(t, "12345", req.Accounts[0].AccountNumber)
					require.Equal(t, "Test Bank", req.Accounts[0].BankName)
					return []*account.Account{{}}, nil
				}
			},
			expectError: false,
		},
		{
			name:      "repository failure bubbles up",
			eventType: event.EventUserCreated,
			payload: event.UserCreatedEvent{
				UserID: userID,
				Name:   "John Doe",
				Accounts: []monnify.ReservedAccount{
					{
						AccountNumber: "12345",
						BankName:      "Test Bank",
					},
				},
			},
			setupRepo: func(repo *fakes.MockAcctRepo) {
				repo.CreateAcctFunc = func(ctx context.Context, req *account.CreateAccountRequest) ([]*account.Account, error) {
					return nil, errors.New("db error")
				}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockRepo := &fakes.MockAcctRepo{}
			mockBus := &fakes.MockEventBus{}
			tt.setupRepo(mockRepo)

			service := &accountsvc.AccountService{
				AcctRepo: mockRepo,
				EventBus: mockBus,
			}

			//  Subscribe
			err := service.SubscribeToUserCreatedEvents(ctx)
			require.NoError(t, err)

			//  Verify subscription metadata
			require.Len(t, mockBus.Subscribed, 1)

			sub := mockBus.Subscribed[0]
			require.Equal(t, event.StreamUserEvents, sub.Topic)
			require.Equal(t, "account-service-group", sub.Group)
			require.NotEmpty(t, sub.Consumer)

			//  Get registered handler
			handler := mockBus.Handlers[event.StreamUserEvents]
			require.NotNil(t, handler)

			//  Prepare event message
			var data string

			switch v := tt.payload.(type) {
			case string:
				data = v
			default:
				bytes, _ := json.Marshal(v)
				data = string(bytes)
			}

			msg := event.EventMessage{
				EventType: tt.eventType,
				Data:      data,
			}

			//  Invoke handler manually
			err = handler(ctx, msg)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// func TestCreditUserBalance(t *testing.T) {
// 	mockRepo := &fakes.MockAcctRepo{}

// 	amount := decimal.NewFromFloat(1500.0)
// 	acctNum := "1234567890"
// 	acctID := uuid.New() // generate a valid UUID

// 	expectedResp := &account.CreditAcctResp{
// 		AcctId:  acctID,
// 		Balance: amount,
// 	}

// 	// Mock the CreditFunc behavior
// 	mockRepo.CreditFunc = func(ctx context.Context, amt decimal.Decimal, accNum string) (*account.CreditAcctResp, error) {
// 		require.Equal(t, amount, amt)
// 		require.Equal(t, acctNum, accNum)
// 		return expectedResp, nil
// 	}

// 	service := &accountsvc.AccountService{AcctRepo: mockRepo}

// 	resp, err := service.CreditUserBalance(context.Background(), amount, acctNum)
// 	require.NoError(t, err)
// 	require.Equal(t, expectedResp, resp)
// }
