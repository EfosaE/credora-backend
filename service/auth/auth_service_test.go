package authsvc_test

import (
	"context"
	"errors"

	// "errors"
	"testing"

	"github.com/EfosaE/credora-backend/domain/account"
	"github.com/EfosaE/credora-backend/domain/auth"
	"github.com/EfosaE/credora-backend/domain/monnify"
	"github.com/EfosaE/credora-backend/test"
	"github.com/EfosaE/credora-backend/test/fakes"

	// "github.com/EfosaE/credora-backend/domain/auth"
	// "github.com/EfosaE/credora-backend/domain/user"
	authsvc "github.com/EfosaE/credora-backend/service/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestAuthService_Login(t *testing.T) {

	tests := []struct {
		name        string
		setup       func(*fakes.AuthSvcMockDeps)
		expectError bool
	}{
		{
			name: "success",
			setup: func(m *fakes.AuthSvcMockDeps) {
				hashed, _ := authsvc.HashPassword(fakes.CorrectPassword)

				// Override the default GetUserByAccountNumber method because we use hashed passwords and we didnt set the 11111111 in the accounts map for the login
				m.AcctRepo.GetUserByAccountNumberFunc = func(ctx context.Context, acctNo string) (*account.GetUserDetailsWithAccountRow, error) {
					return &account.GetUserDetailsWithAccountRow{
						UserId:        uuid.New(),
						AccountNumber: acctNo,
						FullName:      "John Doe",
						Password:      hashed,
					}, nil
				}

				m.TokenService.GenerateTokenFunc = func(ctx context.Context, payload auth.TokenPayload) (string, error) {
					return "mock-token", nil
				}
			},
			expectError: false,
		},
		{
			name: "account not found",
			setup: func(m *fakes.AuthSvcMockDeps) {
				m.AcctRepo.GetUserByAccountNumberFunc = func(ctx context.Context, accountNumber string) (*account.GetUserDetailsWithAccountRow, error) {
					return nil, errors.New("not found")
				}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			userID := uuid.New()

			m := fakes.NewAuthSvcMockDeps()
			testLogger := test.SetupTestLogger()
			tt.setup(m)

			m.AcctRepo.CreateAcct(ctx, &account.CreateAccountRequest{
				UserId:         userID,
				Username:       "John Doe",
				AccountType:    "savings",
				MonnifyCustRef: "auth_ref",
				Accounts: []monnify.ReservedAccount{
					{
						AccountNumber: "1111111111",
						BankName:      "Zenith",
						AccountName:   "John Doe",
						BankCode:      "057",
					},
				},
			})

			service := authsvc.NewAuthService(
				m.TxManager,
				m.UserRepo,
				m.PasswordResetRepo,
				m.TokenService,
				m.AcctRepo,
				m.Mailer,
				testLogger,
			)

			_, _, err := service.Login(ctx, "1111111111", fakes.CorrectPassword)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
