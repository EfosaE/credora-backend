package account

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Account struct {
	ID             uuid.UUID       `json:"id"`
	UserId         string          `json:"user_id"`
	AccountNumber  string          `json:"account_number"`
	AccountType    string          `json:"account_type"`
	MonnifyCustRef string          `json:"monnify_cust_ref"`
	BankName       string          `json:"bank_name"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	Balance        decimal.Decimal `json:"balance"`
}

type CreateAccountRequest struct {
	UserId         uuid.UUID `json:"user_id" validate:"required,uuid4"`
	AccountNumber  string    `json:"account_number" validate:"required,len=10,numeric"`
	AccountType    string    `json:"account_type" validate:"required,oneof=savings current"`
	BankName       string    `json:"bank_name" validate:"required,min=2,max=100"`
	MonnifyCustRef string    `json:"monnify_cust_ref" validate:"required,min=5,max=50"`
}

type TransferRequest struct {
	AccountNumber string // Virtual account to credit
	BankName      string
	Amount        float64
	Description   string
	ForceFail     bool // Optional: trigger failure simulation
}

type CreditAcctResp struct {
	AcctId  uuid.UUID       `json:"acctID"`
	Balance decimal.Decimal `json:"balance"`
}


