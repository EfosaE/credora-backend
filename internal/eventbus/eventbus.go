package eventbus

import "context"

type EventBus interface {
	Publish(ctx context.Context, topic string, payload map[string]any) error
	Subscribe(ctx context.Context, topic, group, consumer string, handler func(values map[string]any) error) error
}
