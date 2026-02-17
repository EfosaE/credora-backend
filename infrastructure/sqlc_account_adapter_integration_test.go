//go:build integration
// +build integration

package infrastructure_test

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/EfosaE/credora-backend/infrastructure"
	"github.com/EfosaE/credora-backend/internal/db/sqlc"
	"github.com/EfosaE/credora-backend/internal/utils"
	"github.com/EfosaE/credora-backend/test"
)

func TestCreditAccount(t *testing.T) {
	ctx := context.Background()

	// Spin up Postgres test DB
	pool, cleanup := test.SetupTestDB(t)
	defer cleanup()

	queries := sqlc.New(pool)

	// ---------- Seed user + account ----------
	user, err := queries.CreateUser(ctx, sqlc.CreateUserParams{
		FullName:    "Test User",
		PhoneNumber: "08011112222",
		Password:    "pass",
		Nin:         "12345678901",
	})
	require.NoError(t, err)

	acct, err := queries.CreateAccountWithMonnify(ctx, sqlc.CreateAccountWithMonnifyParams{
		UserID:             utils.ToPgUUID(user.ID),
		AccountNumber:      "3822115398",
		AccountType:        "TEST",
		MonnifyCustomerRef: utils.ToPgText("ref123"),
		VirtualAccountBank: utils.ToPgText("bank123"),
	})
	require.NoError(t, err)

	// Initial balance
	_, err = queries.CreditAccountBalance(ctx, sqlc.CreditAccountBalanceParams{
		Amount:        utils.NewPgNumericFromString("1000"),
		AccountNumber: acct.AccountNumber,
	})
	require.NoError(t, err)

	// ---------- Create repository ----------
	repo := infrastructure.NewSqlcAccountRepository(pool)

	creditAmt := decimal.NewFromInt(500)

	// ---------- Perform action inside transaction wrapper using TxManager ----------
	txManager := infrastructure.NewTxManager(pool)

	err = txManager.WithTx(ctx, func(txCtx context.Context) error {

		// Fetch original balance inside the transaction
		before, err := repo.GetAccountForUpdate(txCtx, acct.AccountNumber)
		require.NoError(t, err)

		beforeBal := before.Balance

		// Perform credit
		resp, err := repo.CreditAccount(txCtx, creditAmt, acct.AccountNumber)
		require.NoError(t, err)

		expected := beforeBal.Add(creditAmt)

		require.True(
			t,
			resp.Balance.Equal(expected),
			"expected balance %s, got %s",
			expected,
			resp.Balance,
		)

		return nil
	})

	require.NoError(t, err)
}
