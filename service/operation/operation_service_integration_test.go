//go:build integration
// +build integration

package operationsvc_test

import (
	"context"
	// "log"

	"testing"

	"github.com/stretchr/testify/require"

	"github.com/EfosaE/credora-backend/domain/operation"
	operationsvc "github.com/EfosaE/credora-backend/service/operation"
	"github.com/EfosaE/credora-backend/test"
	"github.com/EfosaE/credora-backend/test/mocks"

	"github.com/EfosaE/credora-backend/infrastructure"
	"github.com/EfosaE/credora-backend/internal/db/sqlc"
	"github.com/EfosaE/credora-backend/internal/utils"
)

func TestInternalTransfer_Success(t *testing.T) {
	ctx := context.Background()
	pool, cleanup := test.SetupTestDB(t)
	defer cleanup()

	queries := sqlc.New(pool)

	// Create users first
	user1, err := queries.CreateUser(ctx, sqlc.CreateUserParams{
		FullName:    "John Doe",
		PhoneNumber: "08011112222",
		Password:    "hashedpassword789",
		Nin:         "12345678908",
	})
	require.NoError(t, err)

	user2, err := queries.CreateUser(ctx, sqlc.CreateUserParams{
		FullName:    "Jane Doe",
		PhoneNumber: "08033334444",
		Password:    "hashedpassword012",
		Nin:         "12345678956",
	})
	require.NoError(t, err)

	// seed accounts
	fromAcct, err := queries.CreateAccountWithMonnify(ctx, sqlc.CreateAccountWithMonnifyParams{
		UserID:             utils.ToPgUUID(user1.ID),
		AccountNumber:      "1111111111",
		AccountType:        "TESTING ACCOUNT",
		MonnifyCustomerRef: utils.ToPgText("ref-111"),
		VirtualAccountBank: utils.ToPgText("bank-111"),
	})
	require.NoError(t, err)

	toAcct, err := queries.CreateAccountWithMonnify(ctx, sqlc.CreateAccountWithMonnifyParams{
		AccountNumber:      "2222222222",
		UserID:             utils.ToPgUUID(user2.ID),
		AccountType:        "TESTING ACCOUNT",
		MonnifyCustomerRef: utils.ToPgText("ref-222"),
		VirtualAccountBank: utils.ToPgText("bank-222"),
	})
	require.NoError(t, err)

	// Add initial balances
	_, err = queries.CreditAccountBalance(ctx, sqlc.CreditAccountBalanceParams{
		Amount:        utils.NewPgNumericFromString("5000"),
		AccountNumber: fromAcct.AccountNumber,
	})

	require.NoError(t, err)

	_, err = queries.CreditAccountBalance(ctx, sqlc.CreditAccountBalanceParams{
		Amount:        utils.NewPgNumericFromString("2000"),
		AccountNumber: toAcct.AccountNumber,
	})

	require.NoError(t, err)

	// repos
	acctRepo := infrastructure.NewSqlcAccountRepository(pool)
	txRepo := infrastructure.NewSqlcTransactionRepository(pool)
	idempRepo := infrastructure.NewSqlcIdempotencyRepository(pool)
	eventBus := &mocks.MockEventBus{}
	// idemp := mocks.NewMockIdempotencyRepo()

	// transaction manager used by service for running operations inside a db tx
	txManager := infrastructure.NewTxManager(pool)

	svc := operationsvc.NewOperationService(
		txManager,
		acctRepo,
		txRepo,
		idempRepo,
		test.SetupTestLogger(),
		eventBus,
	)

	req := &operation.InternalTransferReq{
		FromAcctNum:    "1111111111",
		ToAcctNum:      "2222222222",
		Amount:         utils.MustPgNumericToDecimal(utils.NewPgNumericFromString("2000")),
		IdempotencyKey: "idem-xxx",
	}

	err = svc.InternalTransfer(ctx, req)
	require.NoError(t, err)

	// check balances
	updatedFrom, _ := queries.GetAccountByAccountNumber(ctx, fromAcct.AccountNumber)
	updatedTo, _ := queries.GetAccountByAccountNumber(ctx, toAcct.AccountNumber)
	t.Log("from:", updatedFrom)
	t.Log("to:", updatedTo)

	require.Equal(t, "3000.00", utils.PgNumericToString(updatedFrom.Balance))
	require.Equal(t, "4000.00", utils.PgNumericToString(updatedTo.Balance))

	// ledger entries
	// ledgers, err := queries.ListTransactions(ctx, fromAcct.ID)
	// require.NoError(t, err)
	// require.Len(t, ledgers, 1)
	// require.Equal(t, "INTERNAL_TRANSFER", ledgers[0].Channel)

	// idempotency
	_, errP := idempRepo.Get(ctx, "idem-xxx")
	require.NoError(t, errP)
}

