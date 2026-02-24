package mocks

import (
	"context"
	"sync"

	"github.com/EfosaE/credora-backend/domain/account"
	"github.com/EfosaE/credora-backend/domain/auth"
	"github.com/EfosaE/credora-backend/domain/user"
	"github.com/google/uuid"
)

var (
	CorrectPassword   = "pass"
	IncorrectPassword = "fail"
)

type MockTokenService struct {
	mu sync.Mutex

	GenerateTokenFunc func(ctx context.Context, payload auth.TokenPayload) (string, error)

	// Tracking
	GenerateTokenCalled bool
	LastPayload         auth.TokenPayload
}

func (m *MockTokenService) GenerateToken(
	ctx context.Context,
	payload auth.TokenPayload,
) (string, error) {

	m.mu.Lock()
	m.GenerateTokenCalled = true
	m.LastPayload = payload
	m.mu.Unlock()

	if m.GenerateTokenFunc != nil {
		return m.GenerateTokenFunc(ctx, payload)
	}

	// Default behavior (safe fallback)
	return "mock-token", nil
}

type AuthSvcMockDeps struct {
	TxManager         *MockTxManager
	UserRepo          *MockUserRepo
	PasswordResetRepo *MockPasswordResetRepo
	AcctRepo          *MockAcctRepo
	TokenService      *MockTokenService
	Mailer            *MockEmailService
}

func NewAuthSvcMockDeps() *AuthSvcMockDeps {
	return &AuthSvcMockDeps{
		TxManager:         &MockTxManager{},
		UserRepo:          &MockUserRepo{Users: make(map[uuid.UUID]*user.User)},
		PasswordResetRepo: &MockPasswordResetRepo{},
		AcctRepo:          &MockAcctRepo{Accounts: make(map[string]*account.Account)},
		TokenService:      &MockTokenService{},
		Mailer:            &MockEmailService{},
	}
}
