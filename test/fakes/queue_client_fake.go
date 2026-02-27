package fakes

import (
	"github.com/EfosaE/credora-backend/domain/webhook"
	"github.com/EfosaE/credora-backend/internal/queues"
)

type MockQueueClient struct {
	EnqueueWelcomeEmailFunc           func(payload queues.WelcomeEmailPayload) error
	EnqueueAccountNumberEmailFunc     func(payload queues.AccountNumberEmailPayload) error
	EnqueueInternalTransferFunc       func(payload queues.InternalTransferTaskPayload) error
	EnqueueWebhookInboundTransferFunc func(payload webhook.InboundTransferPayload) error
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

func (m *MockQueueClient) EnqueueInternalTransfer(payload queues.InternalTransferTaskPayload) error {
	if m.EnqueueInternalTransferFunc != nil {
		return m.EnqueueInternalTransferFunc(payload)
	}
	return nil
}

func (m *MockQueueClient) EnqueueWebhookInboundTransfer(payload *webhook.InboundTransferPayload) error {
	if m.EnqueueWebhookInboundTransferFunc != nil {
		return m.EnqueueWebhookInboundTransferFunc(*payload)
	}
	return nil
}
