package infrastructure

import (
	"context"
	"encoding/json"
	"log"

	"github.com/EfosaE/credora-backend/domain/idempotency"
	"github.com/EfosaE/credora-backend/domain/operation"
	"github.com/EfosaE/credora-backend/domain/transaction"
	"github.com/EfosaE/credora-backend/internal/db/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SqlcIdempotencyRepository implements both IdempotencyRepo and IdempotencyTx
type SqlcIdempotencyRepository struct {
	db *pgxpool.Pool
	tx pgx.Tx // nil when not inside a transaction
	q  *sqlc.Queries
}

func NewSqlcIdempotencyRepository(db *pgxpool.Pool) *SqlcIdempotencyRepository {
	return &SqlcIdempotencyRepository{
		db: db,
		q:  sqlc.New(db),
	}
}

// create a repo bound to a transaction
func (r *SqlcIdempotencyRepository) WithTx(tx pgx.Tx) idempotency.IdempotencyTx {
	return &SqlcIdempotencyRepository{
		db: r.db,
		tx: tx,
		q:  sqlc.New(tx),
	}
}

// ------------------------------------
// Normal methods / Tx methods
// ------------------------------------

// Check if an idempotency key exists
func (i *SqlcIdempotencyRepository) Check(ctx context.Context, key string) (bool, error) {
	return i.q.CheckIdempotency(ctx, key)
}

// Insert a new idempotency key with a status
func (i *SqlcIdempotencyRepository) Insert(ctx context.Context, key string, opType operation.OperationType, payload any, status transaction.TransactionStatus) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return i.q.InsertIdempotencyKey(ctx, sqlc.InsertIdempotencyKeyParams{
		IdemKey:       key,
		OperationType: string(opType),
		Payload:       b,
		Status:        string(status),
	})
}

// Get an idempotency record
func (i *SqlcIdempotencyRepository) Get(ctx context.Context, key string) (*idempotency.IdempotencyData, error) {
	data, err := i.q.GetIdempotencyKey(ctx, key)
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
	return i.q.DeleteIdempotencyKey(ctx, key)
}

// Upsert an idempotency key with status
func (i *SqlcIdempotencyRepository) Upsert(ctx context.Context, key string, operationType operation.OperationType, payload any, status transaction.TransactionStatus) error {
	log.Println("Upserting idempotency key:", key, "with status:", status)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return i.q.UpsertIdempotencyKey(ctx, sqlc.UpsertIdempotencyKeyParams{
		IdemKey:       key,
		OperationType: string(operationType),
		Payload:       jsonPayload,
		Status:        string(status),
	})
}

// Update the status of an existing key
func (i *SqlcIdempotencyRepository) UpdateStatus(ctx context.Context, key string, status transaction.TransactionStatus) error {
	return i.q.UpdateIdempotencyStatus(ctx, sqlc.UpdateIdempotencyStatusParams{
		IdemKey: key,
		Status:  string(status),
	})
}

// Mark as SUCCESS and save the final payload
func (i *SqlcIdempotencyRepository) SaveSuccess(ctx context.Context, key string) error {

	return i.q.SaveIdempotencySuccess(ctx, key)
}

// Mark as FAILED and save the final payload/error
func (i *SqlcIdempotencyRepository) SaveFailure(ctx context.Context, key string) error {

	return i.q.SaveIdempotencyFailure(ctx, key)
}

//Unless the IdempotencyTx interface requires access to the underlying pgx.Tx, this method is unnecessary.
// func (r *SqlcIdempotencyRepository) Tx() pgx.Tx {
// 	return r.tx
// }

// ------------------------------------
// Transaction wrapper
// ------------------------------------
// func (r *SqlcIdempotencyRepository) WithTx(ctx context.Context, fn func(tx idempotency.IdempotencyTx) error) error {
// 	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
// 	if err != nil {
// 		return err
// 	}

// 	// bind repo to transaction
// 	txRepo := r.withTx(tx)

// 	if err := fn(txRepo); err != nil {
// 		_ = tx.Rollback(ctx)
// 		return err
// 	}

// 	return tx.Commit(ctx)
// }
