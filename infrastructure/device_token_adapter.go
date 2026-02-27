package infrastructure

import (
	"context"

	"github.com/EfosaE/credora-backend/domain/user"
	"github.com/EfosaE/credora-backend/internal/db/sqlc"
	"github.com/EfosaE/credora-backend/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SqlcDeviceTokenRepository struct {
	db *pgxpool.Pool
}

func NewSqlcDeviceTokenRepository(db *pgxpool.Pool) *SqlcDeviceTokenRepository {
	return &SqlcDeviceTokenRepository{
		db: db,
	}
}

func (r *SqlcDeviceTokenRepository) queries(ctx context.Context) *sqlc.Queries {
	if tx, ok := GetTx(ctx); ok {
		return sqlc.New(tx)
	}
	return sqlc.New(r.db)
}

// ------------------------------------
// Methods
// ------------------------------------

func (r *SqlcDeviceTokenRepository) Create(ctx context.Context, userID uuid.UUID, token, platform string) (*user.DeviceToken, error) {
	dt, err := r.queries(ctx).CreateDeviceToken(ctx, sqlc.CreateDeviceTokenParams{
		UserID:   userID,
		Token:    token,
		Platform: utils.ToPgText(platform),
	})
	if err != nil {
		return nil, err
	}
	return toDomainDeviceToken(dt), nil
}

func (r *SqlcDeviceTokenRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*user.DeviceToken, error) {
	rows, err := r.queries(ctx).GetDeviceTokensByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	tokens := make([]*user.DeviceToken, len(rows))
	for i, row := range rows {
		tokens[i] = toDomainDeviceToken(row)
	}
	return tokens, nil
}

func (r *SqlcDeviceTokenRepository) Update(ctx context.Context, id int64, token, platform string) (*user.DeviceToken, error) {
	dt, err := r.queries(ctx).UpdateDeviceToken(ctx, sqlc.UpdateDeviceTokenParams{
		ID:       id,
		Token:    token,
		Platform: utils.ToPgText(platform),
	})
	if err != nil {
		return nil, err
	}
	return toDomainDeviceToken(dt), nil
}

func (r *SqlcDeviceTokenRepository) Delete(ctx context.Context, id int64) error {
	return r.queries(ctx).DeleteDeviceToken(ctx, id)
}

func (r *SqlcDeviceTokenRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.queries(ctx).DeleteDeviceTokensByUserID(ctx, userID)
}

// ------------------------------------
// Helpers
// ------------------------------------

func toDomainDeviceToken(dt sqlc.DeviceToken) *user.DeviceToken {
	return &user.DeviceToken{
		ID:        dt.ID,
		UserID:    dt.UserID,
		Token:     dt.Token,
		Platform:  dt.Platform.String,
		CreatedAt: dt.CreatedAt.Time,
	}
}
