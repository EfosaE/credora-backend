package service_test

import (
	"context"
	"testing"

	"github.com/EfosaE/credora-backend/domain/email"
	"github.com/EfosaE/credora-backend/domain/user"
	"github.com/EfosaE/credora-backend/service"
	"github.com/EfosaE/credora-backend/test/fakes"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSendWelcomeEmail_Success(t *testing.T) {
	mockEventBus := &fakes.MockEventBus{}
	mockEmailAdapter := &fakes.MockEmailAdapter{
		SendEmailFunc: func(ctx context.Context, req email.SendEmailRequest) error {
			return nil
		},
	}

	svc := service.NewEmailService(mockEmailAdapter, mockEventBus)

	user := user.User{ID: uuid.New(), Name: "Efosa", Email: "efosa@example.com"}

	err := svc.SendWelcomeEmail(context.Background(), user)
	assert.NoError(t, err)
	// sender.AssertExpectations(t)
}
