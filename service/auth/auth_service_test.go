package authsvc_test

import (
	"context"
	"errors"
	"testing"

	"github.com/EfosaE/credora-backend/domain/account"
	"github.com/EfosaE/credora-backend/domain/auth"
	domainerr "github.com/EfosaE/credora-backend/domain/domianerrors"
	"github.com/EfosaE/credora-backend/domain/user"
	authsvc "github.com/EfosaE/credora-backend/service/auth"
	"github.com/EfosaE/credora-backend/test"
	"github.com/EfosaE/credora-backend/test/fakes"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func cloneUser(u *user.User) *user.User {
	copy := *u
	return &copy
}

func TestAuthService_Login(t *testing.T) {
	hashedCorrectPassword, _ := authsvc.HashPassword(fakes.CorrectPassword)

	userID := uuid.New()
	baseUser := &user.User{
		ID:         userID,
		FullName:   "John Doe",
		Email:      "john@example.com",
		Password:   hashedCorrectPassword,
		IsVerified: false,
	}

	verifiedUser := &user.User{
		ID:         userID,
		FullName:   "John Doe",
		Email:      "john@example.com",
		Password:   hashedCorrectPassword,
		IsVerified: true,
	}

	tests := []struct {
		name        string
		identifier  string
		password    string
		setup       func(*fakes.MockUserRepo, *fakes.MockTokenService)
		expectError bool
		expectedErr error
		checkResult func(t *testing.T, u *user.User, token string)
	}{
		// ── Account number login ──────────────────────────────────────────────
		{
			name:       "account number login: success + first login activates user",
			identifier: "1111111111",
			password:   fakes.CorrectPassword,
			setup: func(ur *fakes.MockUserRepo, ts *fakes.MockTokenService) {
				baseUser.Password = hashedCorrectPassword
				baseUser.Accounts = []account.Account{
					{AccountNumber: "1111111111"},
				}
				if ur.Users == nil {
					ur.Users = make(map[uuid.UUID]*user.User)
				}
				ur.Users[baseUser.ID] = baseUser

				ts.GenerateTokenFunc = func(_ context.Context, _ auth.TokenPayload) (string, error) {
					return "mock-token", nil
				}
			},
			expectError: false,
			checkResult: func(t *testing.T, u *user.User, token string) {
				require.Equal(t, "mock-token", token)
				require.True(t, u.IsVerified, "user should be activated on first login")
			},
		},
		{
			name:       "account number login: success, already verified user",
			identifier: "1111111111",
			password:   fakes.CorrectPassword,
			setup: func(ur *fakes.MockUserRepo, ts *fakes.MockTokenService) {
				ur.GetUserAccountsByAccountNumberFunc = func(_ context.Context, _ string) (*user.User, error) {
					return verifiedUser, nil
				}
				ts.GenerateTokenFunc = func(_ context.Context, _ auth.TokenPayload) (string, error) {
					return "mock-token", nil
				}
			},
			expectError: false,
			checkResult: func(t *testing.T, u *user.User, token string) {
				require.Equal(t, "mock-token", token)
			},
		},
		{
			name:       "account number login: account not found",
			identifier: "0000000000",
			password:   fakes.CorrectPassword,
			setup: func(ur *fakes.MockUserRepo, ts *fakes.MockTokenService) {
				ur.GetUserAccountsByAccountNumberFunc = func(_ context.Context, _ string) (*user.User, error) {
					return nil, errors.New("not found")
				}
			},
			expectError: true,
			expectedErr: domainerr.ErrInvalidCredentials,
		},
		{
			name:       "account number login: wrong password",
			identifier: "1111111111",
			password:   "wrongpassword",
			setup: func(ur *fakes.MockUserRepo, ts *fakes.MockTokenService) {
				ur.GetUserAccountsByAccountNumberFunc = func(_ context.Context, _ string) (*user.User, error) {
					return baseUser, nil
				}
			},
			expectError: true,
			expectedErr: domainerr.ErrInvalidCredentials,
		},
		{
			name:       "account number login: VerifyUser fails on first login",
			identifier: "1111111111",
			password:   fakes.CorrectPassword,
			setup: func(ur *fakes.MockUserRepo, ts *fakes.MockTokenService) {
				freshUser := cloneUser(baseUser)
				freshUser.IsVerified = false
				ur.GetUserAccountsByAccountNumberFunc = func(_ context.Context, _ string) (*user.User, error) {
					return freshUser, nil
				}
				ur.VerifyUserFunc = func(_ context.Context, _ uuid.UUID) error {
					return errors.New("db error")
				}
			},
			expectError: true,
		},
		{
			name:       "account number login: token generation fails",
			identifier: "1111111111",
			password:   fakes.CorrectPassword,
			setup: func(ur *fakes.MockUserRepo, ts *fakes.MockTokenService) {
				ur.GetUserAccountsByAccountNumberFunc = func(_ context.Context, _ string) (*user.User, error) {
					return verifiedUser, nil
				}
				ts.GenerateTokenFunc = func(_ context.Context, _ auth.TokenPayload) (string, error) {
					return "", errors.New("token error")
				}
			},
			expectError: true,
		},

		// ── Email login ───────────────────────────────────────────────────────
		{
			name:       "email login: success with verified user",
			identifier: "john@example.com",
			password:   fakes.CorrectPassword,
			setup: func(ur *fakes.MockUserRepo, ts *fakes.MockTokenService) {
				ur.GetUserAccountsByEmailFunc = func(_ context.Context, _ string) (*user.User, error) {
					return verifiedUser, nil
				}
				ts.GenerateTokenFunc = func(_ context.Context, _ auth.TokenPayload) (string, error) {
					return "mock-token", nil
				}
			},
			expectError: false,
			checkResult: func(t *testing.T, u *user.User, token string) {
				require.Equal(t, "mock-token", token)
			},
		},
		{
			name:       "email login: rejected if user not verified",
			identifier: "john@example.com",
			password:   fakes.CorrectPassword,
			setup: func(ur *fakes.MockUserRepo, ts *fakes.MockTokenService) {
				freshUser := cloneUser(baseUser)
				freshUser.IsVerified = false
				ur.GetUserAccountsByEmailFunc = func(_ context.Context, _ string) (*user.User, error) {
					return freshUser, nil
				}
			},
			expectError: true,
			expectedErr: domainerr.ErrAccountNotActivated,
		},
		{
			name:       "email login: user not found",
			identifier: "ghost@example.com",
			password:   fakes.CorrectPassword,
			setup: func(ur *fakes.MockUserRepo, ts *fakes.MockTokenService) {
				ur.GetUserAccountsByEmailFunc = func(_ context.Context, _ string) (*user.User, error) {
					return nil, errors.New("not found")
				}
			},
			expectError: true,
			expectedErr: domainerr.ErrInvalidCredentials,
		},
		{
			name:       "email login: wrong password",
			identifier: "john@example.com",
			password:   "wrongpassword",
			setup: func(ur *fakes.MockUserRepo, ts *fakes.MockTokenService) {
				ur.GetUserAccountsByEmailFunc = func(_ context.Context, _ string) (*user.User, error) {
					return verifiedUser, nil
				}
			},
			expectError: true,
			expectedErr: domainerr.ErrInvalidCredentials,
		},
		{
			name:       "email login: token generation fails",
			identifier: "john@example.com",
			password:   fakes.CorrectPassword,
			setup: func(ur *fakes.MockUserRepo, ts *fakes.MockTokenService) {
				ur.GetUserAccountsByEmailFunc = func(_ context.Context, _ string) (*user.User, error) {
					return verifiedUser, nil
				}
				ts.GenerateTokenFunc = func(_ context.Context, _ auth.TokenPayload) (string, error) {
					return "", errors.New("token error")
				}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			m := fakes.NewAuthSvcMockDeps()
			testLogger := test.SetupTestLogger()

			tt.setup(m.UserRepo, m.TokenService)

			service := authsvc.NewAuthService(
				m.TxManager,
				m.UserRepo,
				m.PasswordResetRepo,
				m.TokenService,
				m.AcctRepo,
				m.Mailer,
				testLogger,
			)

			u, token, err := service.Login(ctx, tt.identifier, tt.password)

			if tt.expectError {
				require.Error(t, err)
				if tt.expectedErr != nil {
					require.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, token)
				if tt.checkResult != nil {
					tt.checkResult(t, u, token)
				}
			}
		})
	}
}
