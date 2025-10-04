package transactionsvc

import (
	"context"

	"github.com/EfosaE/credora-backend/domain/logger"
	"github.com/EfosaE/credora-backend/domain/transaction"
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
