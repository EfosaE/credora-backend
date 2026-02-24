package mocks

import (
	"context"
	"sync"

	"github.com/EfosaE/credora-backend/domain/event"
)

type MockEventBus struct {
	mu sync.Mutex

	// Optional function overrides (for table-driven tests)
	PublishFunc   func(ctx context.Context, topic string, eventType string, payload map[string]any) error
	SubscribeFunc func(ctx context.Context, topic, group, consumer string,
		handler func(ctx context.Context, msg event.EventMessage) error) error

	// Tracking
	Published  []PublishedEvent
	Subscribed []Subscription

	// Captured handlers by topic
	Handlers map[string]func(ctx context.Context, msg event.EventMessage) error
}

type PublishedEvent struct {
	Topic     string
	EventType string
	Payload   map[string]any
}

type Subscription struct {
	Topic    string
	Group    string
	Consumer string
}

func (m *MockEventBus) Publish(
	ctx context.Context,
	topic string,
	eventType string,
	payload map[string]any,
) error {
	if m.PublishFunc != nil {
		return m.PublishFunc(ctx, topic, eventType, payload)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.Published = append(m.Published, PublishedEvent{
		Topic:     topic,
		EventType: eventType,
		Payload:   payload,
	})

	return nil
}

func (m *MockEventBus) Subscribe(
	ctx context.Context,
	topic, group, consumer string,
	handler func(ctx context.Context, msg event.EventMessage) error,
) error {

	// If override provided, use it
	if m.SubscribeFunc != nil {
		return m.SubscribeFunc(ctx, topic, group, consumer, handler)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.Handlers == nil {
		m.Handlers = make(map[string]func(context.Context, event.EventMessage) error)
	}

	m.Subscribed = append(m.Subscribed, Subscription{
		Topic:    topic,
		Group:    group,
		Consumer: consumer,
	})

	m.Handlers[topic] = handler

	return nil
}
