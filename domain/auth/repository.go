package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	// "github.com/google/uuid"
)

type PasswordReset struct {
	ID        int64
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

// PasswordResetRepository defines the methods that the PasswordReset repository should implement.
type PasswordResetRepository interface {
	Create(ctx context.Context, userID uuid.UUID,
		tokenHash string,
		expiresAt time.Time) (*PasswordReset, error)
	MarkUsed(ctx context.Context, resetID int64, usedAt time.Time) error
	GetActiveToken(ctx context.Context, userID uuid.UUID) (*PasswordReset, error)
	Delete(ctx context.Context, resetID int64) error
}
