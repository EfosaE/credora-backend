package idempotency

import (
	"context"

	"github.com/EfosaE/credora-backend/domain/operation"
	"github.com/EfosaE/credora-backend/domain/transaction"
	"github.com/jackc/pgx/v5"
)

// IdempotencyTx defines transactional operations
type IdempotencyTx interface {
	Check(ctx context.Context, key string) (bool, error)
	Insert(ctx context.Context, key string, operationType operation.OperationType, payload any, status transaction.TransactionStatus) error
	Get(ctx context.Context, key string) (*IdempotencyData, error)
	Delete(ctx context.Context, key string) error
	Upsert(ctx context.Context, key string, operationType operation.OperationType, payload any, status transaction.TransactionStatus) error
	UpdateStatus(ctx context.Context, key string, status transaction.TransactionStatus) error
	SaveSuccess(ctx context.Context, key string, payload any) error
	SaveFailure(ctx context.Context, key string, payload any) error
}

// IdempotencyTable defines non-transactional operations
type IdempotencyTable interface {
	Check(ctx context.Context, key string) (bool, error)
	Insert(ctx context.Context, key string, operationType operation.OperationType, payload any, status transaction.TransactionStatus) error
	Get(ctx context.Context, key string) (*IdempotencyData, error)
	Delete(ctx context.Context, key string) error
	Upsert(ctx context.Context, key string, operationType operation.OperationType, payload any, status transaction.TransactionStatus) error
	UpdateStatus(ctx context.Context, key string, status transaction.TransactionStatus) error
	SaveSuccess(ctx context.Context, key string, payload any) error
	SaveFailure(ctx context.Context, key string, payload any) error

	// Bind a pgx.Tx to this repo
	WithTx(tx pgx.Tx) IdempotencyTx
}
