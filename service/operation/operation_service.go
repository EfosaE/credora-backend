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
	"github.com/google/uuid"
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
	// var tRef, recipientName, senderName string
	var tRef string
	var toAcctId, fromAcctId uuid.UUID

	// ---- Insert PROCESSING state OUTSIDE transaction ----
	// t0 := time.Now()
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
			Err(err).
			Str("key", req.IdempotencyKey).
			Msg("Failed to upsert idempotency key PROCESSING")
		return fmt.Errorf("internal error: %w", err)
	}
	// s.logger.Info().Int64("step_ms", time.Since(t0).Milliseconds()).Msg("step: idem_upsert_1")

	// ---- 2 Execute the transfer transaction ----
	err := s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		// t1 := time.Now()

		result, err := s.acctRepo.InternalMoneyTransfer(txCtx, req.Amount, req.FromAcctNum, req.ToAcctNum)
		// s.logger.Info().Int64("step_ms", time.Since(t1).Milliseconds()).Msg("step: make_internal_transfer")
		if err != nil {
			//The worker would parse the err if it is of operation.ErrAccountNotFound and not retry
			return err
		}

		// ---- Record ledger transactions ----
		// The transaction service subscribes to the Transfer Success Event
		tRef = utils.GenerateTransactionReference(result.FromAccountId)
		fromAcctId = result.FromAccountId
		toAcctId = result.ToAccountId
		return nil
	})

	// ---- Persist final idempotency state OUTSIDE transaction ----
	status := transaction.StatusSuccess
	if err != nil {
		s.logger.Warn().
			Err(err).
			Str("key", req.IdempotencyKey).
			Msg("Internal transfer failed, marking idempotency as FAILED")
		status = transaction.StatusFailed
	}

	// t4 := time.Now()
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
			Err(upsertErr).
			Str("key", req.IdempotencyKey).
			Msg("Failed to persist final idempotency state")
		// don't overwrite original error
		if err == nil {
			err = fmt.Errorf("failed to persist idempotency: %w", upsertErr)
		}
	}
	// s.logger.Info().Int64("step_ms", time.Since(t4).Milliseconds()).Msg("step: idem_upsert_2")

	if err == nil {
		evt := event.InternalTransferEvent{
			ToAcctId:   toAcctId,
			FromAcctId: fromAcctId,
			Amount:     req.Amount,
			// RecipientName:  recipientName,
			// SenderName:     senderName,
			FromAcctNum:    req.FromAcctNum,
			ToAcctNum:      req.ToAcctNum,
			OccurredAt:     time.Now().UTC(),
			TransactionRef: tRef,
		}

		payload, err := utils.StructToMap(evt)
		if err != nil {
			return fmt.Errorf("failed to convert typed struct to map: %w", err)
		}

		// t5 := time.Now()
		s.eventBus.Publish(ctx, event.StreamTransferEvents, event.EventInternalTransferSuccess, payload)
		// s.logger.Info().Int64("step_ms", time.Since(t5).Milliseconds()).Msg("step: event_publish")

		s.logger.Info().
			Str("fromAccount", req.FromAcctNum).
			Str("toAccount", req.ToAcctNum).
			Str("amount", req.Amount.String()).
			Str("transactionRef", tRef).
			Msg("Internal transfer successful")
	}

	return err
}

// package operationsvc

// import (
// 	"context"
// 	"fmt"
// 	"time"

// 	"github.com/EfosaE/credora-backend/domain/account"
// 	"github.com/EfosaE/credora-backend/domain/event"
// 	"github.com/EfosaE/credora-backend/domain/idempotency"
// 	"github.com/EfosaE/credora-backend/domain/operation"
// 	"github.com/EfosaE/credora-backend/domain/transaction"
// 	"github.com/EfosaE/credora-backend/infrastructure"
// 	"github.com/EfosaE/credora-backend/internal/utils"
// 	"github.com/google/uuid"
// 	"github.com/rs/zerolog"
// )

// type OperationService struct {
// 	txManager       infrastructure.TxManager
// 	acctRepo        account.AccountRepository
// 	transactionRepo transaction.TransactionRepository
// 	idempRepo       idempotency.IdempotencyRepo
// 	logger          zerolog.Logger
// 	eventBus        event.EventBus
// }

// func NewOperationService(
// 	txManager infrastructure.TxManager,
// 	acctRepo account.AccountRepository,
// 	transactionRepo transaction.TransactionRepository,
// 	idempRepo idempotency.IdempotencyRepo,
// 	logger zerolog.Logger,
// 	event event.EventBus,
// ) *OperationService {
// 	return &OperationService{
// 		txManager:       txManager,
// 		acctRepo:        acctRepo,
// 		transactionRepo: transactionRepo,
// 		idempRepo:       idempRepo,
// 		logger:          logger,
// 		eventBus:        event,
// 	}
// }

// func (s *OperationService) InternalTransfer(
// 	ctx context.Context,
// 	req *operation.InternalTransferReq,
// ) error {
// 	var tRef, recipientName, senderName string
// 	var toAcctId, fromAcctId uuid.UUID

