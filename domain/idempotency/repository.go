package idempotency

import (
	"context"

	"github.com/EfosaE/credora-backend/domain/operation"
	"github.com/EfosaE/credora-backend/domain/transaction"
)

// IdempotencyRepo defines methods that you can use with redi directly
type IdempotencyRepo interface {
	Check(ctx context.Context, key string) (bool, error)
	Insert(ctx context.Context, key string, operationType operation.OperationType, payload any, status transaction.TransactionStatus) error
	Get(ctx context.Context, key string) (*IdempotencyData, error)
	Delete(ctx context.Context, key string) error
	Upsert(ctx context.Context, key string, operationType operation.OperationType, payload any, status transaction.TransactionStatus) error
	UpdateStatus(ctx context.Context, key string, status transaction.TransactionStatus) error
	SaveSuccess(ctx context.Context, key string) error
	SaveFailure(ctx context.Context, key string) error
}
