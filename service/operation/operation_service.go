package operationsvc

import (
	"context"
	"fmt"

	"github.com/EfosaE/credora-backend/domain/account"
	"github.com/EfosaE/credora-backend/domain/idempotency"
	"github.com/EfosaE/credora-backend/domain/logger"
	"github.com/EfosaE/credora-backend/domain/operation"
	"github.com/EfosaE/credora-backend/domain/transaction"
	"github.com/EfosaE/credora-backend/infrastructure"
	"github.com/EfosaE/credora-backend/internal/utils"
)

type OperationService struct {
	txManager       infrastructure.TxManager
	acctRepo        account.AccountRepository
	transactionRepo transaction.TransactionRepository
	idempRepo       idempotency.IdempotencyRepo
	logger          *logger.Logger
}

func NewOperationService(
	txManager infrastructure.TxManager,
	acctRepo account.AccountRepository,
	transactionRepo transaction.TransactionRepository,
	idempRepo idempotency.IdempotencyRepo,
	logger *logger.Logger,
) *OperationService {
	return &OperationService{
		txManager:       txManager,
		acctRepo:        acctRepo,
		transactionRepo: transactionRepo,
		idempRepo:       idempRepo,
		logger:          logger,
	}
}

// // InternalTransfer performs debit + credit + ledger atomically using DB-backed idempotency
// func (s *OperationService) InternalTransfer(ctx context.Context, req *operation.InternalTransferReq) error {

// 	// ---- 1️⃣ Wrap everything in a DB transaction ----
// 	return s.acctRepo.WithTx(ctx, func(accTx account.AccountTx) error {
// 		sqlTx := accTx.Tx()

// 		txTransactionRepo := s.transactionRepo.WithTx(sqlTx)
// 		txIdempRepo := s.idempRepo.WithTx(sqlTx)

// 		from := req.FromAcctNum
// 		to := req.ToAcctNum

// 		// ---- 2️⃣ Lock accounts in order (deadlock prevention) ----
// 		first, second := from, to
// 		if to < from {
// 			first, second = to, from
// 		}

// 		// Lock both accounts at once to reduce round trips
// 		accounts, err := accTx.GetAccountsForUpdate(ctx, []string{first, second})
// 		if err != nil {
// 			return fmt.Errorf("failed to lock accounts: %w", err)
// 		}

// 		// Map accounts by account number
// 		accountMap := make(map[string]*account.Account)
// 		for _, acc := range accounts {
// 			accountMap[acc.AccountNumber] = acc
// 		}

// 		fromAcct, ok := accountMap[from]
// 		if !ok {
// 			return fmt.Errorf("from account %s not found", from)
// 		}

// 		toAcct, ok := accountMap[to]
// 		if !ok {
// 			return fmt.Errorf("to account %s not found", to)
// 		}

// 		// ---- 3️⃣ Validate balance ----
// 		if fromAcct.Balance.LessThan(req.Amount) {
// 			// outside tx, but SYNC
// 			err = s.idempRepo.Upsert(ctx, req.IdempotencyKey, operation.OperationTypeInternalTransfer, map[string]any{
// 				"error": "insufficient_funds",
// 			}, transaction.StatusFailed)

// 			if err != nil {
// 				s.logger.Error("Failed to upsert idempotency after insufficient funds", map[string]any{
// 					"error": err.Error(),
// 				})
// 			}

// 			return operation.ErrInsufficientFunds
// 		}

// 		// ---- 4️⃣ Perform debit & credit ----
// 		if _, err := accTx.DebitAccount(ctx, req.Amount, from); err != nil {
// 			return err
// 		}
// 		if _, err := accTx.CreditAccount(ctx, req.Amount, to); err != nil {
// 			return err
// 		}

// 		// ---- 5️⃣ Record ledger transactions ----

// 		// Use ONE shared reference for both sides
// 		ref := utils.GenerateTransactionReference(fromAcct.ID)

// 		// Debit entry (from account)
// 		if _, err := txTransactionRepo.RecordTransaction(ctx, &transaction.NewTransactionInput{
// 			AccountID:      fromAcct.ID,
// 			CounterpartyID: &toAcct.ID,
// 			Amount:         req.Amount,
// 			Direction:      transaction.TransactionTypeDebit,
// 			Description:    fmt.Sprintf("Internal transfer to %s", to),
// 			Reference:      ref,
// 			Channel:        "INTERNAL_TRANSFER",
// 			Status:         transaction.StatusSuccess,
// 		}); err != nil {
// 			return fmt.Errorf("failed to record debit transaction: %w", err)
// 		}

// 		// Credit entry (to account)
// 		if _, err = txTransactionRepo.RecordTransaction(ctx, &transaction.NewTransactionInput{
// 			AccountID:      toAcct.ID,
// 			CounterpartyID: &fromAcct.ID,
// 			Amount:         req.Amount,
// 			Direction:      transaction.TransactionTypeCredit,
// 			Description:    fmt.Sprintf("Internal transfer from %s", from),
// 			Reference:      ref,
// 			Channel:        "INTERNAL_TRANSFER",
// 			Status:         transaction.StatusSuccess,
// 		}); err != nil {
// 			return fmt.Errorf("failed to record credit transaction: %w", err)
// 		}

