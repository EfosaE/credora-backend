package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type IdempotencyCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewIdempotencyCache(client *redis.Client, ttl time.Duration) *IdempotencyCache {
	return &IdempotencyCache{
		client: client,
		ttl:    ttl,
	}
}

// TryAcquire attempts to lock the idempotency key using SETNX.
// Returns true if this is a new operation and should proceed.
// Returns false if the key already exists (duplicate request).
func (c *IdempotencyCache) TryAcquire(ctx context.Context, key string) (bool, error) {
	redisKey := "idem:" + key

	ok, err := c.client.SetNX(ctx, redisKey, "processing", c.ttl).Result()
	if err != nil {
		return false, fmt.Errorf("redis SETNX failed: %w", err)
	}

	return ok, nil
}

// MarkDone stores the final result metadata of a processed request.
// This is optional but useful for debugging or returning saved responses.
func (c *IdempotencyCache) MarkDone(
	ctx context.Context,
	key string,
	reference string,
) error {
	redisKey := "idem:" + key

	data := map[string]any{
		"processed": true,
		"reference": reference,
		"timestamp": time.Now().Unix(),
	}

	b, _ := json.Marshal(data)

	err := c.client.Set(ctx, redisKey, b, c.ttl).Err()
	if err != nil {
		return fmt.Errorf("redis SET failed: %w", err)
	}

	return nil
}

// Get returns the stored metadata for a completed idempotent request.
// Useful if you want to return cached responses (like Stripe does).
func (c *IdempotencyCache) Get(ctx context.Context, key string) (map[string]any, error) {
	redisKey := "idem:" + key

	raw, err := c.client.Get(ctx, redisKey).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis GET failed: %w", err)
	}

	var result map[string]any
	_ = json.Unmarshal(raw, &result)

	return result, nil
}

// Delete removes the key (optional, depending on design).
func (c *IdempotencyCache) Delete(ctx context.Context, key string) error {
	redisKey := "idem:" + key
	return c.client.Del(ctx, redisKey).Err()
}
