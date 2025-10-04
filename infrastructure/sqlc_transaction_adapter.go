package infrastructure

import (
	"context"
	"fmt"

	"github.com/EfosaE/credora-backend/domain/transaction"
	"github.com/EfosaE/credora-backend/internal/db/sqlc"
	"github.com/EfosaE/credora-backend/internal/utils"
	// "github.com/google/uuid"
)

func NewSqlcTransactionRepository(ctx context.Context, q *sqlc.Queries) *SqlcRepository {
	return &SqlcRepository{
		q: q,
	}
}

// this SqlcRepository implements the TransactionRepository interface because it has all the methods defined in the interface
func (s *SqlcRepository) RecordTransaction(ctx context.Context, trx *transaction.NewTransactionInput) (*transaction.Transaction, error) {
	pgNumericAmount, _ := utils.DecimalToPgNumeric(trx.Amount)
	// metaBytes, err := json.Marshal(trx.Meta)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to marshal meta: %w", err)
	// }
	sqlcTransaction, err := s.q.RecordNewTransaction(ctx, sqlc.RecordNewTransactionParams{
		AccountID:   utils.ToPgUUID(trx.AccountID),
		Amount:      pgNumericAmount,
		Status:      string(trx.Status),
		Description: utils.ToPgText(trx.Description),
		Reference:   utils.ToPgText(trx.Reference),
		Channel:     utils.ToPgText(trx.Channel),
		Meta:        trx.Meta,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to record transaction: %w", err)
	}

	// Convert sqlc.Transaction to Transaction
	return toDomainTransaction(sqlcTransaction), nil
}

// toDomainTransaction maps sqlc.Transaction → domain.Transaction
func toDomainTransaction(sqlcTrx sqlc.Transaction) *transaction.Transaction {

	return &transaction.Transaction{
		ID:          sqlcTrx.ID,
		AccountID:   utils.FromPgUUID(sqlcTrx.AccountID),
		Amount:      utils.MustPgNumericToDecimal(sqlcTrx.Amount), // convert to decimal.Decimal
		Status:      transaction.TransactionStatus(sqlcTrx.Status),
		Description: sqlcTrx.Description.String,
		Reference:   sqlcTrx.Reference.String,
		Channel:     sqlcTrx.Channel.String,
		Meta:        sqlcTrx.Meta,
		CreatedAt:   sqlcTrx.CreatedAt.Time,
	}
}
