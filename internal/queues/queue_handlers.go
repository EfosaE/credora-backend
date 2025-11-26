package queues

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/EfosaE/credora-backend/domain/logger"
	"github.com/EfosaE/credora-backend/domain/operation"
	"github.com/EfosaE/credora-backend/domain/transaction"
	"github.com/EfosaE/credora-backend/domain/webhook"
	"github.com/EfosaE/credora-backend/service"
	accountsvc "github.com/EfosaE/credora-backend/service/account"
	operationsvc "github.com/EfosaE/credora-backend/service/operation"
	transactionsvc "github.com/EfosaE/credora-backend/service/transaction"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"
)

type Handlers struct {
	EmailSvc     service.EmailService
	OperationSvc operationsvc.OperationService
	AcctSvc      *accountsvc.AccountService
	TrxSvc       *transactionsvc.TransactionService
	AppLogger    *logger.Logger
}

func NewHandlers(es service.EmailService, ops operationsvc.OperationService, acctSvc *accountsvc.AccountService, trxSvc *transactionsvc.TransactionService, appLogger *logger.Logger) *Handlers {
	return &Handlers{es, ops, acctSvc, trxSvc, appLogger}
}

// EMAIL NOTIFICATION
func (h *Handlers) HandleSendEmail(ctx context.Context, t *asynq.Task) error {
	var p WelcomeEmailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		h.logError(ctx, err, "failed to unmarshal WelcomeEmailPayload", map[string]any{
			"task": "HandleSendEmail",
		})
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	h.logInfo("sending welcome email", map[string]any{
		"user_id":     p.User.ID,
		"template_id": p.TemplateID,
	})

	if err := h.EmailSvc.SendWelcomeEmail(ctx, *p.User); err != nil {
		h.logError(ctx, err, "failed to send welcome email", map[string]any{
			"user_id": p.User.ID,
		})
		return err
	}

	return nil
}

func (h *Handlers) HandleSendAccountNumberEmail(ctx context.Context, t *asynq.Task) error {
	var p AccountNumberEmailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		h.logError(ctx, err, "failed to unmarshal AccountNumberEmailPayload", nil)
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	h.logInfo("sending account number email", map[string]any{
		"to":            p.To,
		"bank":          p.Bank,
		"accountNumber": p.AccountNumber,
	})

	if err := h.EmailSvc.SendAccountNumberEmail(ctx, p.To, p.Bank, p.AccountNumber); err != nil {
		h.logError(ctx, err, "failed to send account number email", map[string]any{
			"to": p.To,
		})
		return err
	}

	return nil
}

// TRANSFER HANDLER
func (h *Handlers) HandleInternalTransfer(ctx context.Context, t *asynq.Task) error {
	var p operation.InternalTransferReq
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		h.logError(ctx, err, "failed to unmarshal InternalTransferReq", nil)
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	h.logInfo("processing internal transfer", map[string]any{
		"from":   p.FromAcctNum,
		"to":     p.ToAcctNum,
		"amount": p.Amount.String(),
	})

	err := h.OperationSvc.InternalTransfer(ctx, &p)
	if err != nil {
		switch {
		case errors.Is(err, operation.ErrInsufficientFunds):
			h.logInfo("transfer failed - insufficient funds", map[string]any{
				"from": p.FromAcctNum,
			})
			return nil

		case errors.Is(err, operation.ErrAccountNotFound):
			h.logInfo("transfer failed - account not found", map[string]any{
				"from": p.FromAcctNum,
				"to":   p.ToAcctNum,
			})
			return nil

		default:
			h.logError(ctx, err, "internal transfer failed (retrying)", map[string]any{
				"from": p.FromAcctNum,
				"to":   p.ToAcctNum,
			})
			return err
		}
	}

	h.logInfo("internal transfer completed successfully", map[string]any{
		"from":   p.FromAcctNum,
		"to":     p.ToAcctNum,
		"amount": p.Amount.String(),
	})

	return nil
}

func (h *Handlers) HandleInboundTransferWebhook(ctx context.Context, t *asynq.Task) error {
	var p webhook.InboundTransferPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		h.logError(ctx, err, "failed to unmarshal InboundTransferPayload", nil)
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	amount, _ := decimal.NewFromString(p.TransactionDetails.SettlementAmount)
	h.logInfo("processing inbound transfer webhook", map[string]any{
		"amount":      amount.String(),
		"destination": p.TransactionDetails.DestinationAccountInfo.AccountNumber,
		"reference":   p.TransactionDetails.PaymentReference,
	})

	// Step 1: Credit account
	result, err := h.AcctSvc.CreditUserBalance(ctx, amount,
		p.TransactionDetails.DestinationAccountInfo.AccountNumber,
	)
	if err != nil {
		h.logError(ctx, err, "failed to credit account", map[string]any{
			"account": p.TransactionDetails.DestinationAccountInfo.AccountNumber,
		})
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
				h.logInfo("duplicate transaction detected, skipping", map[string]any{
					"account_id": result.AcctId,
					"trx_ref":    recordInput.Reference,
				})
				return nil
			}
		}

		// Normal error path (log + return error)
		h.logError(ctx, err, "credit succeeded but failed to record transaction", map[string]any{
			"account_id": result.AcctId,
		})
		return err
	}

	h.logInfo("inbound transfer processed successfully", map[string]any{
		"reference": p.TransactionDetails.PaymentReference,
	})

	return nil
}
