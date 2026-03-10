package auth

import (
	"context"
	"github.com/EfosaE/credora-backend/domain/user"

	"github.com/EfosaE/credora-backend/domain/account"
	"github.com/google/uuid"
	// "time"
)

// AuthService handles authentication operations
type AuthService interface {
	Login(ctx context.Context, accountNumber, password string) error
}

type TokenService interface {
	GenerateToken(ctx context.Context, payload TokenPayload) (string, error)
}

type TokenPayload struct {
	Name   string    `json:"name"`
	UserID uuid.UUID `json:"userId"`
}

type LoginResponse struct {
	User     *user.User
	Accounts []account.Account
}
