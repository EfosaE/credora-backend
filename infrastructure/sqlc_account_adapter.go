package infrastructure

import (
	"context"

	"github.com/EfosaE/credora-backend/domain/account"
	"github.com/EfosaE/credora-backend/internal/db/sqlc"
	"github.com/EfosaE/credora-backend/internal/utils"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

// SqlcAccountRepository implements both AccountRepository and AccountTx
type SqlcAccountRepository struct {
	db *pgxpool.Pool
	tx pgx.Tx      // nil when not inside a transaction
	q  *sqlc.Queries
}

func NewSqlcAccountRepository(db *pgxpool.Pool) *SqlcAccountRepository {
	return &SqlcAccountRepository{
		db: db,
		q:  sqlc.New(db),
	}
}

// create a repo bound to a transaction
func (r *SqlcAccountRepository) withTx(tx pgx.Tx) *SqlcAccountRepository {
	return &SqlcAccountRepository{
		db: r.db,
		tx: tx,
		q:  sqlc.New(tx),
	}
}

func (r *SqlcAccountRepository) Tx() pgx.Tx {
	return r.tx
}

// ------------------------------------
// Transaction wrapper
// ------------------------------------
func (r *SqlcAccountRepository) WithTx(ctx context.Context, fn func(tx account.AccountTx) error) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	// bind repo to transaction
	txRepo := r.withTx(tx)

	if err := fn(txRepo); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

// ------------------------------------
// Normal methods
// ------------------------------------

func (r *SqlcAccountRepository) CreateAcct(ctx context.Context, req *account.CreateAccountRequest) (*account.Account, error) {
	sqlcAcct, err := r.q.CreateAccountWithMonnify(ctx, sqlc.CreateAccountWithMonnifyParams{
		UserID:             utils.ToPgUUID(req.UserId),
		AccountNumber:      req.AccountNumber,
		AccountType:        req.AccountType,
		MonnifyCustomerRef: utils.ToPgText(req.MonnifyCustRef),
		VirtualAccountBank: utils.ToPgText(req.BankName),
	})
	if err != nil {
		return nil, err
	}

	return toDomain(sqlcAcct), nil
}

func (r *SqlcAccountRepository) GetUserByAccountNumber(ctx context.Context, acct string) (*sqlc.GetUserByAccountNumberRow, error) {
	row, err := r.q.GetUserByAccountNumber(ctx, acct)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// ------------------------------------
// Transaction-bound methods
// ------------------------------------

func (r *SqlcAccountRepository) GetAccountForUpdate(ctx context.Context, accountNumber string) (*account.Account, error) {
	row, err := r.q.GetAccountForUpdate(ctx, accountNumber)
	if err != nil {
		return nil, err
	}

	return &account.Account{
		ID:            row.ID,
		UserId:        row.UserID.String(),
		AccountNumber: row.AccountNumber,
		Balance:       utils.MustPgNumericToDecimal(row.Balance),
	}, nil
}

func (r *SqlcAccountRepository) CreditAccount(ctx context.Context, amount decimal.Decimal, accountNumber string) (*account.CreditAcctResp, error) {
	pgAmount, _ := utils.DecimalToPgNumeric(amount)

	row, err := r.q.CreditAccountBalance(ctx, sqlc.CreditAccountBalanceParams{
		Amount:        pgAmount,
		AccountNumber: accountNumber,
	})
	if err != nil {
		return nil, err
	}

	bal, _ := utils.PgNumericToDecimal(row.Balance)

	return &account.CreditAcctResp{
		AcctId:  row.ID,
		Balance: bal,
	}, nil
}

func (r *SqlcAccountRepository) DebitAccount(ctx context.Context, amount decimal.Decimal, accountNumber string) (*account.CreditAcctResp, error) {
	pgAmount, _ := utils.DecimalToPgNumeric(amount)

	row, err := r.q.DebitAccountBalance(ctx, sqlc.DebitAccountBalanceParams{
		Amount:        pgAmount,
		AccountNumber: accountNumber,
	})
	if err != nil {
		return nil, err
	}

	bal, _ := utils.PgNumericToDecimal(row.Balance)

	return &account.CreditAcctResp{
		AcctId:  row.ID,
		Balance: bal,
	}, nil
}

// ------------------------------------
// Helpers
// ------------------------------------

func toDomain(a sqlc.Account) *account.Account {
	return &account.Account{
		ID:             a.ID,
		UserId:         a.UserID.String(),
		AccountNumber:  a.AccountNumber,
		AccountType:    a.AccountType,
		MonnifyCustRef: a.MonnifyCustomerRef.String,
		BankName:       a.VirtualAccountBank.String,
		CreatedAt:      a.CreatedAt.Time,
		UpdatedAt:      a.UpdatedAt.Time,
	}
}
