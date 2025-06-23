package user

import (
	"context"

	// "github.com/google/uuid"
)

// UserRepository defines the methods that the user repository should implement.
type UserRepository interface {
	Create(ctx context.Context, req *CreateUserRequest) (*User, error)
	// GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	// ListUsers(ctx context.Context, query *ListUsersQuery) ([]*sqlc.User, error)
	// UpdateUser(ctx context.Context, id int64, req *UpdateUserRequest) (*sqlc.User, error)
	// DeleteUser(ctx context.Context, id int64) error
}
