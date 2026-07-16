package infrastructure

import (
	"context"

	"github.com/EfosaE/credora-backend/domain/account"
	domainerr "github.com/EfosaE/credora-backend/domain/domianerrors"
	"github.com/EfosaE/credora-backend/domain/user"
	"github.com/EfosaE/credora-backend/internal/db/sqlc"
	"github.com/EfosaE/credora-backend/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SqlcUserRepository struct {
	db *pgxpool.Pool
}

func NewSqlcUserRepository(db *pgxpool.Pool) *SqlcUserRepository {
	return &SqlcUserRepository{
		db: db,
	}
}

func (r *SqlcUserRepository) queries(ctx context.Context) *sqlc.Queries {
	if tx, ok := GetTx(ctx); ok {
		return sqlc.New(tx)
	}
	return sqlc.New(r.db)
}

// ------------------------------------
// Methods
// ------------------------------------

func (r *SqlcUserRepository) Create(ctx context.Context, u *user.CreateUserRequest) (*user.User, error) {
	sqlcUser, err := r.queries(ctx).CreateUser(ctx, sqlc.CreateUserParams{
		FullName:    u.Name,
		Email:       u.Email,
		Password:    u.Password,
		PhoneNumber: u.PhoneNumber,
		Nin:         u.Nin,
	})
	if err != nil {
		return nil, err
	}

	return toDomainUser(sqlcUser), nil
}

func (r *SqlcUserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	sqlcUser, err := r.queries(ctx).GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return toDomainUser(sqlcUser), nil
}

func (r *SqlcUserRepository) VerifyUser(ctx context.Context, id uuid.UUID) error {
	err := r.queries(ctx).VerifyUser(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (r *SqlcUserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, hashedPassword string) error {
	return r.queries(ctx).UpdateUserPassword(ctx, sqlc.UpdateUserPasswordParams{
		ID:       id,
		Password: hashedPassword,
	})
}

func (r *SqlcUserRepository) GetUserAccountsByAccountNumber(
	ctx context.Context,
	acct string,
) (*user.User, error) {

	rows, err := r.queries(ctx).GetUserWithAccountsByAccountNumber(ctx, acct)
	if err != nil {
		return nil, err
	}

	return mapUserAccounts(rows)
}

func (r *SqlcUserRepository) GetUserAccountsByEmail(
	ctx context.Context,
	email string,
) (*user.User, error) {

	rows, err := r.queries(ctx).GetUserWithAccountsByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return mapUserAccountsEmailMethod(rows)
}

func (r *SqlcUserRepository) GetUserAccountsByUserID(
	ctx context.Context,
	userID uuid.UUID,
) (*user.User, error) {

	rows, err := r.queries(ctx).GetUserWithAccountsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return mapUserAccountsByUserID(rows)
}

// ------------------------------------
// Helpers
// ------------------------------------

func toDomainUser(sqlcUser sqlc.User) *user.User {
	return &user.User{
		ID:        sqlcUser.ID,
		FullName:  sqlcUser.FullName,
		Email:     sqlcUser.Email,
		CreatedAt: sqlcUser.CreatedAt.Time,
		UpdatedAt: sqlcUser.UpdatedAt.Time,
	}
}

func mapUserAccounts(rows []sqlc.GetUserWithAccountsByAccountNumberRow) (*user.User, error) {

	if len(rows) == 0 {
		return nil, domainerr.ErrNoRowsFound
	}

	u := &user.User{
		ID:          rows[0].ID,
		Password:    rows[0].Password,
		FullName:    rows[0].FullName,
		Email:       rows[0].Email,
		PhoneNumber: rows[0].PhoneNumber,
		IsVerified:  rows[0].IsVerified.Bool,
		Accounts:    make([]account.Account, 0, len(rows)),
	}

	for _, r := range rows {
		acc := account.Account{
			ID:            r.AccountID,
			UserId:        r.ID,
			AccountNumber: r.AccountNumber,
			UserName:      r.FullName,
			AccountType:   r.AccountType,
			Balance:       utils.MustPgNumericToDecimal(r.Balance),
			BankName:      r.VirtualAccountBank.String,
			Currency:      r.Currency,
		}

		u.Accounts = append(u.Accounts, acc)
	}

	return u, nil
}

func mapUserAccountsByUserID(rows []sqlc.GetUserWithAccountsByUserIDRow) (*user.User, error) {

	if len(rows) == 0 {
		return nil, domainerr.ErrNoRowsFound
	}

	u := &user.User{
		ID:          rows[0].ID,
		Password:    rows[0].Password,
		FullName:    rows[0].FullName,
		Email:       rows[0].Email,
		PhoneNumber: rows[0].PhoneNumber,
		IsVerified:  rows[0].IsVerified.Bool,
		Accounts:    make([]account.Account, 0, len(rows)),
	}

	for _, r := range rows {
		acc := account.Account{
			ID:            r.AccountID,
			UserId:        r.ID,
			UserName:      r.FullName,
			AccountNumber: r.AccountNumber,
			AccountType:   r.AccountType,
			Balance:       utils.MustPgNumericToDecimal(r.Balance),
			BankName:      r.VirtualAccountBank.String,
			Currency:      r.Currency,
		}

		u.Accounts = append(u.Accounts, acc)
	}

	return u, nil
}

func mapUserAccountsEmailMethod(rows []sqlc.GetUserWithAccountsByEmailRow) (*user.User, error) {

	if len(rows) == 0 {
		return nil, domainerr.ErrNoRowsFound
	}

	u := &user.User{
		ID:          rows[0].ID,
		FullName:    rows[0].FullName,
		Email:       rows[0].Email,
		Password:    rows[0].Password,
		PhoneNumber: rows[0].PhoneNumber,
		IsVerified:  rows[0].IsVerified.Bool,
		Accounts:    make([]account.Account, 0, len(rows)),
	}

	for _, r := range rows {

		// If LEFT JOIN produced a NULL account row, skip it
		if !r.AccountNumber.Valid {
			continue
		}

		acc := account.Account{
			UserId:        r.ID,
			UserName:      r.FullName,
			AccountNumber: r.AccountNumber.String,
			AccountType:   r.AccountType.String,
			Balance:       utils.MustPgNumericToDecimal(r.Balance),
			BankName:      r.VirtualAccountBank.String,
			Currency:      r.Currency.String,
		}

		u.Accounts = append(u.Accounts, acc)
	}

	return u, nil
}

// To ensure this repo satisfies the interface defined in the domain.
var _ user.UserRepository = (*SqlcUserRepository)(nil)
