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

func TestInternalTransfer_Integration(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name              string
		transferAmount    string
		expectedFrom      string
		expectedTo        string
		expectError       error
		expectedIdemState string
	}{
		{
			name:              "success transfer",
			transferAmount:    "2000",
			expectedFrom:      "3000.00",
			expectedTo:        "4000.00",
			expectError:       nil,
			expectedIdemState: "SUCCESS",
		},
		{
			name:              "insufficient funds",
			transferAmount:    "7000",
			expectedFrom:      "5000.00",
			expectedTo:        "2000.00",
			expectError:       operation.ErrAccountNotFound, // The sql query is one that returns no rows if there is no account that satisifes the constraints like balance > amount etc
			expectedIdemState: "FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool, cleanup := test.SetupTestDB(t)
			defer cleanup()

			queries := sqlc.New(pool)

			// ---- Seed users ----
			user1, err := queries.CreateUser(ctx, sqlc.CreateUserParams{
				FullName:    "John Doe",
				PhoneNumber: "08011112222",
				Password:    "hashed",
				Nin:         "12345678908",
			})
			require.NoError(t, err)

			user2, err := queries.CreateUser(ctx, sqlc.CreateUserParams{
				FullName:    "Jane Doe",
				PhoneNumber: "08033334444",
				Password:    "hashed",
				Nin:         "12345678956",
			})
			require.NoError(t, err)

			// ---- Seed accounts ----
			fromAcct, err := queries.CreateAccountWithMonnify(ctx, sqlc.CreateAccountWithMonnifyParams{
				UserID:             utils.ToPgUUID(user1.ID),
				Username:           "John Doe",
				AccountNumber:      "1111111111",
				AccountType:        "TESTING ACCOUNT",
				MonnifyCustomerRef: utils.ToPgText("ref-111"),
				VirtualAccountBank: utils.ToPgText("bank-111"),
			})
			require.NoError(t, err)

			toAcct, err := queries.CreateAccountWithMonnify(ctx, sqlc.CreateAccountWithMonnifyParams{
				UserID:             utils.ToPgUUID(user2.ID),
				Username:           "Jane Doe",
				AccountNumber:      "2222222222",
				AccountType:        "TESTING ACCOUNT",
				MonnifyCustomerRef: utils.ToPgText("ref-222"),
				VirtualAccountBank: utils.ToPgText("bank-222"),
			})
			require.NoError(t, err)

			// ---- Seed balances ----
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

			// ---- Build service ----
			acctRepo := infrastructure.NewSqlcAccountRepository(pool)
			txRepo := infrastructure.NewSqlcTransactionRepository(pool)
			idempRepo := infrastructure.NewSqlcIdempotencyRepository(pool)
			txManager := infrastructure.NewTxManager(pool)
			eventBus := &mocks.MockEventBus{}

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
				Amount:         utils.MustPgNumericToDecimal(utils.NewPgNumericFromString(tt.transferAmount)),
				IdempotencyKey: "idem-" + tt.name,
			}

			err = svc.InternalTransfer(ctx, req)

			if tt.expectError != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.expectError)
			} else {
				require.NoError(t, err)
			}

			// ---- Assert balances ----
			updatedFrom, err := queries.GetAccountByAccountNumber(ctx, fromAcct.AccountNumber)
			require.NoError(t, err)
			updatedTo, err := queries.GetAccountByAccountNumber(ctx, toAcct.AccountNumber)
			require.NoError(t, err)

			require.Equal(t, tt.expectedFrom, utils.PgNumericToString(updatedFrom.Balance))
			require.Equal(t, tt.expectedTo, utils.PgNumericToString(updatedTo.Balance))

			// ---- Assert idempotency ----
			idem, err := idempRepo.Get(ctx, req.IdempotencyKey)
			require.NoError(t, err)
			require.Equal(t, tt.expectedIdemState, string(idem.Status))
		})
	}
}
