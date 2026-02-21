package mocks

import (
	"context"
	"sync"

	"github.com/EfosaE/credora-backend/domain/event"
)

type MockEventBus struct {
	mu             sync.Mutex
	Published      []PublishedEvent
	Subscribed     []Subscription
	HandlerInvoked bool
}

type PublishedEvent struct {
	Topic   string
	Payload map[string]any
}

type Subscription struct {
	Topic    string
	Group    string
	Consumer string
}

func (m *MockEventBus) Publish(ctx context.Context, topic string, eventType string, payload map[string]any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Published = append(m.Published, PublishedEvent{
		Topic:   topic,
		Payload: payload,
	})

	return nil
}

func (m *MockEventBus) Subscribe(ctx context.Context, topic, group, consumer string, handler func(ctx context.Context, msg event.EventMessage) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Subscribed = append(m.Subscribed, Subscription{
		Topic:    topic,
		Group:    group,
		Consumer: consumer,
	})

	// You can simulate handler invocation in tests if needed
	m.HandlerInvoked = true

	return nil
}
