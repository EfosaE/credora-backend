package event

import (
	"context"
)

type EventBus interface {
	Publish(ctx context.Context, stream, eventType string, payload map[string]any) error
	Subscribe(ctx context.Context, stream, group, consumer string, handler func(ctx context.Context, msg EventMessage) error) error
}
