package infrastructure

import (
	"context"
	"errors"
	"fmt"

	"github.com/EfosaE/credora-backend/domain/account"
	"github.com/EfosaE/credora-backend/domain/operation"
	"github.com/EfosaE/credora-backend/internal/db/sqlc"
	"github.com/EfosaE/credora-backend/internal/utils"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

// SqlcAccountRepository implements both AccountRepository and AccountTx
type SqlcAccountRepository struct {
	db *pgxpool.Pool
}

func NewSqlcAccountRepository(db *pgxpool.Pool) *SqlcAccountRepository {
	return &SqlcAccountRepository{
		db: db,
	}
}

func (r *SqlcAccountRepository) queries(ctx context.Context) *sqlc.Queries {
	// create a repo bound to a transaction if it exists in the context, otherwise use the main DB connection
	if tx, ok := GetTx(ctx); ok {
		return sqlc.New(tx)
	}
	return sqlc.New(r.db)
}

// ------------------------------------
// Normal methods
// ------------------------------------

func (r *SqlcAccountRepository) CreateAcct(ctx context.Context, req *account.CreateAccountRequest) (*account.Account, error) {
	sqlcAcct, err := r.queries(ctx).CreateAccountWithMonnify(ctx, sqlc.CreateAccountWithMonnifyParams{
		UserID:             utils.ToPgUUID(req.UserId),
		Username:           req.Username,
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

func (r *SqlcAccountRepository) GetUserByAccountNumber(ctx context.Context, acct string) (*account.GetUserDetailsWithAccountRow, error) {
	row, err := r.queries(ctx).GetUserByAccountNumber(ctx, acct)
	if err != nil {
		return nil, err
	}
	return &account.GetUserDetailsWithAccountRow{
		UserId:        row.ID,
		Password:      row.Password,
		FullName:      row.FullName,
		Email:         row.Email.String,
		PhoneNumber:   row.PhoneNumber,
		IsVerified:    row.IsVerified.Bool,
		AccountNumber: row.AccountNumber,
		AccountType:   row.AccountType,
		Balance:       utils.MustPgNumericToDecimal(row.Balance).String(),
		Currency:      row.Currency,
	}, nil
}

func (r *SqlcAccountRepository) GetAccountByAccountNumber(ctx context.Context, acct string) (*account.Account, error) {
	row, err := r.queries(ctx).GetAccountByAccountNumber(ctx, acct)
	if err != nil {
		return nil, err
	}
	return &account.Account{
		ID:            row.ID,
		AccountNumber: row.AccountNumber,
		Balance:       utils.MustPgNumericToDecimal(row.Balance),
	}, nil
}

// ------------------------------------
// Transaction-bound methods
// ------------------------------------

func (r *SqlcAccountRepository) GetAccountsForUpdate(
	ctx context.Context,
	accountNumbers []string,
) ([]*account.Account, error) {

	rows, err := r.queries(ctx).GetAccountsForUpdate(ctx, accountNumbers)
	if err != nil {
		return nil, err
	}

	accounts := make([]*account.Account, 0, len(rows))

	for _, row := range rows {
		accounts = append(accounts, &account.Account{
			ID:            row.ID,
			UserName:      row.Username,
			UserId:        row.UserID.String(),
			AccountNumber: row.AccountNumber,
			Balance:       utils.MustPgNumericToDecimal(row.Balance),
		})
	}

	// Safety check: ensure all requested accounts were locked/found
	if len(accounts) != len(accountNumbers) {
		return nil, fmt.Errorf("one or more accounts not found or locked")
	}

	return accounts, nil
}

func (r *SqlcAccountRepository) GetAccountForUpdate(ctx context.Context, accountNumber string) (*account.Account, error) {
	row, err := r.queries(ctx).GetAccountForUpdate(ctx, accountNumber)
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

	row, err := r.queries(ctx).CreditAccountBalance(ctx, sqlc.CreditAccountBalanceParams{
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

	row, err := r.queries(ctx).DebitAccountBalance(ctx, sqlc.DebitAccountBalanceParams{
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

func (r *SqlcAccountRepository) InternalMoneyTransfer(
	ctx context.Context,
	amount decimal.Decimal,
	fromAcctNum, toAcctNum string,
) (*account.InternalTransferResp, error) {

	pgAmount, _ := utils.DecimalToPgNumeric(amount)

	result, err := r.queries(ctx).TransferMoneyInternal(ctx,
		sqlc.TransferMoneyInternalParams{
			Amount:      pgAmount,
			FromAccount: fromAcctNum,
			ToAccount:   toAcctNum,
		},
	)

	if err != nil {
		// --- DB → Domain mapping ---
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, operation.ErrAccountNotFound    // Basically, no account found that satisfies the defined query contraints, insufficient balnce or the accoutn number doesnt exist
		}

		return nil, err
	}

	return &account.InternalTransferResp{
		FromAccountId: result.FromID,
		ToAccountId:   result.ToID,
		FromBalance:   utils.MustPgNumericToDecimal(result.FromBalance),
		ToBalance:     utils.MustPgNumericToDecimal(result.ToBalance),
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
