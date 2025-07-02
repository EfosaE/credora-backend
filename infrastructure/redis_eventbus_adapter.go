package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type StreamEventBus struct {
	client *redis.Client
}

func NewStreamEventBus(client *redis.Client) *StreamEventBus {
	return &StreamEventBus{client: client}
}

// Publish an event to a stream using XADD
func (s *StreamEventBus) Publish(ctx context.Context, stream string, payload map[string]any) error {
	_, err := s.client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: payload,
	}).Result()
	return err
}

// Subscribe to a stream using Consumer Groups
func (s *StreamEventBus) Subscribe(
	ctx context.Context,
	stream string,
	group string,
	consumer string,
	handler func(values map[string]any) error,
) error {
	// Create the consumer group if it doesn't exist
	err := s.client.XGroupCreateMkStream(ctx, stream, group, "0").Err()
	if err != nil && !isBusyGroupErr(err) {
		return fmt.Errorf("failed to create consumer group: %w", err)
	}

	go func() {
		for {
			// Read messages from the group
			res, err := s.client.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    group,
				Consumer: consumer,
				Streams:  []string{stream, ">"},
				Block:    5 * time.Second,
				Count:    10,
			}).Result()

			if err != nil && err != redis.Nil {
				fmt.Println("❌ Stream read error:", err)
				continue
			}

			for _, stream := range res {
				for _, msg := range stream.Messages {
					handler(msg.Values)

					// Acknowledge message
					_ = s.client.XAck(ctx, stream.Stream, group, msg.ID).Err()
				}
			}
		}
	}()

	return nil
}

func isBusyGroupErr(err error) bool {
	return err != nil && (err.Error() == "BUSYGROUP Consumer Group name already exists")
}
