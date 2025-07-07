package auth

import (
	"context"

	"github.com/google/uuid"
	// "time"
)

// AuthService handles authentication operations
type AuthService interface {
	Login(ctx context.Context, account_number, password string) error
	// Register(ctx context.Context, req RegisterRequest) (*AuthResult, error)
	// RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error)
	// Logout(ctx context.Context, token string) error
	// ValidateToken(ctx context.Context, token string) (*TokenPayload, error)
}

type TokenService interface {
	GenerateToken(ctx context.Context, payload TokenPayload) (string, error)
	// ParseToken(ctx context.Context, tokenString string) (TokenPayload, error)
	// GenerateRefreshToken(ctx context.Context, userID uuid.UUID) (string, error)
}

type TokenPayload struct {
	AccountNumber string
	Name          string
	UserID        uuid.UUID
}
