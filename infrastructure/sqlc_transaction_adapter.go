package infrastructure

import (
	"context"
	"fmt"

	"github.com/EfosaE/credora-backend/domain/transaction"
	"github.com/EfosaE/credora-backend/internal/db/sqlc"
	"github.com/EfosaE/credora-backend/internal/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	// "github.com/google/uuid"
)

type SqlcTransactionRepository struct {
	db *pgxpool.Pool
	q  *sqlc.Queries
	tx pgx.Tx
}

func NewSqlcTransactionRepository(db *pgxpool.Pool) *SqlcTransactionRepository {
	return &SqlcTransactionRepository{
		db: db,
		q:  sqlc.New(db),
	}
}

func (r *SqlcTransactionRepository) Tx() pgx.Tx {
	return r.tx
}

// Bind repo to existing TX
func (r *SqlcTransactionRepository) WithTx(tx pgx.Tx) transaction.TransactionTx {
	return &SqlcTransactionRepository{
		db: r.db,
		tx: tx,
		q:  sqlc.New(tx),
	}
}

// RecordTransaction inside TX
// this SqlcRepository implements the TransactionRepository interface because it has all the methods defined in the interface
func (s *SqlcTransactionRepository) RecordTransaction(ctx context.Context, trx *transaction.NewTransactionInput) (*transaction.Transaction, error) {
	pgNumericAmount, _ := utils.DecimalToPgNumeric(trx.Amount)
	// metaBytes, err := json.Marshal(trx.Meta)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to marshal meta: %w", err)
	// }
	sqlcTransaction, err := s.q.RecordNewTransaction(ctx, sqlc.RecordNewTransactionParams{
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

	// Convert sqlc.Transaction to Transaction
	return toDomainTransaction(sqlcTransaction), nil
}

// toDomainTransaction maps sqlc.Transaction → domain.Transaction
func toDomainTransaction(sqlcTrx sqlc.Transaction) *transaction.Transaction {

	counterpartyID := utils.FromPgUUID(sqlcTrx.CounterpartyAccountID)
	return &transaction.Transaction{
		ID:             sqlcTrx.ID,
		AccountID:      utils.FromPgUUID(sqlcTrx.AccountID),
		CounterpartyID: &counterpartyID, // value can be null, so we use a pointer
		Amount:         utils.MustPgNumericToDecimal(sqlcTrx.Amount), // convert to decimal.Decimal
		Status:         transaction.TransactionStatus(sqlcTrx.Status),
		Description:    sqlcTrx.Description.String,
		Reference:      sqlcTrx.Reference.String,
		Channel:        sqlcTrx.Channel.String,
		Meta:           sqlcTrx.Meta,
		CreatedAt:      sqlcTrx.CreatedAt.Time,
	}
}
