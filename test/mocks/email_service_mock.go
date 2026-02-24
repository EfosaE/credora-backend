package mocks

import (
	"context"
	"sync"

	// "github.com/EfosaE/credora-backend/domain/email"
	// "github.com/EfosaE/credora-backend/domain/event"
	"github.com/EfosaE/credora-backend/domain/user"
)

type MockEmailService struct {
	mu sync.Mutex

	// Optional overrides
	SendAccountNumberEmailFunc func(ctx context.Context, to, bank, accountNumber string) error
	SendPasswordResetEmailFunc func(ctx context.Context, to, resetLink string) error
	SendWelcomeEmailFunc       func(ctx context.Context, user user.User) error

	// Tracking
	AccountEmailCalled bool
	ResetEmailCalled   bool
	WelcomeEmailCalled bool
}

func (m *MockEmailService) SendAccountNumberEmail(
	ctx context.Context,
	to, bank, accountNumber string,
) error {

	m.mu.Lock()
	m.AccountEmailCalled = true
	m.mu.Unlock()

	if m.SendAccountNumberEmailFunc != nil {
		return m.SendAccountNumberEmailFunc(ctx, to, bank, accountNumber)
	}

	return nil
}

func (m *MockEmailService) SendPasswordResetEmail(
	ctx context.Context,
	to, resetLink string,
) error {

	m.mu.Lock()
	m.ResetEmailCalled = true
	m.mu.Unlock()

	if m.SendPasswordResetEmailFunc != nil {
		return m.SendPasswordResetEmailFunc(ctx, to, resetLink)
	}

	return nil
}

func (m *MockEmailService) SendWelcomeEmail(
	ctx context.Context,
	u user.User,
) error {

	m.mu.Lock()
	m.WelcomeEmailCalled = true
	m.mu.Unlock()

	if m.SendWelcomeEmailFunc != nil {
		return m.SendWelcomeEmailFunc(ctx, u)
	}

	return nil
}

// import (
// 	"context"

// 	"github.com/EfosaE/credora-backend/domain/email"
// )

// type MockEmailSender struct {
// 	Sent []email.SendEmailRequest
// }

// func (m *MockEmailSender) SendEmail(
// 	ctx context.Context,
// 	req email.SendEmailRequest,
// ) error {
// 	m.Sent = append(m.Sent, req)
// 	return nil
// }
