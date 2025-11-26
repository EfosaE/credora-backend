package operation

import (
	"errors"
	"github.com/shopspring/decimal"
)

// ============================================================================
// ENUMS
// ============================================================================

type OperationType string

const (
	OperationTypeInternalTransfer       OperationType = "INTERNAL_TRANSFER"
	OperationTypeExternalTransfer       OperationType = "EXTERNAL_TRANSFER"
	OperationTypeWebhookInboundTransfer OperationType = "WEBHOOK_INBOUND_TRANSFER"
)

type InternalTransferReq struct {
	FromAcctNum    string          `json:"from_acct_num"`
	ToAcctNum      string          `json:"to_acct_num"`
	Amount         decimal.Decimal `json:"amount"`
	IdempotencyKey string          `json:"idempotency_key"`
}

type InternalTransferDTO struct {
	ToAccount string `json:"to_account" validate:"required"`
	Amount    string `json:"amount" validate:"required,decimal"`
}

var ErrInsufficientFunds = errors.New("insufficient funds")
var ErrAccountNotFound = errors.New("account not found")
var ErrDuplicateRequest = errors.New("duplicate request detected")
var ErrInvalidAmount = errors.New("invalid transfer amount")
var ErrInvalidTransfer = errors.New("invalid transfer: source and destination accounts must differ")
