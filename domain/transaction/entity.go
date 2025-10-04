package transaction

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionStatus string

const (
	StatusPending TransactionStatus = "PENDING"
	StatusSuccess TransactionStatus = "SUCCESS"
	StatusFailed  TransactionStatus = "FAILED"
)

type Transaction struct {
	ID          uuid.UUID         `json:"id"`
	AccountID   uuid.UUID         `json:"account_id"`
	Amount      decimal.Decimal   `json:"amount"`
	Status      TransactionStatus `json:"status"`
	Description string            `json:"description"`
	Reference   string            `json:"reference"`
	Channel     string            `json:"channel"`
	Meta        []byte            `json:"meta"`
	CreatedAt   time.Time         `json:"created_at"`
}

type NewTransactionInput struct {
	AccountID   uuid.UUID
	Amount      decimal.Decimal
	Description string
	Reference   string
	Channel     string
	Status      TransactionStatus
	Meta        []byte // for optional metadata
}

type AddMoneyToWalletRequest struct {
	Amount decimal.Decimal
}

// type Account struct {
// 	ID             uuid.UUID       `json:"id"`
// 	UserId         string          `json:"user_id"`
// 	AccountNumber  string          `json:"account_number"`
// 	AccountType    string          `json:"account_type"`
// 	MonnifyCustRef string          `json:"monnify_cust_ref"`
// 	BankName       string          `json:"bank_name"`
// 	CreatedAt      time.Time       `json:"created_at"`
// 	UpdatedAt      time.Time       `json:"updated_at"`
// 	Balance        decimal.Decimal `json:"balance"`
// }

// type CreateAccountRequest struct {
// 	UserId         uuid.UUID `json:"user_id" validate:"required,uuid4"`
// 	AccountNumber  string    `json:"account_number" validate:"required,len=10,numeric"`
// 	AccountType    string    `json:"account_type" validate:"required,oneof=savings current"`
// 	BankName       string    `json:"bank_name" validate:"required,min=2,max=100"`
// 	MonnifyCustRef string    `json:"monnify_cust_ref" validate:"required,min=5,max=50"`
// }
