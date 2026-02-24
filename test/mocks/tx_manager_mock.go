// File: mocks/mock_tx_manager.go
package mocks

import (
	"context"
	"sync"
)

type MockTxManager struct {
	mu sync.Mutex

	// Custom behavior override (optional)
	WithTxFunc func(ctx context.Context, fn func(txCtx context.Context) error) error

	// Tracking
	WithTxCalled bool
	CommitCalled bool
	RollbackCalled bool
}

func (m *MockTxManager) WithTx(
	ctx context.Context,
	fn func(txCtx context.Context) error,
) error {

	m.mu.Lock()
	m.WithTxCalled = true
	m.mu.Unlock()

	// Allow test override
	if m.WithTxFunc != nil {
		return m.WithTxFunc(ctx, fn)
	}

	// Default behavior: execute function directly
	err := fn(ctx)

	m.mu.Lock()
	defer m.mu.Unlock()

	if err != nil {
		m.RollbackCalled = true
		return err
	}

	m.CommitCalled = true
	return nil
}