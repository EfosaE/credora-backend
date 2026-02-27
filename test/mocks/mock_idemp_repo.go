package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/EfosaE/credora-backend/domain/idempotency"
	"github.com/EfosaE/credora-backend/domain/operation"
	"github.com/EfosaE/credora-backend/domain/transaction"
)

type MockIdempotencyRepository struct {
	mock.Mock
}

func (m *MockIdempotencyRepository) Check(
	ctx context.Context,
	key string,
) (bool, error) {

	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

func (m *MockIdempotencyRepository) Insert(
	ctx context.Context,
	key string,
	operationType operation.OperationType,
	payload any,
	status transaction.TransactionStatus,
) error {

	args := m.Called(ctx, key, operationType, payload, status)
	return args.Error(0)
}

func (m *MockIdempotencyRepository) Get(
	ctx context.Context,
	key string,
) (*idempotency.IdempotencyData, error) {

	args := m.Called(ctx, key)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*idempotency.IdempotencyData), args.Error(1)
}

func (m *MockIdempotencyRepository) Delete(
	ctx context.Context,
	key string,
) error {

	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockIdempotencyRepository) Upsert(
	ctx context.Context,
	key string,
	operationType operation.OperationType,
	payload any, // must match interface exactly
	status transaction.TransactionStatus,
) error {

	args := m.Called(ctx, key, operationType, payload, status)
	return args.Error(0)
}

func (m *MockIdempotencyRepository) UpdateStatus(
	ctx context.Context,
	key string,
	status transaction.TransactionStatus,
) error {

	args := m.Called(ctx, key, status)
	return args.Error(0)
}

func (m *MockIdempotencyRepository) SaveSuccess(
	ctx context.Context,
	key string,
) error {

	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockIdempotencyRepository) SaveFailure(
	ctx context.Context,
	key string,
) error {

	args := m.Called(ctx, key)
	return args.Error(0)
}

var _ idempotency.IdempotencyRepo = (*MockIdempotencyRepository)(nil)
