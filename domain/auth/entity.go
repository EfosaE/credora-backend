package auth

import (
	"context"
	"github.com/EfosaE/credora-backend/domain/user"
	"github.com/google/uuid"
	// "time"
)

// AuthService handles authentication operations
type AuthService interface {
	Login(ctx context.Context, identifier, password string) error
}

type TokenService interface {
	GenerateToken(ctx context.Context, payload TokenPayload) (string, error)
}

type TokenPayload struct {
	Name   string    `json:"name"`
	UserID uuid.UUID `json:"userId"`
}

type LoginUserRequest struct {
	Identifier string `json:"identifier" validate:"required"`
	Password   string `json:"password" validate:"required,min=8,max=100"`
}

type LoginResponse struct {
	AccessToken string    `json:"accessToken"`
	User        user.User `json:"user"`
}

type ResetPasswordRequest struct {
	Email string `json:"email" validate:"required,email,max=255"`
}

type ValidatePasswordRequest struct {
	Email              string `json:"email" validate:"required,email,max=255"`
	PasswordResetToken string `json:"passwordResetToken" validate:"required"`
	NewPassword        string `json:"newPassword" validate:"required,min=8,max=100"`
}
