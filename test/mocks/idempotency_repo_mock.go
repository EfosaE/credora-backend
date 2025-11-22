package mocks

import (
	"context"
	"encoding/json"
	"errors"
	"sync"

	"github.com/EfosaE/credora-backend/domain/idempotency"
	"github.com/EfosaE/credora-backend/domain/operation"
	"github.com/EfosaE/credora-backend/domain/transaction"
)

// MockIdempotencyRepo implements IdempotencyRepo and IdempotencyTx
type MockIdempotencyRepo struct {
	mu   sync.Mutex
	data map[string]idempotency.IdempotencyData
}

func NewMockIdempotencyRepo() *MockIdempotencyRepo {
	return &MockIdempotencyRepo{
		data: make(map[string]idempotency.IdempotencyData),
	}
}

// WithTx returns the same mock for transaction-bound repo
func (m *MockIdempotencyRepo) WithTx(_ any) idempotency.IdempotencyTx {
	return m
}

// Check if key exists
func (m *MockIdempotencyRepo) Check(ctx context.Context, key string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.data[key]
	return ok, nil
}

// Get returns the stored record
func (m *MockIdempotencyRepo) Get(ctx context.Context, key string) (*idempotency.IdempotencyData, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.data[key]
	if !ok {
		return nil, nil
	}
	return &v, nil
}

// Insert a new key
func (m *MockIdempotencyRepo) Insert(ctx context.Context, key string, opType operation.OperationType, payload any, status transaction.TransactionStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	b, _ := json.Marshal(payload)
	m.data[key] = idempotency.IdempotencyData{
		IdemKey:       key,
		OperationType: string(opType),
		Payload:       b,
		Status:        status,
	}
	return nil
}

// Upsert updates or inserts
func (m *MockIdempotencyRepo) Upsert(ctx context.Context, key string, opType operation.OperationType, payload any, status transaction.TransactionStatus) error {
	return m.Insert(ctx, key, opType, payload, status)
}

// SaveSuccess marks a key as SUCCESS
func (m *MockIdempotencyRepo) SaveSuccess(ctx context.Context, key string, payload any) error {
	return m.Upsert(ctx, key, operation.OperationTypeInternalTransfer, payload, transaction.StatusSuccess)
}

// SaveFailure marks a key as FAILED
func (m *MockIdempotencyRepo) SaveFailure(ctx context.Context, key string, payload any) error {
	return m.Upsert(ctx, key, operation.OperationTypeInternalTransfer, payload, transaction.StatusFailed)
}

// Delete removes a key from the store
func (m *MockIdempotencyRepo) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	return nil
}

// UpdateStatus updates only the status of an existing record
func (m *MockIdempotencyRepo) UpdateStatus(ctx context.Context, key string, status transaction.TransactionStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	rec, ok := m.data[key]
	if !ok {
		return errors.New("idempotency key not found")
	}
	rec.Status = status
	m.data[key] = rec
	return nil
}
