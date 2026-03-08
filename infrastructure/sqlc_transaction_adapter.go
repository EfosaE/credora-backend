package infrastructure

import (
	"context"
	"fmt"

	"github.com/EfosaE/credora-backend/domain/transaction"
	"github.com/EfosaE/credora-backend/internal/db/sqlc"
	"github.com/EfosaE/credora-backend/internal/utils"
	"github.com/google/uuid"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SqlcTransactionRepository struct {
	db *pgxpool.Pool
}

func NewSqlcTransactionRepository(db *pgxpool.Pool) *SqlcTransactionRepository {
	return &SqlcTransactionRepository{
		db: db,
	}
}

func (r *SqlcTransactionRepository) queries(ctx context.Context) *sqlc.Queries {
	// Bind to tx in context if present, otherwise use main DB
	if tx, ok := GetTx(ctx); ok {
		return sqlc.New(tx)
	}
	return sqlc.New(r.db)
}

// ------------------------------------
// Normal + Transaction-safe methods
// ------------------------------------

func (r *SqlcTransactionRepository) RecordTransaction(
	ctx context.Context,
	trx *transaction.NewTransactionInput,
) (*transaction.Transaction, error) {

	pgNumericAmount, _ := utils.DecimalToPgNumeric(trx.Amount)

	sqlcTransaction, err := r.queries(ctx).RecordNewTransaction(ctx, sqlc.RecordNewTransactionParams{
		AccountID:             utils.ToPgUUID(trx.AccountID),
		CounterpartyAccountID: utils.ToPgNullableUUID(trx.CounterpartyID),
		Amount:                pgNumericAmount,
		Status:                string(trx.Status),
		Direction:             utils.ToPgText(string(trx.Direction)),
		Description:           utils.ToPgText(trx.Description),
		Reference:             utils.ToPgText(trx.Reference),
		Channel:               utils.ToPgText(trx.Channel),
		Meta:                  trx.Meta,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to record transaction: %w", err)
	}

	return toDomainTransaction(sqlcTransaction), nil
}

func (t *SqlcTransactionRepository) GetUserTransactions(
	ctx context.Context,
	userID uuid.UUID,
	cursor *transaction.Cursor,
	limit int32,
) (*[]transaction.Transaction, *transaction.Cursor, error) {

	params := sqlc.GetUserTransactionHistoryParams{
		UserID:    userID,
		PageLimit: limit,
	}

	if cursor != nil {
		params.CursorCreatedAt = utils.TimeToPgTimestampz(cursor.CreatedAt)
		params.CursorID = cursor.ID
	}

	sqlcTxns, err := t.queries(ctx).GetUserTransactionHistory(ctx, params)
	if err != nil {
		return nil, nil, err
	}

	// Convert sqlc slice to domain slice
	domainTxns := make([]transaction.Transaction, len(sqlcTxns))
	for i, t := range sqlcTxns {
		domainTxns[i] = *toDomainTransaction(t)
	}

	// Build next cursor
	var nextCursor *transaction.Cursor
	if len(domainTxns) > 0 {
		last := domainTxns[len(domainTxns)-1]
		nextCursor = &transaction.Cursor{
			CreatedAt: last.CreatedAt,
			ID:        last.ID,
		}
	}

	return &domainTxns, nextCursor, nil
}

// ------------------------------------
// Helpers
// ------------------------------------

func toDomainTransaction(sqlcTrx sqlc.Transaction) *transaction.Transaction {
	counterpartyID := utils.FromPgUUID(sqlcTrx.CounterpartyAccountID)

	return &transaction.Transaction{
		ID:             sqlcTrx.ID,
		AccountID:      utils.FromPgUUID(sqlcTrx.AccountID),
		CounterpartyID: &counterpartyID,
		Direction:      transaction.TransactionType(sqlcTrx.Direction.String),
		Amount:         utils.MustPgNumericToDecimal(sqlcTrx.Amount),
		Status:         transaction.TransactionStatus(sqlcTrx.Status),
		Description:    sqlcTrx.Description.String,
		Reference:      sqlcTrx.Reference.String,
		Channel:        sqlcTrx.Channel.String,
		Meta:           sqlcTrx.Meta,
		CreatedAt:      sqlcTrx.CreatedAt.Time,
	}
}
