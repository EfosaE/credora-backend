package idempotencysvc

import (
	"context"

	"github.com/EfosaE/credora-backend/domain/idempotency"
	"github.com/EfosaE/credora-backend/domain/operation"
	"github.com/EfosaE/credora-backend/domain/transaction"
)

type IdempotencyService struct {
	idempRepo idempotency.IdempotencyRepo
}


func NewIdempotencyService(idempRepo idempotency.IdempotencyRepo) *IdempotencyService {
	return &IdempotencyService{
		idempRepo: idempRepo,
	}
}



func (s *IdempotencyService) AddToIdempotencyTable(ctx context.Context, key string, opType operation.OperationType, data any, status transaction.TransactionStatus) error {
	return s.idempRepo.Insert(ctx, key, opType, data, status)
}


func (s *IdempotencyService) GetRecord(ctx context.Context, key string) (*idempotency.IdempotencyData, error) {
	return s.idempRepo.Get(ctx, key)
}