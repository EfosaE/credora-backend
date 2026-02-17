package infrastructure

import (
	"context"

	"github.com/EfosaE/credora-backend/domain/user"
	"github.com/EfosaE/credora-backend/internal/db/sqlc"
	"github.com/EfosaE/credora-backend/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SqlcUserRepository struct {
	db *pgxpool.Pool
}

func NewSqlcUserRepository(db *pgxpool.Pool) *SqlcUserRepository {
	return &SqlcUserRepository{
		db: db,
	}
}

func (r *SqlcUserRepository) queries(ctx context.Context) *sqlc.Queries {
	if tx, ok := GetTx(ctx); ok {
		return sqlc.New(tx)
	}
	return sqlc.New(r.db)
}

// ------------------------------------
// Methods
// ------------------------------------

func (r *SqlcUserRepository) Create(ctx context.Context, u *user.CreateUserRequest) (*user.User, error) {
	sqlcUser, err := r.queries(ctx).CreateUser(ctx, sqlc.CreateUserParams{
		FullName:    u.Name,
		Email:       utils.ToPgText(u.Email),
		Password:    u.Password,
		PhoneNumber: u.PhoneNumber,
		Nin:         u.Nin,
	})
	if err != nil {
		return nil, err
	}

	return toDomainUser(sqlcUser), nil
}

func (r *SqlcUserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	sqlcUser, err := r.queries(ctx).GetUserByEmail(ctx, utils.ToPgText(email))
	if err != nil {
		return nil, err
	}
	return toDomainUser(sqlcUser), nil
}

func (r *SqlcUserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, hashedPassword string) error {
	return r.queries(ctx).UpdateUserPassword(ctx, sqlc.UpdateUserPasswordParams{
		ID:       id,
		Password: hashedPassword,
	})
}

// ------------------------------------
// Helpers
// ------------------------------------

func toDomainUser(sqlcUser sqlc.User) *user.User {
	return &user.User{
		ID:        sqlcUser.ID,
		Name:      sqlcUser.FullName,
		Email:     sqlcUser.Email.String,
		CreatedAt: sqlcUser.CreatedAt.Time,
		UpdatedAt: sqlcUser.UpdatedAt.Time,
	}
}
