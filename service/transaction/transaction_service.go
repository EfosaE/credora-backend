package transactionsvc

import (
	"context"

	"github.com/EfosaE/credora-backend/domain/logger"
	"github.com/EfosaE/credora-backend/domain/transaction"
	"github.com/google/uuid"
)

type TransactionService struct {
	trxRepo transaction.TransactionRepository
	logger  *logger.Logger
	// eventBus eventbus.EventBus
}

func NewTransactionService(trxRepo transaction.TransactionRepository, logger *logger.Logger) *TransactionService {
	return &TransactionService{
		trxRepo: trxRepo,
		logger:  logger,
	}
}

func (t *TransactionService) RecordTransaction(ctx context.Context, req *transaction.NewTransactionInput) (*transaction.Transaction, error) {
	trx, err := t.trxRepo.RecordTransaction(ctx, req)
	if err != nil {
		return nil, err
	}
	return trx, nil
}

func (t *TransactionService) GetUserTransactions(ctx context.Context, userID uuid.UUID, cursor *transaction.Cursor, limit int32) (*[]transaction.Transaction, *transaction.Cursor, error) {
	trx, nextCursor, err := t.trxRepo.GetUserTransactions(ctx, userID, cursor, limit)
	if err != nil {
		return nil, nil, err
	}
	return trx, nextCursor, nil
}
