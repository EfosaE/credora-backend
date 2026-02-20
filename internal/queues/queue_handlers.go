package queues

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/EfosaE/credora-backend/domain/operation"
	"github.com/EfosaE/credora-backend/domain/transaction"
	"github.com/EfosaE/credora-backend/domain/webhook"
	"github.com/EfosaE/credora-backend/service"
	accountsvc "github.com/EfosaE/credora-backend/service/account"
	operationsvc "github.com/EfosaE/credora-backend/service/operation"
	transactionsvc "github.com/EfosaE/credora-backend/service/transaction"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
)

type Handlers struct {
	EmailSvc     service.EmailService
	OperationSvc operationsvc.OperationService
	AcctSvc      *accountsvc.AccountService
	TrxSvc       *transactionsvc.TransactionService
	logger       zerolog.Logger
}

func NewHandlers(es service.EmailService, ops operationsvc.OperationService, acctSvc *accountsvc.AccountService, trxSvc *transactionsvc.TransactionService, logger zerolog.Logger) *Handlers {
	return &Handlers{es, ops, acctSvc, trxSvc, logger}
}

// EMAIL NOTIFICATION
func (h *Handlers) HandleSendWelcomeEmail(ctx context.Context, t *asynq.Task) error {
	var p WelcomeEmailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		h.logger.Error().
			Err(err).
			Str("task", "HandleSendWelcomeEmail").
			Msg("failed to unmarshal WelcomeEmailPayload")

		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	h.logger.Info().
		Str("task", "HandleSendWelcomeEmail").
		Str("user_id", p.User.ID.String()).
		Str("template_id", p.TemplateID).
		Msg("sending welcome email")

	if err := h.EmailSvc.SendWelcomeEmail(ctx, *p.User); err != nil {
		h.logger.Error().
			Err(err).
			Str("task", "HandleSendWelcomeEmail").
			Str("user_id", p.User.ID.String()).
			Msg("failed to send welcome email")
		return err
	}

	return nil
}

func (h *Handlers) HandleSendAccountNumberEmail(ctx context.Context, t *asynq.Task) error {
	var p AccountNumberEmailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		h.logger.Error().
			Err(err).
			Str("task", "HandleSendAccountNumberEmail").
			Msg("failed to unmarshal AccountNumberEmailPayload")
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	h.logger.Info().
		Str("task", "HandleSendAccountNumberEmail").
		Str("to", p.To).
		Str("bank", p.Bank).
		Str("accountNumber", p.AccountNumber).
		Msg("sending account number email")

	if err := h.EmailSvc.SendAccountNumberEmail(ctx, p.To, p.Bank, p.AccountNumber); err != nil {
		h.logger.Error().
			Err(err).
			Str("task", "HandleSendAccountNumberEmail").
			Str("to", p.To).
			Msg("failed to send account number email")
		return err
	}

	return nil
}

