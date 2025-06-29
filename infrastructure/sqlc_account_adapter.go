package infrastructure

import (
	"context"

	"github.com/EfosaE/credora-backend/domain/account"
	"github.com/EfosaE/credora-backend/internal/db/sqlc"
	"github.com/EfosaE/credora-backend/internal/utils"
	// "github.com/google/uuid"
)

func NewSqlcAccountRepository(ctx context.Context, q *sqlc.Queries) *SqlcUserRepository {
	return &SqlcUserRepository{
		q: q,
	}
}

// this SqlcUserRepository implements the UserRepository interface because it has all the methods defined in the interface
func (s *SqlcUserRepository) CreateAcct(ctx context.Context, acct *account.CreateAccountRequest) (*account.Account, error) {
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
