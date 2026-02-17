package transaction

import (
	"context"
)

// TransactionRepository defines repo methods including WithTx
type TransactionRepository interface {
	RecordTransaction(ctx context.Context, input *NewTransactionInput) (*Transaction, error)
}

// transaction-bound methods
// WithTx(ctx context.Context, fn func(tx TransactionTx) error) error vs WithExistingTx(accTx interface{ Tx() any }) TransactionTx
