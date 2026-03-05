package queues

import (
	"github.com/EfosaE/credora-backend/domain/monnify"
	"github.com/EfosaE/credora-backend/domain/operation"
	"github.com/EfosaE/credora-backend/domain/user"
)

// A list of task types.
const (
	TypeWelcomeEmail           = "email:welcome"
	TypeAccountNumberEmail     = "email:account_number"
	TypeInternalTransfer       = "operation:internal_transfer"
	TypeWebhookInboundTransfer = "webhook:inbound_transfer"
	// TypeExternalTransfer = "operation:external_transfer"
	// TypeVACredit         = "webhook:va_credit"
	// TypeImageResize     = "image:resize"
)

type AccountNumberEmailPayload struct {
	To            string
	Accounts []monnify.ReservedAccount
}

type WelcomeEmailPayload struct {
	User       *user.User
	TemplateID string
}

type InternalTransferTaskPayload struct {
	Req      *operation.InternalTransferReq `json:"req"`
	QueuedAt int64                          `json:"queued_at"` // unix nano
}
