package operationsvc

import (
	"context"
	"fmt"
	"time"

	"github.com/EfosaE/credora-backend/domain/account"
	"github.com/EfosaE/credora-backend/domain/event"
	"github.com/EfosaE/credora-backend/domain/idempotency"
	"github.com/EfosaE/credora-backend/domain/operation"
	"github.com/EfosaE/credora-backend/domain/transaction"
	"github.com/EfosaE/credora-backend/infrastructure"
	"github.com/EfosaE/credora-backend/internal/utils"
	"github.com/rs/zerolog"
)

type OperationService struct {
	txManager       infrastructure.TxManager
	acctRepo        account.AccountRepository
	transactionRepo transaction.TransactionRepository
	idempRepo       idempotency.IdempotencyRepo
	logger          zerolog.Logger
	eventBus        event.EventBus
}

func NewOperationService(
	txManager infrastructure.TxManager,
	acctRepo account.AccountRepository,
	transactionRepo transaction.TransactionRepository,
	idempRepo idempotency.IdempotencyRepo,
	logger zerolog.Logger,
	event event.EventBus,
) *OperationService {
	return &OperationService{
		txManager:       txManager,
		acctRepo:        acctRepo,
		transactionRepo: transactionRepo,
		idempRepo:       idempRepo,
		logger:          logger,
		eventBus:        event,
	}
}

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
		s.logger.Error().
			Str("msg", "Failed to upsert idempotency key PROCESSING").
			Str("key", req.IdempotencyKey).
			Str("error", err.Error()).
			Msg("")
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
		s.logger.Warn().
			Str("msg", "Internal transfer failed, marking idempotency as FAILED").
			Str("key", req.IdempotencyKey).
			Str("error", err.Error()).
			Msg("")
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
		s.logger.Error().
			Str("msg", "Failed to persist final idempotency state").
			Str("key", req.IdempotencyKey).
			Str("error", upsertErr.Error()).
			Msg("")
		// don't overwrite original error
		if err == nil {
			err = fmt.Errorf("failed to persist idempotency: %w", upsertErr)
		}
	}

	if err == nil {
		evt := event.MoneyTransferredEvent{
			Amount:      req.Amount,
			TransferID:  req.IdempotencyKey,
			FromAcctNum: req.FromAcctNum,
			ToAcctNum:   req.ToAcctNum,
			OccurredAt:  time.Now().UTC(),
		}

		payload, err := utils.StructToMap(evt)
		if err != nil {
			return fmt.Errorf("failed to convert typed struct to map: %w", err)
		}
		s.eventBus.Publish(ctx, event.StreamTransferEvents, event.EventInternalTransferSuccess, payload)
		s.logger.Info().
			Str("msg", "Internal transfer successful").
			Str("from_account", req.FromAcctNum).
			Str("toAccount", req.ToAcctNum).
			Str("amount", req.Amount.String()).
			Msg("")
	}

	return err
}
