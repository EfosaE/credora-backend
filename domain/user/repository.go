package user

import (
	"context"

	"github.com/google/uuid"
	// "github.com/google/uuid"
)

// UserRepository defines the methods that the user repository should implement.
type UserRepository interface {
	Create(ctx context.Context, req *CreateUserRequest) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	UpdatePassword(ctx context.Context, id uuid.UUID, hashedPassword string) error
	// GetByNIN(ctx context.Context, nin string) (*User, error)
	// GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	// ListUsers(ctx context.Context, query *ListUsersQuery) ([]*sqlc.User, error)
	// UpdateUser(ctx context.Context, id int64, req *UpdateUserRequest) (*sqlc.User, error)
	// DeleteUser(ctx context.Context, id int64) error
}

type DeviceRepository interface {
	Create(ctx context.Context, userID uuid.UUID, token, platform string) (*DeviceToken, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*DeviceToken, error)
	Update(ctx context.Context, id int64, token, platform string) (*DeviceToken, error)
	Delete(ctx context.Context, id int64) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}