// TRANSFER HANDLER
func (h *Handlers) HandleInternalTransfer(ctx context.Context, t *asynq.Task) error {
	var payload InternalTransferTaskPayload

	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		h.logger.Error().
			Err(err).
			Str("task", "HandleInternalTransfer").
			Msg("failed to unmarshal InternalTransferReq")
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	startedAt := time.Now()
	queuedAt := time.Unix(0, payload.QueuedAt)

	queueWait := startedAt.Sub(queuedAt)

	p := payload.Req

	h.logger.Info().
		Str("task", "HandleInternalTransfer").
		Str("from", p.FromAcctNum).
		Str("to", p.ToAcctNum).
		Str("amount", p.Amount.String()).
		Str("queue_wait", queueWait.String()).
		Msg("processing internal transfer")

	err := h.OperationSvc.InternalTransfer(ctx, p)
	finishedAt := time.Now()
	processingTime := finishedAt.Sub(startedAt)
	totalTime := finishedAt.Sub(queuedAt)
	h.logger.Info().
		Str("task", "HandleInternalTransfer").
		Str("from", p.FromAcctNum).
		Str("to", p.ToAcctNum).
		Int64("queue_wait_ms", queueWait.Milliseconds()).
		Int64("processing_ms", processingTime.Milliseconds()).
		Int64("total_time_ms", totalTime.Milliseconds()).
		Msg("internal transfer timing")

	if err != nil {
		switch {
		case errors.Is(err, operation.ErrInsufficientFunds):
			h.logger.Info().
				Str("task", "HandleInternalTransfer").
				Str("from", p.FromAcctNum).
				Msg("transfer failed - insufficient funds")
			return nil

		case errors.Is(err, operation.ErrAccountNotFound):
			h.logger.Info().
				Str("task", "HandleInternalTransfer").
				Str("from", p.FromAcctNum).
				Str("to", p.ToAcctNum).
				Msg("transfer failed - account not found")
			return nil
		// like lock errors
		default:
			h.logger.Error().
				Err(err).
				Str("task", "HandleInternalTransfer").
				Str("from", p.FromAcctNum).
				Str("to", p.ToAcctNum).
				Msg("internal transfer failed (retrying)")
			return err
		}
	}

	// h.logInfo("internal transfer completed successfully", map[string]any{
	// 	"from":   p.FromAcctNum,
	// 	"to":     p.ToAcctNum,
	// 	"amount": p.Amount.String(),
	// })

	return nil
}

func (h *Handlers) HandleInboundTransferWebhook(ctx context.Context, t *asynq.Task) error {
	var p webhook.InboundTransferPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		h.logger.Error().
			Err(err).
			Str("task", "HandleInboundTransferWebhook").
			Msg("failed to unmarshal InboundTransferPayload")
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	amount, _ := decimal.NewFromString(p.TransactionDetails.SettlementAmount)
	h.logger.Info().
		Str("task", "HandleInboundTransferWebhook").
		Str("amount", amount.String()).
		Str("destination", p.TransactionDetails.DestinationAccountInfo.AccountNumber).
		Str("reference", p.TransactionDetails.PaymentReference).
		Msg("processing inbound transfer webhook")

	// Step 1: Credit account
	result, err := h.AcctSvc.CreditUserBalance(ctx, amount,
		p.TransactionDetails.DestinationAccountInfo.AccountNumber,
	)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("task", "HandleInboundTransferWebhook").
			Str("account", p.TransactionDetails.DestinationAccountInfo.AccountNumber).
			Msg("failed to credit account")
		return err
	}

	// Step 2: Record transaction
	metaBytes, _ := json.Marshal(p.TransactionDetails.Metadata)

	recordInput := transaction.NewTransactionInput{
		AccountID: result.AcctId,
		Amount:    amount,
		Status:    transaction.StatusSuccess,
		Description: fmt.Sprintf("Credit of ₦%s from %s",
			p.TransactionDetails.SettlementAmount,
			p.TransactionDetails.Customer.Name),
		Reference: p.TransactionDetails.PaymentReference,
		Channel:   "Monnify",
		Meta:      metaBytes,
	}

	if _, err := h.TrxSvc.RecordTransaction(ctx, &recordInput); err != nil {
		// Check for unique key violation
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" { // unique_violation
				// Duplicate transaction — means we've already recorded it
				// stop worker processing
				h.logger.Info().
					Str("task", "HandleInboundTransferWebhook").
					Str("account_id", result.AcctId.String()).
					Str("trx_ref", recordInput.Reference).
					Msg("duplicate transaction detected, skipping")
				return nil
			}
		}

		// Normal error path (log + return error)
		h.logger.Error().
			Err(err).
			Str("task", "HandleInboundTransferWebhook").
			Str("account_id", result.AcctId.String()).
			Msg("credit succeeded but failed to record transaction")
		return err
	}

	h.logger.Info().
		Str("task", "HandleInboundTransferWebhook").
		Str("reference", p.TransactionDetails.PaymentReference).
		Msg("inbound transfer processed successfully")

	return nil
}
