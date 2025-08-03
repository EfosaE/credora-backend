package infrastructure

import (
	"context"
	"fmt"

	"github.com/EfosaE/credora-backend/domain/account"
	"github.com/EfosaE/credora-backend/internal/db/sqlc"
	"github.com/EfosaE/credora-backend/internal/utils"
	"github.com/shopspring/decimal"
	// "github.com/google/uuid"
)

func NewSqlcAccountRepository(ctx context.Context, q *sqlc.Queries) *SqlcRepository {
	return &SqlcRepository{
		q: q,
	}
}

// this SqlcRepository implements the AccountRepository interface because it has all the methods defined in the interface
func (s *SqlcRepository) CreateAcct(ctx context.Context, acct *account.CreateAccountRequest) (*account.Account, error) {
	sqlcAccount, err := s.q.CreateAccountWithMonnify(ctx, sqlc.CreateAccountWithMonnifyParams{
		UserID:             utils.ConvertUUID(acct.UserId),
		AccountNumber:      acct.AccountNumber,
		AccountType:        acct.AccountType,
		MonnifyCustomerRef: utils.ToPgText(acct.MonnifyCustRef),
		VirtualAccountBank: utils.ToPgText(acct.BankName),
	})

	if err != nil {
		return nil, err
	}

	// Convert sqlc.User to User
	return toDomainAccount(sqlcAccount), nil
}

func (s *SqlcRepository) GetUserByAccountNumber(ctx context.Context, accountNumber string) (*sqlc.GetUserByAccountNumberRow, error) {
	result, err := s.q.GetUserByAccountNumber(ctx, accountNumber)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *SqlcRepository) CreditAccount(ctx context.Context, amount decimal.Decimal, accountNumber string) (*account.CreditAcctResp, error) {

	numericAmount, _ := utils.DecimalToPgNumeric(amount)
	result, err := s.q.CreditAccountBalance(ctx, sqlc.CreditAccountBalanceParams{
		Amount:        numericAmount,
		AccountNumber: accountNumber,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to credit account: %w", err)
	}

	numericBal, err := utils.PgNumericToDecimal(result.Balance)
	if err != nil {
		return nil, fmt.Errorf("failed to convert amount to numeric: %w", err)
	}
	return &account.CreditAcctResp{
		AcctId:  result.ID,
		Balance: numericBal,
	}, nil
}

func toDomainAccount(sqlcAcct sqlc.Account) *account.Account {
	return &account.Account{
		ID:             sqlcAcct.ID,
		UserId:         sqlcAcct.UserID.String(),
		AccountNumber:  sqlcAcct.AccountNumber,
		AccountType:    sqlcAcct.AccountType,
		MonnifyCustRef: sqlcAcct.MonnifyCustomerRef.String,
		BankName:       sqlcAcct.VirtualAccountBank.String,
		CreatedAt:      sqlcAcct.CreatedAt.Time,
		UpdatedAt:      sqlcAcct.UpdatedAt.Time,
	}
}
