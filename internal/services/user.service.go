package services

import (
	"context"

	"github.com/EfosaE/credora-backend/internal/db/sqlc"
	"github.com/EfosaE/credora-backend/internal/domain"
	"github.com/google/uuid"
)

// UserService defines the methods that the user service should implement.
type UserService interface {
	CreateUser(ctx context.Context, req *domain.CreateUserRequest) (*domain.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	// ListUsers(ctx context.Context, query *domain.ListUsersQuery) ([]*sqlc.User, error)
	// UpdateUser(ctx context.Context, id int64, req *domain.UpdateUserRequest) (*sqlc.User, error)
	// DeleteUser(ctx context.Context, id int64) error
}

type SqlcUserService struct {
	q   *sqlc.Queries
	ctx context.Context
}

func NewSqlcUserService(ctx context.Context, q *sqlc.Queries) *SqlcUserService {
	return &SqlcUserService{
		q:   q,
		ctx: ctx,
	}
}

// this SqlcUserService implements the UserService interface because it has all the methods defined in the interface
func (s *SqlcUserService) GetUserByID(id uuid.UUID) (*domain.User, error) {
	sqlcUser, err := s.q.GetUser(s.ctx, id)
	if err != nil {
		return nil, err
	}

	// Convert sqlc.User to domain.User
	return toDomainUser(sqlcUser), nil
}

// 🔑 KEY: Conversion between sqlc and domain types
// This is where you control what gets exposed vs hidden
func toDomainUser(sqlcUser sqlc.User) *domain.User {
	return &domain.User{
		ID:        sqlcUser.ID,
		Name:      sqlcUser.FullName,
		Email:     sqlcUser.Email.String,
		CreatedAt: sqlcUser.CreatedAt.Time,
		UpdatedAt: sqlcUser.UpdatedAt.Time,
		// Notice: password_hash, internal_notes, etc. are NOT mapped
		// This prevents accidental exposure of sensitive data
	}
}


