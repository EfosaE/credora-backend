package transaction

import (
	"context"

	"github.com/google/uuid"
)

// TransactionRepository defines repo methods including WithTx
type TransactionRepository interface {
	RecordTransaction(ctx context.Context, input *NewTransactionInput) (*Transaction, error)
	GetUserTransactions(ctx context.Context, userID uuid.UUID, cursor *Cursor, limit int32) (*[]Transaction, *Cursor, error)
}

// transaction-bound methods
// WithTx(ctx context.Context, fn func(tx TransactionTx) error) error vs WithExistingTx(accTx interface{ Tx() any }) TransactionTx
