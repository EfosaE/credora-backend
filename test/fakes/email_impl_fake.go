package fakes

import (
	"context"
	"sync"

	"github.com/EfosaE/credora-backend/domain/email"
)

type MockEmailAdapter struct {
	mu sync.Mutex

	// Optional override
	SendEmailFunc func(ctx context.Context, req email.SendEmailRequest) error

	// Tracking
	SendEmailCalled bool
	LastRequest     email.SendEmailRequest
}

func (m *MockEmailAdapter) SendEmail(
	ctx context.Context,
	req email.SendEmailRequest,
) error {

	m.mu.Lock()
	m.SendEmailCalled = true
	m.LastRequest = req
	m.mu.Unlock()

	// If override exists, use it
	if m.SendEmailFunc != nil {
		return m.SendEmailFunc(ctx, req)
	}

	// Default safe behavior (no panic)
	return nil
}
