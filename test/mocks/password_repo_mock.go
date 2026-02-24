package mocks

import (
	"context"
	"time"

	"github.com/EfosaE/credora-backend/domain/auth"
	"github.com/google/uuid"

	"crypto/sha256"
	"encoding/hex"
)

func HashPasswordForTest(password string) string {
	hash := sha256.Sum256([]byte("test-salt:" + password))
	return hex.EncodeToString(hash[:])
}

type MockPasswordResetRepo struct {
	CreateFunc         func(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (*auth.PasswordReset, error)
	MarkUsedFunc       func(ctx context.Context, resetID int64, usedAt time.Time) error
	GetActiveTokenFunc func(ctx context.Context, userID uuid.UUID) (*auth.PasswordReset, error)
	DeleteFunc         func(ctx context.Context, resetID int64) error
}

func (m *MockPasswordResetRepo) Create(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (*auth.PasswordReset, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, userID, tokenHash, expiresAt)
	}
	return nil, nil
}

func (m *MockPasswordResetRepo) MarkUsed(ctx context.Context, resetID int64, usedAt time.Time) error {
	if m.MarkUsedFunc != nil {
		return m.MarkUsedFunc(ctx, resetID, usedAt)
	}
	return nil
}

func (m *MockPasswordResetRepo) GetActiveToken(ctx context.Context, userID uuid.UUID) (*auth.PasswordReset, error) {
	if m.GetActiveTokenFunc != nil {
		return m.GetActiveTokenFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockPasswordResetRepo) Delete(ctx context.Context, resetID int64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, resetID)
	}
	return nil
}
