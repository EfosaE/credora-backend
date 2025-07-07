package service

import (
	"context"
	"testing"

	"github.com/EfosaE/credora-backend/domain/email"
	"github.com/EfosaE/credora-backend/domain/user"
	"github.com/EfosaE/credora-backend/test/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSendWelcomeEmail_Success(t *testing.T) {
	mockEventBus := &mocks.MockEventBus{}
	mockEmailAdapter := &mocks.MockEmailAdapter{
		SendEmailFunc: func(ctx context.Context, req email.SendEmailRequest) error {
			return nil
		},
	}

	svc := NewEmailService(mockEmailAdapter, mockEventBus)

	user := user.User{ID: uuid.New(), Name: "Efosa", Email: "efosa@example.com"}

	err := svc.SendWelcomeEmail(context.Background(), user)
	assert.NoError(t, err)
	// sender.AssertExpectations(t)
}