// 	// ---- Insert PROCESSING state OUTSIDE transaction ----
// 	t0 := time.Now()
// 	if err := s.idempRepo.Upsert(
// 		ctx,
// 		req.IdempotencyKey,
// 		operation.OperationTypeInternalTransfer,
// 		map[string]any{
// 			"from":   req.FromAcctNum,
// 			"to":     req.ToAcctNum,
// 			"amount": req.Amount.String(),
// 		},
// 		transaction.StatusProcessing,
// 	); err != nil {
// 		s.logger.Error().
// 			Err(err).
// 			Str("key", req.IdempotencyKey).
// 			Msg("Failed to upsert idempotency key PROCESSING")
// 		return fmt.Errorf("internal error: %w", err)
// 	}
// 	s.logger.Info().Int64("step_ms", time.Since(t0).Milliseconds()).Msg("step: idem_upsert_1")

// 	// ---- 3️⃣ Execute the transfer transaction ----
// 	err := s.txManager.WithTx(ctx, func(txCtx context.Context) error {
// 		from := req.FromAcctNum
// 		to := req.ToAcctNum

// 		// ---- Lock accounts in deterministic order to prevent deadlocks ----
// 		first, second := from, to
// 		if to < from {
// 			first, second = to, from
// 		}

// 		t1 := time.Now()
// 		accounts, err := s.acctRepo.GetAccountsForUpdate(txCtx, []string{first, second})
// 		s.logger.Info().Int64("step_ms", time.Since(t1).Milliseconds()).Msg("step: get_accounts")
// 		if err != nil {
// 			return fmt.Errorf("failed to lock accounts: %w", err)
// 		}

// 		accountMap := make(map[string]*account.Account)
// 		for _, acc := range accounts {
// 			accountMap[acc.AccountNumber] = acc
// 		}

// 		fromAcct, ok := accountMap[from]
// 		if !ok {
// 			return fmt.Errorf("from account %s not found", from)
// 		}
// 		senderName = fromAcct.UserName

// 		toAcct, ok := accountMap[to]
// 		if !ok {
// 			return fmt.Errorf("to account %s not found", to)
// 		}
// 		recipientName = toAcct.UserName

// 		// ---- Validate balance ----
// 		if fromAcct.Balance.LessThan(req.Amount) {
// 			return operation.ErrInsufficientFunds
// 		}

// 		// ---- Perform debit & credit ----
// 		t2 := time.Now()
// 		if _, err := s.acctRepo.DebitAccount(txCtx, req.Amount, from); err != nil {
// 			return fmt.Errorf("failed to debit account: %w", err)
// 		}
// 		s.logger.Info().Int64("step_ms", time.Since(t2).Milliseconds()).Msg("step: debit")

// 		t3 := time.Now()
// 		if _, err := s.acctRepo.CreditAccount(txCtx, req.Amount, to); err != nil {
// 			return fmt.Errorf("failed to credit account: %w", err)
// 		}
// 		s.logger.Info().Int64("step_ms", time.Since(t3).Milliseconds()).Msg("step: credit")

// 		// ---- Record ledger transactions ----
// 		// The transaction service subscribes to the Transfer Success Event
// 		tRef = utils.GenerateTransactionReference(fromAcct.ID)
// 		toAcctId = toAcct.ID
// 		fromAcctId = fromAcct.ID

// 		return nil
// 	})

// 	// ---- 4️⃣ Persist final idempotency state OUTSIDE transaction ----
// 	status := transaction.StatusSuccess
// 	if err != nil {
// 		s.logger.Warn().
// 			Err(err).
// 			Str("key", req.IdempotencyKey).
// 			Msg("Internal transfer failed, marking idempotency as FAILED")
// 		status = transaction.StatusFailed
// 	}

// 	t4 := time.Now()
// 	if upsertErr := s.idempRepo.Upsert(
// 		ctx,
// 		req.IdempotencyKey,
// 		operation.OperationTypeInternalTransfer,
// 		map[string]any{
// 			"from":   req.FromAcctNum,
// 			"to":     req.ToAcctNum,
// 			"amount": req.Amount.String(),
// 		},
// 		status,
// 	); upsertErr != nil {
// 		s.logger.Error().
// 			Err(upsertErr).
// 			Str("key", req.IdempotencyKey).
// 			Msg("Failed to persist final idempotency state")
// 		// don't overwrite original error
// 		if err == nil {
// 			err = fmt.Errorf("failed to persist idempotency: %w", upsertErr)
// 		}
// 	}
// 	s.logger.Info().Int64("step_ms", time.Since(t4).Milliseconds()).Msg("step: idem_upsert_2")

// 	if err == nil {
// 		evt := event.InternalTransferEvent{
// 			ToAcctId:       toAcctId,
// 			FromAcctId:     fromAcctId,
// 			Amount:         req.Amount,
// 			RecipientName:  recipientName,
// 			SenderName:     senderName,
// 			FromAcctNum:    req.FromAcctNum,
// 			ToAcctNum:      req.ToAcctNum,
// 			OccurredAt:     time.Now().UTC(),
// 			TransactionRef: tRef,
// 		}

// 		payload, err := utils.StructToMap(evt)
// 		if err != nil {
// 			return fmt.Errorf("failed to convert typed struct to map: %w", err)
// 		}

// 		t5 := time.Now()
// 		s.eventBus.Publish(ctx, event.StreamTransferEvents, event.EventInternalTransferSuccess, payload)
// 		s.logger.Info().Int64("step_ms", time.Since(t5).Milliseconds()).Msg("step: event_publish")

// 		s.logger.Info().
// 			Str("fromAccount", req.FromAcctNum).
// 			Str("toAccount", req.ToAcctNum).
// 			Str("amount", req.Amount.String()).
// 			Str("transactionRef", tRef).
// 			Msg("Internal transfer successful")
// 	}

// 	return err
// }
