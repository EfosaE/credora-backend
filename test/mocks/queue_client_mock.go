package mocks

import (
	"github.com/EfosaE/credora-backend/internal/queues"
)

type MockQueueClient struct {
	EnqueueWelcomeEmailFunc func(payload queues.WelcomeEmailPayload) error
	EnqueueAccountNumberEmailFunc func(payload queues.AccountNumberEmailPayload) error
}

func (m *MockQueueClient) EnqueueWelcomeEmail(payload queues.WelcomeEmailPayload) error {
	if m.EnqueueWelcomeEmailFunc != nil {
		return m.EnqueueWelcomeEmailFunc(payload)
	}
	return nil
}

func (m *MockQueueClient) EnqueueAccountNumberEmail(payload queues.AccountNumberEmailPayload) error {
	if m.EnqueueAccountNumberEmailFunc != nil {
		return m.EnqueueAccountNumberEmailFunc(payload)
	}
	return nil
}