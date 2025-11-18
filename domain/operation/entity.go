package operation

import (
	"errors"
	"github.com/shopspring/decimal"
)

type InternalTransferReq struct {
	FromAcctNum    string          `json:"from_acct_num"`
	ToAcctNum      string          `json:"to_acct_num"`
	Amount         decimal.Decimal `json:"amount"`
	IdempotencyKey string          `json:"idempotency_key"`
}

type InternalTransferDTO struct {
	FromAccount string `json:"from_account" validate:"required"`
	ToAccount   string `json:"to_account" validate:"required"`
	Amount      string `json:"amount" validate:"required,decimal"`
}

var ErrInsufficientFunds = errors.New("insufficient funds")
var ErrDuplicateRequest = errors.New("duplicate request detected")
var ErrInvalidAmount = errors.New("invalid transfer amount")
var ErrInvalidTransfer = errors.New("invalid transfer: source and destination accounts must differ")