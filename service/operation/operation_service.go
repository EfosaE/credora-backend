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
	"github.com/EfosaE/credora-backend/domain/txmanager"
	"github.com/EfosaE/credora-backend/internal/utils"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type OperationService struct {
	txManager       txmanager.TxManager
	acctRepo        account.AccountRepository
	transactionRepo transaction.TransactionRepository
	idempRepo       idempotency.IdempotencyRepo
	logger          zerolog.Logger
	eventBus        event.EventBus
}

func NewOperationService(
	txManager txmanager.TxManager,
	acctRepo account.AccountRepository,
	transactionRepo transaction.TransactionRepository,
	idempRepo idempotency.IdempotencyRepo,
	logger zerolog.Logger,
	event event.EventBus,
) *OperationService {

	serviceLogger := logger.With().
		Str("service", "operation-service").
		Logger()

	return &OperationService{
		txManager:       txManager,
		acctRepo:        acctRepo,
		transactionRepo: transactionRepo,
		idempRepo:       idempRepo,
		logger:          serviceLogger,
		eventBus:        event,
	}
}

func (s *OperationService) InternalTransfer(
	ctx context.Context,
	req *operation.InternalTransferReq,
) error {

	var tRef string
	var toAcctId, fromAcctId, toUserId, fromUserId uuid.UUID

	logCtx := s.logger.With().
		Str("idempotency_key", req.IdempotencyKey).
		Str("from_account", req.FromAcctNum).
		Str("to_account", req.ToAcctNum).
		Str("amount", req.Amount.String()).
		Logger()

	logCtx.Info().Msg("initiating internal transfer")

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
		logCtx.Error().
			Err(err).
			Msg("failed to upsert idempotency key as PROCESSING")
		return fmt.Errorf("internal error: %w", err)
	}

	// ---- 2 Execute the transfer transaction ----
	err := s.txManager.WithTx(ctx, func(txCtx context.Context) error {

		result, err := s.acctRepo.InternalMoneyTransfer(txCtx, req.Amount, req.FromAcctNum, req.ToAcctNum)
		if err != nil {
			logCtx.Error().
				Err(err).
				Msg("failed to execute internal money transfer")
			return err
		}

		tRef = utils.GenerateTransactionReference(result.FromAccountId)
		fromAcctId = result.FromAccountId
		toAcctId = result.ToAccountId
		toUserId = result.ToUserId
		fromUserId = result.FromUserId

		logCtx.Info().
			Str("transaction_ref", tRef).
			Msg("internal transfer executed successfully")

		return nil
	})

	// ---- Persist final idempotency state OUTSIDE transaction ----
	status := transaction.StatusSuccess
	if err != nil {
		logCtx.Warn().
			Err(err).
			Msg("internal transfer failed, marking idempotency as FAILED")
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
		logCtx.Error().
			Err(upsertErr).
			Msg("failed to persist final idempotency state")
		if err == nil {
			err = fmt.Errorf("failed to persist idempotency: %w", upsertErr)
		}
	}

	// ---- Publish success event if no error ----
	if err == nil {
		evt := event.InternalTransferEvent{
			ToAcctId:       toAcctId,
			FromAcctId:     fromAcctId,
			Amount:         req.Amount,
			ToAcctUserId:   toUserId,
			FromAcctNum:    req.FromAcctNum,
			ToAcctNum:      req.ToAcctNum,
			FromAcctUserId: fromUserId,
			OccurredAt:     time.Now().UTC(),
			TransactionRef: tRef,
		}

		payload, err := utils.StructToMap(evt)
		if err != nil {
			logCtx.Error().
				Err(err).
				Msg("failed to convert internal transfer event to payload")
			return fmt.Errorf("failed to convert typed struct to map: %w", err)
		}

		s.eventBus.Publish(ctx, event.StreamTransferEvents, event.EventInternalTransferSuccess, payload)

		logCtx.Info().
			Str("transaction_ref", tRef).
			Msg("internal transfer completed successfully and event published")
	}

	return err
}
