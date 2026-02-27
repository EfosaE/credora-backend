package infrastructure

import (
	"context"
	"time"

	"github.com/EfosaE/credora-backend/domain/auth"
	"github.com/EfosaE/credora-backend/internal/db/sqlc"
	"github.com/EfosaE/credora-backend/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SqlcPasswordResetRepository struct {
	db *pgxpool.Pool
}

func NewSqlcPasswordResetRepository(db *pgxpool.Pool) *SqlcPasswordResetRepository {
	return &SqlcPasswordResetRepository{
		db: db,
	}
}

func (r *SqlcPasswordResetRepository) queries(ctx context.Context) *sqlc.Queries {
	if tx, ok := GetTx(ctx); ok {
		return sqlc.New(tx)
	}
	return sqlc.New(r.db)
}

func (r *SqlcPasswordResetRepository) Create(
	ctx context.Context,
	userID uuid.UUID,
	tokenHash string,
	expiresAt time.Time,
) (*auth.PasswordReset, error) {

	sqlcReset, err := r.queries(ctx).CreatePasswordReset(ctx, sqlc.CreatePasswordResetParams{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: utils.TimeToPgTimestamp(expiresAt),
		UsedAt:    utils.NullTimeToPgTimestamp(nil),
	})
	if err != nil {
		return nil, err
	}

	return toDomainPasswordReset(sqlcReset), nil
}

func (r *SqlcPasswordResetRepository) GetActiveToken(
	ctx context.Context,
	userID uuid.UUID,
) (*auth.PasswordReset, error) {

	sqlcReset, err := r.queries(ctx).GetActivePasswordReset(ctx, userID)
	if err != nil {
		return nil, err
	}

	return toDomainPasswordReset(sqlcReset), nil
}

func (r *SqlcPasswordResetRepository) MarkUsed(
	ctx context.Context,
	resetID int64,
	usedAt time.Time,
) error {

	err := r.queries(ctx).UpdatePasswordResetUsedAt(ctx, sqlc.UpdatePasswordResetUsedAtParams{
		ID:     resetID,
		UsedAt: utils.TimeToPgTimestamp(usedAt),
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *SqlcPasswordResetRepository) Delete(
	ctx context.Context,
	resetID int64,
) error {
	return r.queries(ctx).DeletePasswordReset(ctx, resetID)
}

func toDomainPasswordReset(sqlcReset sqlc.PasswordReset) *auth.PasswordReset {
	return &auth.PasswordReset{
		ID:        sqlcReset.ID,
		UserID:    sqlcReset.UserID,
		TokenHash: sqlcReset.TokenHash,
		ExpiresAt: utils.PgTimestampToTime(sqlcReset.ExpiresAt),
		UsedAt:    utils.PgTimestampToNullTime(sqlcReset.UsedAt),
		CreatedAt: sqlcReset.CreatedAt.Time,
	}
}


// To ensure this repo satisfies the interface defined in the domain.
var _ auth.PasswordResetRepository = (*SqlcPasswordResetRepository)(nil)