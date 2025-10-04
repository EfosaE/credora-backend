//go:build integration
// +build integration

package infrastructure

import (
	"context"
	"fmt"
	"testing"

	"github.com/EfosaE/credora-backend/internal/db"
	"github.com/EfosaE/credora-backend/internal/utils"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestCreditAccount(t *testing.T) {
	ctx := context.Background()
	acctNum := "3822115398"

	testDB, err := db.InitTestDB(ctx)
	require.NoError(t, err)
	defer testDB.Pool.Close()

	tx, err := testDB.Pool.Begin(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	queries := testDB.Queries.WithTx(tx)
	acctRepo := NewSqlcAccountRepository(ctx, queries)

	// --- Setup: get user account ---
	acct, err := acctRepo.GetUserByAccountNumber(ctx, acctNum)
	require.NoError(t, err)

	// --- Action: credit ₦500 ---
	creditAmt := decimal.NewFromFloat(500)
	resp, err := acctRepo.CreditAccount(ctx, creditAmt, acctNum)
	// fmt.Printf("%s, %s/n", resp.AcctId, acct.ID)
	require.NoError(t, err)
	// require.Equal(t, acct.ID, resp.AcctId)

	// --- Assert: balance increased correctly ---
	originalBal := utils.MustPgNumericToDecimal(acct.Balance)
	fmt.Println(acct.Balance)
	fmt.Println(originalBal)
	expectedBal := originalBal.Add(creditAmt)
	fmt.Println(expectedBal)
	require.True(t, resp.Balance.Equal(expectedBal), "expected %s, got %s", expectedBal, resp.Balance)
}
