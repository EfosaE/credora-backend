package transactionsvc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	// "time"

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

func NewTransactionService(trxRepo transaction.TransactionRepository, logger zerolog.Logger, eventBus event.EventBus) *TransactionService {
	return &TransactionService{
		trxRepo:  trxRepo,
		logger:   logger,
		eventBus: eventBus,
	}
}

func (t *TransactionService) RecordTransaction(ctx context.Context, req *transaction.NewTransactionInput) (*transaction.Transaction, error) {
	trx, err := t.trxRepo.RecordTransaction(ctx, req)
	if err != nil {
		return nil, err
	}
	return trx, nil
}

func (t *TransactionService) GetUserTransactions(ctx context.Context, userID uuid.UUID, cursor *transaction.Cursor, limit int32) (*[]transaction.Transaction, *transaction.Cursor, error) {
	trx, nextCursor, err := t.trxRepo.GetUserTransactions(ctx, userID, cursor, limit)
	if err != nil {
		return nil, nil, err
	}
	return trx, nextCursor, nil
}

func (t *TransactionService) SubscribeToInternalTransferCompletedEvents(ctx context.Context) error {
	consumer := utils.WorkerID("transaction")

	return t.eventBus.Subscribe(
		ctx,
		event.StreamTransferEvents,
		"transaction-service-group",
		consumer,
		func(ctx context.Context, msg event.EventMessage) error {

			if msg.EventType != event.EventInternalTransferSuccess {
				return nil
			}

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
		return fmt.Errorf("failed to decode %s event: %w",
			event.EventInternalTransferSuccess,
			err,
		)
	}

	// Discard stale events older than 5 minutes
	if time.Since(evt.OccurredAt) > 3*time.Minute {
		fmt.Printf("Skipping stale event from %s, occurred at %s\n", evt.FromAcctNum, evt.OccurredAt)
		return nil
	}
	// fmt.Println("From Transaction Service-Internal Transfer Success")
	// utils.PrintJSON(evt)

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
		fmt.Println("failed to record debit transaction: %w", err)
		return fmt.Errorf("failed to record debit transaction: %w", err)
	}

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
		fmt.Println("failed to record credit transaction: %w", err)
		return fmt.Errorf("failed to record credit transaction: %w", err)
	}

	return nil
}
