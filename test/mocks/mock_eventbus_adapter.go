package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/EfosaE/credora-backend/domain/event"
)

type MockEventBus struct {
	mock.Mock
}

func (m *MockEventBus) Publish(
	ctx context.Context,
	stream string,
	eventType string,
	payload map[string]any,
) error {

	args := m.Called(ctx, stream, eventType, payload)
	return args.Error(0)
}

func (m *MockEventBus) Subscribe(
	ctx context.Context,
	stream string,
	group string,
	consumer string,
	handler func(ctx context.Context, msg event.EventMessage) error,
) error {

	args := m.Called(ctx, stream, group, consumer, handler)
	return args.Error(0)
}

// Compile-time guarantee
var _ event.EventBus = (*MockEventBus)(nil)