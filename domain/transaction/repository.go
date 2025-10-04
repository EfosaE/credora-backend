package transaction

import (
	"context"
	// "github.com/EfosaE/credora-backend/internal/db/sqlc"
	// "github.com/google/uuid"
)

// TransactionRepository defines the methods that the sqlc transaction repository should implement & transactionsvc can call.
type TransactionRepository interface {
	RecordTransaction(ctx context.Context, req *NewTransactionInput) (*Transaction, error)
}
