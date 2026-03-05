package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/EfosaE/credora-backend/domain/event"
	"github.com/redis/go-redis/v9"
)

type StreamEventBus struct {
	client *redis.Client
}

func NewStreamEventBus(client *redis.Client) *StreamEventBus {
	return &StreamEventBus{client: client}
}

// Publish an event to a stream using XADD
func (s *StreamEventBus) Publish(
	ctx context.Context,
	stream string,
	eventType string,
	payload map[string]any,
) error {

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to serialize event payload: %w", err)
	}

	event := map[string]any{
		"event": eventType,
		"data":  string(data),
	}

	_, err = s.client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: event,
	}).Result()

	if err != nil {
		return fmt.Errorf("failed to publish event to stream %s: %w", stream, err)
	}

	return nil
}

// Subscribe to a stream using Consumer Groups
func (s *StreamEventBus) Subscribe(
	ctx context.Context,
	stream string,
	group string,
	consumer string,
	handler func(ctx context.Context, msg event.EventMessage) error,
) error {

	// Create consumer group if not exists
	err := s.client.XGroupCreateMkStream(ctx, stream, group, "0").Err()
	if err != nil && !isBusyGroupErr(err) {
		return fmt.Errorf("failed to create consumer group: %w", err)
	}

	go func() {

		for {
			res, err := s.client.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    group,
				Consumer: consumer,
				Streams:  []string{stream, ">"},
				Block:    5 * time.Second,
				Count:    10,
			}).Result()

			if err != nil && err != redis.Nil {
				continue
			}

			for _, strm := range res {
				for _, msg := range strm.Messages {

					// Convert Redis payload → EventMessage
					eventMsg := event.EventMessage{
						EventType: "",
						Data:      "",
					}

					if v, ok := msg.Values["event"].(string); ok {
						eventMsg.EventType = v
					}

					if v, ok := msg.Values["data"].(string); ok {
						eventMsg.Data = v
					}

					// Execute handler
					err := handler(ctx, eventMsg)

					// ACK ONLY if handler succeeds
					if err == nil {
						_ = s.client.XAck(ctx, strm.Stream, group, msg.ID).Err()
					}
				}
			}
		}
	}()

	return nil
}

func isBusyGroupErr(err error) bool {
	return err != nil && (err.Error() == "BUSYGROUP Consumer Group name already exists")
}

// To ensure this repo satisfies the interface defined in the domain.
var _ event.EventBus = (*StreamEventBus)(nil)
