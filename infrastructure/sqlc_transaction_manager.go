package infrastructure

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TxManager interface {
	WithTx(ctx context.Context, fn func(txCtx context.Context) error) error
}

type txManager struct {
	db *pgxpool.Pool
}

func NewTxManager(db *pgxpool.Pool) TxManager {
	return &txManager{db: db}
}

type txKey struct{}

func (m *txManager) WithTx(
	ctx context.Context,
	fn func(txCtx context.Context) error,
) (err error) {

	// If there's already a transaction in the context, use it (supports nested calls)
	if _, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return fn(ctx)
	}

	tx, err := m.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	txCtx := context.WithValue(ctx, txKey{}, tx)

	if err = fn(txCtx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func GetTx(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	return tx, ok
}
