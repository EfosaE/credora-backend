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

func (s *IdempotencyService) Exists(ctx context.Context, key string) (bool, error) {
	return s.idempRepo.Check(ctx, key)
}

func (s *IdempotencyService) AddToIdempotencyTable(ctx context.Context, key string, opType operation.OperationType, data any, status transaction.TransactionStatus) error {
	return s.idempRepo.Insert(ctx, key, opType, data, status)
}

func (s *IdempotencyService) GetRecord(ctx context.Context, key string) (*idempotency.IdempotencyData, error) {
	return s.idempRepo.Get(ctx, key)
}

func (s *IdempotencyService) MarkProcessed(ctx context.Context, key string) error {
	return s.idempRepo.SaveSuccess(ctx, key)
}