func TestInternalTransfer_RollbackOnFailure(t *testing.T) {
	ctx := context.Background()
	pool, cleanup := test.SetupTestDB(t)
	defer cleanup()

	queries := sqlc.New(pool)

	// Create two users
	user1, err := queries.CreateUser(ctx, sqlc.CreateUserParams{
		FullName:    "John Doe",
		PhoneNumber: "08011112222",
		Password:    "hashedpassword789",
		Nin:         "12345678908",
	})
	require.NoError(t, err)

	user2, err := queries.CreateUser(ctx, sqlc.CreateUserParams{
		FullName:    "Jane Doe",
		PhoneNumber: "08033334444",
		Password:    "hashedpassword012",
		Nin:         "12345678956",
	})
	require.NoError(t, err)

	// Create accounts
	fromAcct, err := queries.CreateAccountWithMonnify(ctx, sqlc.CreateAccountWithMonnifyParams{
		UserID:             utils.ToPgUUID(user1.ID),
		AccountNumber:      "1111111111",
		AccountType:        "TESTING ACCOUNT",
		MonnifyCustomerRef: utils.ToPgText("ref-111"),
		VirtualAccountBank: utils.ToPgText("bank-111"),
	})
	require.NoError(t, err)

	toAcct, err := queries.CreateAccountWithMonnify(ctx, sqlc.CreateAccountWithMonnifyParams{
		AccountNumber:      "2222222222",
		UserID:             utils.ToPgUUID(user2.ID),
		AccountType:        "TESTING ACCOUNT",
		MonnifyCustomerRef: utils.ToPgText("ref-222"),
		VirtualAccountBank: utils.ToPgText("bank-222"),
	})
	require.NoError(t, err)

	// Seed initial balances
	_, err = queries.CreditAccountBalance(ctx, sqlc.CreditAccountBalanceParams{
		Amount:        utils.NewPgNumericFromString("5000"),
		AccountNumber: fromAcct.AccountNumber,
	})
	require.NoError(t, err)

	_, err = queries.CreditAccountBalance(ctx, sqlc.CreditAccountBalanceParams{
		Amount:        utils.NewPgNumericFromString("2000"),
		AccountNumber: toAcct.AccountNumber,
	})
	require.NoError(t, err)

	// Setup repos & service
	acctRepo := infrastructure.NewSqlcAccountRepository(pool)
	txRepo := infrastructure.NewSqlcTransactionRepository(pool)
	idempRepo := infrastructure.NewSqlcIdempotencyRepository(pool)
	eventBus := &mocks.MockEventBus{}
	txManager := infrastructure.NewTxManager(pool)

	svc := operationsvc.NewOperationService(
		txManager,
		acctRepo,
		txRepo,
		idempRepo,
		test.SetupTestLogger(),
		eventBus,
	)

	// Attempt a transfer that should fail (insufficient funds)
	req := &operation.InternalTransferReq{
		FromAcctNum:    "1111111111",
		ToAcctNum:      "2222222222",
		Amount:         utils.MustPgNumericToDecimal(utils.NewPgNumericFromString("7000")), // too high
		IdempotencyKey: "idem-fail-1",
	}

	err = svc.InternalTransfer(ctx, req)
	require.Error(t, err)
	require.ErrorIs(t, err, operation.ErrInsufficientFunds)

	// Reload accounts from DB
	updatedFrom, _ := queries.GetAccountByAccountNumber(ctx, fromAcct.AccountNumber)
	updatedTo, _ := queries.GetAccountByAccountNumber(ctx, toAcct.AccountNumber)

	// Ensure balances were NOT changed
	require.Equal(t, "5000.00", utils.PgNumericToString(updatedFrom.Balance))
	require.Equal(t, "2000.00", utils.PgNumericToString(updatedTo.Balance))

	// Ensure no ledger transaction was created for the failed transfer
	// txs, err := queries.ListTransactions(ctx, fromAcct.ID)
	// require.NoError(t, err)
	// require.Len(t, txs, 0)

	// Idempotency should be stored as FAILED
	idem, err := idempRepo.Get(ctx, "idem-fail-1")
	require.NoError(t, err)
	require.Equal(t, "FAILED", string(idem.Status))

	// ---- Retry same idempotency key → should return same error ----
	// err = svc.InternalTransfer(ctx, req)
	// require.Error(t, err)
	// require.ErrorIs(t, err, operation.ErrInsufficientFunds)

	// ---- Ensure balances still unchanged after retry ----
	// updatedFrom2, _ := queries.GetAccountByAccountNumber(ctx, fromAcct.AccountNumber)
	// updatedTo2, _ := queries.GetAccountByAccountNumber(ctx, toAcct.AccountNumber)
	// require.Equal(t, "5000.00", utils.PgNumericToString(updatedFrom2.Balance))
	// require.Equal(t, "2000.00", utils.PgNumericToString(updatedTo2.Balance))
}

