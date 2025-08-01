package transaction

import (
	"context"
	// "github.com/EfosaE/credora-backend/internal/db/sqlc"
	// "github.com/google/uuid"
)

// AccountRepository defines the methods that the sqlc account repository should implement.
type TransactionRepository interface {
	AddMoney(ctx context.Context, req *AddMoneyToWalletRequest) error
	
	// Init(ctx context.Context, req *CreateAccountRequest) (*Account, error)
	// Debit()
	// GetUserByAccountNumber(ctx context.Context, accountNumber string) (*sqlc.GetUserByAccountNumberRow, error)
}
