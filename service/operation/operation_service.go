package operationsvc

import (
	"context"
	"fmt"

	"github.com/EfosaE/credora-backend/domain/account"
	"github.com/EfosaE/credora-backend/domain/logger"
	"github.com/EfosaE/credora-backend/domain/operation"
	"github.com/EfosaE/credora-backend/domain/transaction"
	"github.com/EfosaE/credora-backend/infrastructure"
	"github.com/EfosaE/credora-backend/internal/utils"
)

type OperationService struct {
	acctRepo        account.AccountRepository
	transactionRepo transaction.TransactionRepository
	logger          *logger.Logger
	idempotency     *infrastructure.IdempotencyCache
}

func NewOperationService(
	acctRepo account.AccountRepository,
	transactionRepo transaction.TransactionRepository,
	idempotency *infrastructure.IdempotencyCache,
	logger *logger.Logger,
) *OperationService {
	return &OperationService{
		acctRepo:        acctRepo,
		transactionRepo: transactionRepo,
		logger:          logger,
		idempotency:     idempotency,
	}
}

// InternalTransfer performs debit + credit + ledger atomically
// InternalTransfer performs debit + credit + ledger atomically
func (s *OperationService) InternalTransfer(ctx context.Context, req *operation.InternalTransferReq) error {

	return s.acctRepo.WithTx(ctx, func(accTx account.AccountTx) error {

		// Bind transaction repo to the same TX
		txTransactionRepo := s.transactionRepo.WithTx(accTx.(*infrastructure.SqlcAccountRepository).Tx())

		// 1. Lock both accounts for update (order by account number to avoid deadlocks)
		//    (choose deterministic order to avoid deadlocks across concurrent transfers)
		fromNum := req.FromAcctNum
		toNum := req.ToAcctNum

		lockFirst := fromNum
		lockSecond := toNum
		if toNum < fromNum {
			lockFirst = toNum
			lockSecond = fromNum
		}
		// lock first account
		if _, err := accTx.GetAccountForUpdate(ctx, lockFirst); err != nil {
			return fmt.Errorf("locking account %s: %w", lockFirst, err)
		}
		// lock second account
		if _, err := accTx.GetAccountForUpdate(ctx, lockSecond); err != nil {
			return fmt.Errorf("locking account %s: %w", lockSecond, err)
		}
		// Re-fetch canonical fromAcct for balance check (safe because we locked)
		fromAcct, err := accTx.GetAccountForUpdate(ctx, fromNum)
		if err != nil {
			return fmt.Errorf("fetching debited account: %w", err)
		}
		toAcct, err := accTx.GetAccountForUpdate(ctx, toNum)
		if err != nil {
			return fmt.Errorf("fetching credited account: %w", err)
		}
		// 2️⃣ Check funds
		if fromAcct.Balance.LessThan(req.Amount) {
			s.logger.Warn("Insufficient funds for transfer", map[string]any{
				"from_account": req.FromAcctNum,
				"to_account":   req.ToAcctNum,
				"amount":       req.Amount,
				"balance":      fromAcct.Balance,
			})

			// Delete the idempotency key since the operation failed
			s.idempotency.Delete(ctx, req.IdempotencyKey)
			return operation.ErrInsufficientFunds
		}

		// 3️⃣ Debit sender
		if _, err := accTx.DebitAccount(ctx, req.Amount, req.FromAcctNum); err != nil {
			return err
		}

		// 4️⃣ Credit receiver
		if _, err := accTx.CreditAccount(ctx, req.Amount, req.ToAcctNum); err != nil {
			return err
		}

		// 5️⃣ Record ledger transaction inside same TX
		_, err = txTransactionRepo.RecordTransaction(ctx, &transaction.NewTransactionInput{
			AccountID:      fromAcct.ID,
			CounterpartyID: &toAcct.ID, //This takes a pointer to allow nulls
			Amount:         req.Amount,
			Description:    fmt.Sprintf("Internal transfer to %s", req.ToAcctNum),
			Reference:      utils.GenerateTransactionReference(fromAcct.ID),
			Channel:        "INTERNAL_TRANSFER",
			Status:         transaction.StatusSuccess,
		})
		if err != nil {
			return err
		}

		s.idempotency.MarkDone(ctx, req.IdempotencyKey, "SUCCESS")
		// ✅ Log success
		s.logger.Info("Internal transfer successful", map[string]any{
			"from_account": req.FromAcctNum,
			"to_account":   req.ToAcctNum,
			"amount":       req.Amount,
			"from_balance": fromAcct.Balance.Sub(req.Amount), // after debit
		})

		return nil
	})
}
