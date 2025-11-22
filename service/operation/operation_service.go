package operationsvc

import (
	"context"
	"fmt"

	"github.com/EfosaE/credora-backend/domain/account"
	"github.com/EfosaE/credora-backend/domain/idempotency"
	"github.com/EfosaE/credora-backend/domain/logger"
	"github.com/EfosaE/credora-backend/domain/operation"
	"github.com/EfosaE/credora-backend/domain/transaction"
	"github.com/EfosaE/credora-backend/internal/utils"
)

type OperationService struct {
	acctRepo        account.AccountRepository
	transactionRepo transaction.TransactionRepository
	idempTable      idempotency.IdempotencyRepo
	logger          *logger.Logger
}

func NewOperationService(
	acctRepo account.AccountRepository,
	transactionRepo transaction.TransactionRepository,
	idempTable idempotency.IdempotencyRepo,
	logger *logger.Logger,
) *OperationService {
	return &OperationService{
		acctRepo:        acctRepo,
		transactionRepo: transactionRepo,
		idempTable:      idempTable,
		logger:          logger,
	}
}

// InternalTransfer performs debit + credit + ledger atomically using DB-backed idempotency
func (s *OperationService) InternalTransfer(ctx context.Context, req *operation.InternalTransferReq) error {

	// ---- 1️⃣ Wrap everything in a DB transaction ----
	return s.acctRepo.WithTx(ctx, func(accTx account.AccountTx) error {
		sqlTx := accTx.Tx()

		txTransactionRepo := s.transactionRepo.WithTx(sqlTx)
		txIdempTable := s.idempTable.WithTx(sqlTx)

		from := req.FromAcctNum
		to := req.ToAcctNum

		// ---- 2️⃣ Lock accounts in order (deadlock prevention) ----
		first, second := from, to
		if to < from {
			first, second = to, from
		}
		if _, err := accTx.GetAccountForUpdate(ctx, first); err != nil {
			return fmt.Errorf("lock account %s: %w", first, err)
		}
		if _, err := accTx.GetAccountForUpdate(ctx, second); err != nil {
			return fmt.Errorf("lock account %s: %w", second, err)
		}

		fromAcct, err := accTx.GetAccountForUpdate(ctx, from)
		if err != nil {
			return fmt.Errorf("get from account: %w", err)
		}
		toAcct, err := accTx.GetAccountForUpdate(ctx, to)
		if err != nil {
			return fmt.Errorf("get to account: %w", err)
		}

		// ---- 3️⃣ Validate balance ----
		if fromAcct.Balance.LessThan(req.Amount) {
			// outside tx, but SYNC
			err = s.idempTable.Upsert(ctx, req.IdempotencyKey, operation.OperationTypeInternalTransfer, map[string]any{
				"error": "insufficient_funds",
			}, transaction.StatusFailed)

			if err != nil {
				s.logger.Error("Failed to upsert idempotency after insufficient funds", map[string]any{
					"error": err.Error(),
				})
			}

			return operation.ErrInsufficientFunds
		}

		// ---- 4️⃣ Perform debit & credit ----
		if _, err := accTx.DebitAccount(ctx, req.Amount, from); err != nil {
			return err
		}
		if _, err := accTx.CreditAccount(ctx, req.Amount, to); err != nil {
			return err
		}

		// ---- 5️⃣ Record ledger transaction ----
		_, err = txTransactionRepo.RecordTransaction(ctx, &transaction.NewTransactionInput{
			AccountID:      fromAcct.ID,
			CounterpartyID: &toAcct.ID,
			Amount:         req.Amount,
			Description:    fmt.Sprintf("Internal transfer to %s", to),
			Reference:      utils.GenerateTransactionReference(fromAcct.ID),
			Channel:        "INTERNAL_TRANSFER",
			Status:         transaction.StatusSuccess,
		})
		if err != nil {
			return err
		}

		// ---- 6️⃣ Upsert idempotency as SUCCESS ----
		if err := txIdempTable.Upsert(ctx, req.IdempotencyKey, operation.OperationTypeInternalTransfer, map[string]any{
			"from":   from,
			"to":     to,
			"amount": req.Amount.String(),
		}, transaction.StatusSuccess); err != nil {
			return fmt.Errorf("idempotency upsert failed: %w", err)
		}

		// ---- 7️⃣ Log success ----
		s.logger.Info("Internal transfer successful", map[string]any{
			"from_account": from,
			"to_account":   to,
			"amount":       req.Amount,
			"from_balance": fromAcct.Balance.Sub(req.Amount),
		})

		return nil
	})

}
