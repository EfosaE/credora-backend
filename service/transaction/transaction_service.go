package transactionsvc

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/EfosaE/credora-backend/domain/event"
	"github.com/EfosaE/credora-backend/domain/transaction"
	"github.com/EfosaE/credora-backend/internal/utils"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type TransactionService struct {
	trxRepo  transaction.TransactionRepository
	logger   zerolog.Logger
	eventBus event.EventBus
}

func NewTransactionService(
	trxRepo transaction.TransactionRepository,
	logger zerolog.Logger,
	eventBus event.EventBus,
) *TransactionService {

	serviceLogger := logger.With().
		Str("service", "transaction-service").
		Logger()

	return &TransactionService{
		trxRepo:  trxRepo,
		logger:   serviceLogger,
		eventBus: eventBus,
	}
}

func (t *TransactionService) RecordTransaction(
	ctx context.Context,
	req *transaction.NewTransactionInput,
) (*transaction.Transaction, error) {

	log := t.logger.With().
		Str("account_id", req.AccountID.String()).
		Str("reference", req.Reference).
		Str("direction", string(req.Direction)).
		Logger()

	log.Info().Msg("recording transaction")

	trx, err := t.trxRepo.RecordTransaction(ctx, req)
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to record transaction")
		return nil, err
	}

	log.Info().
		Str("transaction_id", strconv.FormatInt(trx.ID, 10)).
		Msg("transaction recorded successfully")

	return trx, nil
}

func (t *TransactionService) GetUserTransactions(
	ctx context.Context,
	userID uuid.UUID,
	cursor *transaction.Cursor,
	limit int32,
) (*[]transaction.Transaction, *transaction.Cursor, error) {

	log := t.logger.With().
		Str("user_id", userID.String()).
		Int32("limit", limit).
		Logger()

	log.Info().Msg("fetching user transactions")

	trx, nextCursor, err := t.trxRepo.GetUserTransactions(ctx, userID, cursor, limit)
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to fetch user transactions")
		return nil, nil, err
	}

	log.Info().
		Int("result_count", len(*trx)).
		Msg("user transactions fetched successfully")

	return trx, nextCursor, nil
}

func (t *TransactionService) SubscribeToInternalTransferCompletedEvents(ctx context.Context) error {

	log := t.logger.With().
		Str("stream", event.StreamTransferEvents).
		Str("consumer_group", "transaction-service-group").
		Logger()

	log.Info().Msg("subscribing to internal transfer success events")

	consumer := utils.WorkerID("transaction")

	return t.eventBus.Subscribe(
		ctx,
		event.StreamTransferEvents,
		"transaction-service-group",
		consumer,
		func(ctx context.Context, msg event.EventMessage) error {

			eventLog := log.With().
				Str("event_type", msg.EventType).
				// Str("message_id", msg.ID).
				Logger()

			if msg.EventType != event.EventInternalTransferSuccess {
				eventLog.Debug().Msg("ignoring unrelated event type")
				return nil
			}

			eventLog.Info().Msg("processing internal transfer success event")

			return t.handleInternalTransferSuccess(ctx, msg)
		},
	)
}

func (t *TransactionService) handleInternalTransferSuccess(
	ctx context.Context,
	msg event.EventMessage,
) error {

	var evt event.InternalTransferEvent
	if err := json.Unmarshal([]byte(msg.Data), &evt); err != nil {
		t.logger.Error().
			Err(err).
			Str("event_type", event.EventInternalTransferSuccess).
			Msg("failed to decode internal transfer success event")
		return fmt.Errorf("failed to decode %s event: %w",
			event.EventInternalTransferSuccess,
			err,
		)
	}

	log := t.logger.With().
		Str("transaction_ref", evt.TransactionRef).
		Str("from_account_id", evt.FromAcctId.String()).
		Str("to_account_id", evt.ToAcctId.String()).
		Str("amount", evt.Amount.String()).
		Logger()

	// Discard stale events older than 3 minutes
	if time.Since(evt.OccurredAt) > 3*time.Minute {
		log.Warn().
			Time("occurred_at", evt.OccurredAt).
			Msg("skipping stale internal transfer event")
		return nil
	}

	log.Info().Msg("recording debit and credit transactions")

	// Debit
	if _, err := t.trxRepo.RecordTransaction(ctx, &transaction.NewTransactionInput{
		AccountID:      evt.FromAcctId,
		CounterpartyID: &evt.ToAcctId,
		Amount:         evt.Amount,
		Direction:      transaction.TransactionTypeDebit,
		Description:    fmt.Sprintf("Internal transfer to %s", evt.ToAcctNum),
		Reference:      evt.TransactionRef,
		Channel:        "INTERNAL_TRANSFER",
		Status:         transaction.StatusSuccess,
	}); err != nil {

		log.Error().
			Err(err).
			Msg("failed to record debit transaction")

		return fmt.Errorf("failed to record debit transaction: %w", err)
	}

	// Credit
	if _, err := t.trxRepo.RecordTransaction(ctx, &transaction.NewTransactionInput{
		AccountID:      evt.ToAcctId,
		CounterpartyID: &evt.FromAcctId,
		Amount:         evt.Amount,
		Direction:      transaction.TransactionTypeCredit,
		Description:    fmt.Sprintf("Internal transfer from %s", evt.FromAcctNum),
		Reference:      evt.TransactionRef,
		Channel:        "INTERNAL_TRANSFER",
		Status:         transaction.StatusSuccess,
	}); err != nil {

		log.Error().
			Err(err).
			Msg("failed to record credit transaction")

		return fmt.Errorf("failed to record credit transaction: %w", err)
	}

	log.Info().Msg("internal transfer transactions recorded successfully")

	return nil
}
