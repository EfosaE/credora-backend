package queues

import (
	"encoding/json"

	"github.com/EfosaE/credora-backend/domain/operation"
	"github.com/EfosaE/credora-backend/domain/user"
	"github.com/EfosaE/credora-backend/domain/webhook"
	"github.com/hibiken/asynq"
)

// A list of task types.
const (
	TypeWelcomeEmail       = "email:welcome"
	TypeAccountNumberEmail = "email:account_number"
	TypeInternalTransfer   = "operation:internal_transfer"
	TypeWebhookInboundTransfer = "webhook:inbound_transfer"
	// TypeExternalTransfer = "operation:external_transfer"
	// TypeVACredit         = "webhook:va_credit"
	// TypeImageResize     = "image:resize"
)

type AccountNumberEmailPayload struct {
	To            string
	Bank          string
	AccountNumber string
}

type WelcomeEmailPayload struct {
	User       *user.User
	TemplateID string
}

// type ImageResizePayload struct {
//     SourceURL string
// }

// ----------------------------------------------
// Write a function NewXXXTask to create a task.
// A task consists of a type and a payload.
// ----------------------------------------------
func NewAccountNumberEmailTask(to string, bank string, accountNumber string) (*asynq.Task, error) {
	payload, err := json.Marshal(AccountNumberEmailPayload{To: to, Bank: bank, AccountNumber: accountNumber})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeAccountNumberEmail, payload), nil
}

func NewWelcomeEmailTask(user *user.User, tmplID string) (*asynq.Task, error) {
	payload, err := json.Marshal(WelcomeEmailPayload{User: user, TemplateID: tmplID})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeWelcomeEmail, payload), nil
}

func NewInternalTransferTask(fromAcctNum string, req *operation.InternalTransferReq) (*asynq.Task, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeInternalTransfer, payload), nil
}

func NewWebhookInboundTransferTask(req *webhook.InboundTransferPayload) (*asynq.Task, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeWebhookInboundTransfer, payload), nil
}



// func NewImageResizeTask(src string) (*asynq.Task, error) {
//     payload, err := json.Marshal(ImageResizePayload{SourceURL: src})
//     if err != nil {
//         return nil, err
//     }
//     // task options can be passed to NewTask, which can be overridden at enqueue time.
//     return asynq.NewTask(TypeImageResize, payload, asynq.MaxRetry(5), asynq.Timeout(20 * time.Minute)), nil
// }

//---------------------------------------------------------------
// Write a function HandleXXXTask to handle the input task.
// Note that it satisfies the asynq.HandlerFunc interface.
//
// Handler doesn't need to be a function. You can define a type
// that satisfies asynq.Handler interface. See examples below.
//---------------------------------------------------------------

// // ImageProcessor implements asynq.Handler interface.
// type ImageProcessor struct {
//     // ... fields for struct
// }

// func (processor *ImageProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
//     var p ImageResizePayload
//     if err := json.Unmarshal(t.Payload(), &p); err != nil {
//         return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
//     }
//     log.Printf("Resizing image: src=%s", p.SourceURL)
//     // Image resizing code ...
//     return nil
// }

// func NewImageProcessor() *ImageProcessor {
// 	return &ImageProcessor{}
// }
