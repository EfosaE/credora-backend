package infrastructure

import (
	"context"
	"encoding/json"
	"log"

	"github.com/EfosaE/credora-backend/domain/idempotency"
	"github.com/EfosaE/credora-backend/domain/operation"
	"github.com/EfosaE/credora-backend/domain/transaction"
	"github.com/EfosaE/credora-backend/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SqlcIdempotencyRepository implements both IdempotencyRepo and IdempotencyTx
type SqlcIdempotencyRepository struct {
	db *pgxpool.Pool
}

func NewSqlcIdempotencyRepository(db *pgxpool.Pool) *SqlcIdempotencyRepository {
	return &SqlcIdempotencyRepository{
		db: db,
	}
}

func (r *SqlcIdempotencyRepository) queries(ctx context.Context) *sqlc.Queries {
	// create a repo bound to a transaction if it exists in the context, otherwise use the main DB connection
	if tx, ok := GetTx(ctx); ok {
		return sqlc.New(tx)
	}
	return sqlc.New(r.db)
}

// ------------------------------------
// Normal methods / Tx methods
// ------------------------------------

// Check if an idempotency key exists
func (i *SqlcIdempotencyRepository) Check(ctx context.Context, key string) (bool, error) {
	return i.queries(ctx).CheckIdempotency(ctx, key)
}

// Insert a new idempotency key with a status
func (i *SqlcIdempotencyRepository) Insert(ctx context.Context, key string, opType operation.OperationType, payload any, status transaction.TransactionStatus) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return i.queries(ctx).InsertIdempotencyKey(ctx, sqlc.InsertIdempotencyKeyParams{
		IdemKey:       key,
		OperationType: string(opType),
		Payload:       b,
		Status:        string(status),
	})
}

// Get an idempotency record
func (i *SqlcIdempotencyRepository) Get(ctx context.Context, key string) (*idempotency.IdempotencyData, error) {
	data, err := i.queries(ctx).GetIdempotencyKey(ctx, key)
	if err != nil {
		return nil, err
	}

	return &idempotency.IdempotencyData{
		IdemKey:       data.IdemKey,
		OperationType: data.OperationType,
		Payload:       data.Payload,
		Status:        transaction.TransactionStatus(data.Status),
	}, nil
}

// Delete an idempotency record
func (i *SqlcIdempotencyRepository) Delete(ctx context.Context, key string) error {
	return i.queries(ctx).DeleteIdempotencyKey(ctx, key)
}

// Upsert an idempotency key with status
func (i *SqlcIdempotencyRepository) Upsert(ctx context.Context, key string, operationType operation.OperationType, payload any, status transaction.TransactionStatus) error {
	log.Println("Upserting idempotency key:", key, "with status:", status)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	iErr := i.queries(ctx).UpsertIdempotencyKey(ctx, sqlc.UpsertIdempotencyKeyParams{
		IdemKey:       key,
		OperationType: string(operationType),
		Payload:       jsonPayload,
		Status:        string(status),
	})
	if iErr != nil {
		log.Println("Error upserting idempotency key:", iErr)
		return iErr
	}
	log.Println("Successfully upserted idempotency key:", key)
	return nil
}

// Update the status of an existing key
func (i *SqlcIdempotencyRepository) UpdateStatus(ctx context.Context, key string, status transaction.TransactionStatus) error {
	return i.queries(ctx).UpdateIdempotencyStatus(ctx, sqlc.UpdateIdempotencyStatusParams{
		IdemKey: key,
		Status:  string(status),
	})
}

// Mark as SUCCESS and save the final payload
func (i *SqlcIdempotencyRepository) SaveSuccess(ctx context.Context, key string) error {

	return i.queries(ctx).SaveIdempotencySuccess(ctx, key)
}

// Mark as FAILED and save the final payload/error
func (i *SqlcIdempotencyRepository) SaveFailure(ctx context.Context, key string) error {

	return i.queries(ctx).SaveIdempotencyFailure(ctx, key)
}
