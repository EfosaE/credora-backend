package mocks

import (
	"context"

	"github.com/EfosaE/credora-backend/domain/email"
)

type MockEmailAdapter struct {
	SendEmailFunc func(ctx context.Context, req email.SendEmailRequest) error
}

func (m *MockEmailAdapter) SendEmail(ctx context.Context, req email.SendEmailRequest)  error {
	return m.SendEmailFunc(ctx, req)
}
