package fakes

import "context"

type NoOpTxManager struct{}

func NewNoOpTxManager() *NoOpTxManager {
    return &NoOpTxManager{}
}

func (m *NoOpTxManager) WithTx(
    ctx context.Context,
    fn func(txCtx context.Context) error,
) error {
    return fn(ctx)
}