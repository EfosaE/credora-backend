package transaction

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// TransactionTx defines transactional operations
type TransactionTx interface {
	RecordTransaction(ctx context.Context, input *NewTransactionInput) (*Transaction, error)
}

// TransactionRepository defines repo methods including WithTx
type TransactionRepository interface {
	RecordTransaction(ctx context.Context, input *NewTransactionInput) (*Transaction, error)

	// Bind a pgx.Tx to this repo
    WithTx(tx pgx.Tx) TransactionTx
}

// transaction-bound methods
// WithTx(ctx context.Context, fn func(tx TransactionTx) error) error vs WithExistingTx(accTx interface{ Tx() any }) TransactionTx