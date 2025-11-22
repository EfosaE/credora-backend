package queues

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/EfosaE/credora-backend/domain/operation"
	"github.com/EfosaE/credora-backend/service"
	operationsvc "github.com/EfosaE/credora-backend/service/operation"
	"github.com/hibiken/asynq"
)

type Handlers struct {
	EmailSvc     service.EmailService
	OperationSvc operationsvc.OperationService
}

func NewHandlers(es service.EmailService, ops operationsvc.OperationService) *Handlers {
	return &Handlers{es, ops}
}

// // EXTERNAL TRANSFER
// func (h *Handlers) HandleExternalTransfer(ctx context.Context, t *asynq.Task) error {
// 	var p ExternalTransferPayload
// 	if err := json.Unmarshal(t.Payload(), &p); err != nil {
// 		return err
// 	}

// 	return h.TransferSvc.ProcessExternalTransfer(ctx, p.TransferID)
// }

// // INTERNAL TRANSFER
// func (h *Handlers) HandleInternalTransfer(ctx context.Context, t *asynq.Task) error {
// 	var p InternalTransferPayload
// 	if err := json.Unmarshal(t.Payload(), &p); err != nil {
// 		return err
// 	}

// 	return h.TransferSvc.ProcessInternalTransfer(ctx, p.TxnID)
// }

// // VA WEBHOOK (CREDIT)
// func (h *Handlers) HandleVACredit(ctx context.Context, t *asynq.Task) error {
// 	var p VACreditPayload
// 	if err := json.Unmarshal(t.Payload(), &p); err != nil {
// 		return err
// 	}

// 	return h.WebhookSvc.ProcessVACredit(ctx, p.WebhookID)
// }

// EMAIL NOTIFICATION
func (h *Handlers) HandleSendEmail(ctx context.Context, t *asynq.Task) error {
	var p WelcomeEmailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	log.Printf("Sending Email to User: user_id=%d, template_id=%s", p.User.ID, p.TemplateID)
	return h.EmailSvc.SendWelcomeEmail(ctx, *p.User)
}

func (h *Handlers) HandleSendAccountNumberEmail(ctx context.Context, t *asynq.Task) error {
	var p AccountNumberEmailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	log.Printf("Sending Email to User: to=%s, bank=%s, account_number=%s", p.To, p.Bank, p.AccountNumber)
	return h.EmailSvc.SendAccountNumberEmail(ctx, p.To, p.Bank, p.AccountNumber)
}

func (h *Handlers) HandleInternalTransfer(ctx context.Context, t *asynq.Task) error {
	var p operation.InternalTransferReq
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	log.Printf("Initiating transfer of %d from %s to %s", p.Amount, p.FromAcctNum, p.ToAcctNum)

	return h.OperationSvc.InternalTransfer(ctx, &p)

}
