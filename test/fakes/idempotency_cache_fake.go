package fakes

import (
	"context"
	"sync"
	"time"
)

type MockIdempotencyCache struct {
	ttl   time.Duration
	store sync.Map
}

func NewMockIdempotencyCache(ttl time.Duration) *MockIdempotencyCache {
	return &MockIdempotencyCache{
		ttl: ttl,
	}
}

func (m *MockIdempotencyCache) TryAcquire(ctx context.Context, key string) (bool, error) {
	_, exists := m.store.Load(key)
	if exists {
		return false, nil
	}

	// simulate SETNX
	m.store.Store(key, "processing")
	return true, nil
}

func (m *MockIdempotencyCache) MarkDone(ctx context.Context, key string, reference string) error {
	data := map[string]any{
		"processed": true,
		"reference": reference,
		"timestamp": time.Now().Unix(),
	}

	m.store.Store(key, data)
	return nil
}

func (m *MockIdempotencyCache) Get(ctx context.Context, key string) (map[string]any, error) {
	val, ok := m.store.Load(key)
	if !ok {
		return nil, nil
	}

	if data, ok := val.(map[string]any); ok {
		return data, nil
	}

	// if still “processing”, return nil like Redis does before MarkDone
	return nil, nil
}

func (m *MockIdempotencyCache) Delete(ctx context.Context, key string) error {
	m.store.Delete(key)
	return nil
}