// func TestInternalTransfer_InsufficientFunds(t *testing.T) {
// 	ctx := context.Background()
// 	pool, cleanup := setupDB(t)
// 	defer cleanup()

// 	queries := db.New(pool)

// 	// seed accounts
// 	fromAcct, _ := queries.CreateAccount(ctx, db.CreateAccountParams{
// 		AccountNumber: "111",
// 		Balance:       utils.NewMoney("1000"),
// 	})
// 	toAcct, _ := queries.CreateAccount(ctx, db.CreateAccountParams{
// 		AccountNumber: "222",
// 		Balance:       utils.NewMoney("1000"),
// 	})

// 	// repos
// 	acctRepo := infrastructure.NewSqlcAccountRepository(pool)
// 	txRepo := infrastructure.NewSqlcTransactionRepository(pool)
// 	idem := infrastructure.NewIdempotencyCache()
// 	svc := operationsvc.NewOperationService(acctRepo, txRepo, idem, logger.NewFakeLogger())

// 	req := &operation.InternalTransferReq{
// 		FromAcctNum:    "111",
// 		ToAcctNum:      "222",
// 		Amount:         utils.NewMoney("5000"), // too much
// 		IdempotencyKey: "idem-fail",
// 	}

// 	err := svc.InternalTransfer(ctx, req)
// 	require.Error(t, err)
// 	require.ErrorIs(t, err, operation.ErrInsufficientFunds)

// 	// balances unchanged
// 	from, _ := queries.GetAccount(ctx, fromAcct.ID)
// 	to, _ := queries.GetAccount(ctx, toAcct.ID)

// 	require.Equal(t, "1000", from.Balance.String())
// 	require.Equal(t, "1000", to.Balance.String())

// 	// no ledger entry
// 	list, _ := queries.ListTransactions(ctx, fromAcct.ID)
// 	require.Len(t, list, 0)

// 	// idempotency key should be deleted
// 	_, ok := idem.Get(ctx, "idem-fail")
// 	require.False(t, ok)
// }
