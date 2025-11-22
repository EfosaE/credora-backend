package mocks

import (
	"github.com/EfosaE/credora-backend/domain/operation"
	"github.com/EfosaE/credora-backend/internal/queues"
)

type MockQueueClient struct {
	EnqueueWelcomeEmailFunc       func(payload queues.WelcomeEmailPayload) error
	EnqueueAccountNumberEmailFunc func(payload queues.AccountNumberEmailPayload) error
	EnqueueInternalTransferFunc   func(payload operation.InternalTransferReq) error
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

func (m *MockQueueClient) EnqueueInternalTransfer(payload *operation.InternalTransferReq) error {
	if m.EnqueueInternalTransferFunc != nil {
		return m.EnqueueInternalTransferFunc(*payload)
	}
	return nil
}
