package infrastructure

import (
	"context"

	"github.com/EfosaE/credora-backend/domain/user"
	"github.com/EfosaE/credora-backend/internal/db/sqlc"
	"github.com/EfosaE/credora-backend/internal/utils"
	// "github.com/google/uuid"
)

type SqlcRepository struct {
	q *sqlc.Queries
}

func NewSqlcUserRepository(ctx context.Context, q *sqlc.Queries) *SqlcRepository {
	return &SqlcRepository{
		q: q,
	}
}


// this SqlcRepository implements the UserRepository interface because it has all the methods defined in the interface
func (s *SqlcRepository) Create(ctx context.Context, user *user.CreateUserRequest) (*user.User, error) {
	sqlcUser, err := s.q.CreateUser(ctx, sqlc.CreateUserParams{
		FullName:    user.Name,
		Email:       utils.ToPgText(user.Email),
		Password:    user.Password,
		PhoneNumber: user.PhoneNumber,
		Nin:         user.Nin,
	})
	if err != nil {
		return nil, err
	}

	// Convert sqlc.User to User
	return toDomainUser(sqlcUser), nil
}

// 🔑 KEY: Conversion between sqlc and domain types
// This is where you control what gets exposed vs hidden
func toDomainUser(sqlcUser sqlc.User) *user.User {
	return &user.User{
		ID:        sqlcUser.ID,
		Name:      sqlcUser.FullName,
		Email:     sqlcUser.Email.String,
		CreatedAt: sqlcUser.CreatedAt.Time,
		UpdatedAt: sqlcUser.UpdatedAt.Time,
		// Notice: password_hash, internal_notes, etc. are NOT mapped
		// This prevents accidental exposure of sensitive data
	}
}