// 		// ---- 6️⃣ Upsert idempotency as SUCCESS ----
// 		if err := txIdempRepo.Upsert(ctx, req.IdempotencyKey, operation.OperationTypeInternalTransfer, map[string]any{
// 			"from":   from,
// 			"to":     to,
// 			"amount": req.Amount.String(),
// 		}, transaction.StatusSuccess); err != nil {
// 			return fmt.Errorf("idempotency upsert failed: %w", err)
// 		}

// 		// ---- 7️⃣ Log success ----
// 		s.logger.Info("Internal transfer successful", map[string]any{
// 			"from_account": from,
// 			"to_account":   to,
// 			"amount":       req.Amount,
// 			"from_balance": fromAcct.Balance.Sub(req.Amount),
// 		})

// 		return nil
// 	})

// }

func (s *OperationService) InternalTransfer(
	ctx context.Context,
	req *operation.InternalTransferReq,
) error {

	// ---- Insert PROCESSING state OUTSIDE transaction ----
	if err := s.idempRepo.Upsert(
		ctx,
		req.IdempotencyKey,
		operation.OperationTypeInternalTransfer,
		map[string]any{
			"from":   req.FromAcctNum,
			"to":     req.ToAcctNum,
			"amount": req.Amount.String(),
		},
		transaction.StatusProcessing,
	); err != nil {
		s.logger.Error("Failed to upsert idempotency key PROCESSING", map[string]any{
			"key":   req.IdempotencyKey,
			"error": err.Error(),
		})
		return fmt.Errorf("internal error: %w", err)
	}

	// ---- 3️⃣ Execute the transfer transaction ----
	err := s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		from := req.FromAcctNum
		to := req.ToAcctNum

		// ---- Lock accounts in deterministic order to prevent deadlocks ----
		first, second := from, to
		if to < from {
			first, second = to, from
		}

		accounts, err := s.acctRepo.GetAccountsForUpdate(txCtx, []string{first, second})
		if err != nil {
			return fmt.Errorf("failed to lock accounts: %w", err)
		}

		accountMap := make(map[string]*account.Account)
		for _, acc := range accounts {
			accountMap[acc.AccountNumber] = acc
		}

		fromAcct, ok := accountMap[from]
		if !ok {
			return fmt.Errorf("from account %s not found", from)
		}

		toAcct, ok := accountMap[to]
		if !ok {
			return fmt.Errorf("to account %s not found", to)
		}

		// ---- Validate balance ----
		if fromAcct.Balance.LessThan(req.Amount) {
			return operation.ErrInsufficientFunds
		}

		// ---- Perform debit & credit ----
		if _, err := s.acctRepo.DebitAccount(txCtx, req.Amount, from); err != nil {
			return fmt.Errorf("failed to debit account: %w", err)
		}

		if _, err := s.acctRepo.CreditAccount(txCtx, req.Amount, to); err != nil {
			return fmt.Errorf("failed to credit account: %w", err)
		}

		// ---- Record ledger transactions ----
		ref := utils.GenerateTransactionReference(fromAcct.ID)

		if _, err := s.transactionRepo.RecordTransaction(txCtx, &transaction.NewTransactionInput{
			AccountID:      fromAcct.ID,
			CounterpartyID: &toAcct.ID,
			Amount:         req.Amount,
			Direction:      transaction.TransactionTypeDebit,
			Description:    fmt.Sprintf("Internal transfer to %s", to),
			Reference:      ref,
			Channel:        "INTERNAL_TRANSFER",
			Status:         transaction.StatusSuccess,
		}); err != nil {
			return fmt.Errorf("failed to record debit transaction: %w", err)
		}

		if _, err := s.transactionRepo.RecordTransaction(txCtx, &transaction.NewTransactionInput{
			AccountID:      toAcct.ID,
			CounterpartyID: &fromAcct.ID,
			Amount:         req.Amount,
			Direction:      transaction.TransactionTypeCredit,
			Description:    fmt.Sprintf("Internal transfer from %s", from),
			Reference:      ref,
			Channel:        "INTERNAL_TRANSFER",
			Status:         transaction.StatusSuccess,
		}); err != nil {
			return fmt.Errorf("failed to record credit transaction: %w", err)
		}

		return nil
	})

	// ---- 4️⃣ Persist final idempotency state OUTSIDE transaction ----
	status := transaction.StatusSuccess
	if err != nil {
		s.logger.Warn("Internal transfer failed, marking idempotency as FAILED", map[string]any{
			"key":   req.IdempotencyKey,
			"error": err.Error(),
		})
		status = transaction.StatusFailed
	}

	if upsertErr := s.idempRepo.Upsert(
		ctx,
		req.IdempotencyKey,
		operation.OperationTypeInternalTransfer,
		map[string]any{
			"from":   req.FromAcctNum,
			"to":     req.ToAcctNum,
			"amount": req.Amount.String(),
		},
		status,
	); upsertErr != nil {
		s.logger.Error("Failed to persist final idempotency state", map[string]any{
			"key":   req.IdempotencyKey,
			"error": upsertErr.Error(),
		})
		// don't overwrite original error
		if err == nil {
			err = fmt.Errorf("failed to persist idempotency: %w", upsertErr)
		}
	}

	if err == nil {
		s.logger.Info("Internal transfer successful", map[string]any{
			"from_account": req.FromAcctNum,
			"to_account":   req.ToAcctNum,
			"amount":       req.Amount,
		})
	}

	return err
}
