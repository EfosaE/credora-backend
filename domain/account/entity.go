package account

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TxResult struct {
	AccountNumber string
	NewBalance    decimal.Decimal
	Amount        decimal.Decimal
	Reason        string
}

type Account struct {
	ID             uuid.UUID       `json:"id"`
	UserId         string          `json:"userId"`
	UserName       string          `json:"userName"`
	AccountNumber  string          `json:"accountNumber"`
	AccountType    string          `json:"accountType"`
	MonnifyCustRef string          `json:"monnifyCustRef"`
	BankName       string          `json:"bankName"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
	Balance        decimal.Decimal `json:"balance"`
}

type CreateAccountRequest struct {
	UserId         uuid.UUID `json:"userId" validate:"required,uuid4"`
	Username       string    `json:"userName" validate:"required"`
	AccountNumber  string    `json:"accountNumber" validate:"required,len=10,numeric"`
	AccountType    string    `json:"accountType" validate:"required,oneof=savings current"`
	BankName       string    `json:"bankName" validate:"required,min=2,max=100"`
	MonnifyCustRef string    `json:"monnifyCustRef" validate:"required,min=5,max=50"`
}

type TransferRequest struct {
	AccountNumber string // Virtual account to credit
	BankName      string
	Amount        float64
	Description   string
	ForceFail     bool // Optional: trigger failure simulation
}

type CreditAcctResp struct {
	AcctId  uuid.UUID       `json:"acctId"`
	Balance decimal.Decimal `json:"balance"`
}

type GetUserDetailsWithAccountRow struct {
	UserId        uuid.UUID `json:"userId"`
	Password      string    `json:"-"`
	FullName      string    `json:"fullName"`
	Email         string    `json:"email"`
	PhoneNumber   string    `json:"phoneNumber"`
	IsVerified    bool      `json:"isVerified"`
	AccountNumber string    `json:"accountNumber"`
	AccountType   string    `json:"accountType"`
	Balance       string    `json:"balance"`
	Currency      string    `json:"currency"`
}
