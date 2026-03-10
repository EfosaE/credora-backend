package account

import (
	"time"

	"github.com/EfosaE/credora-backend/domain/monnify"
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
	UserId         uuid.UUID       `json:"userId"`
	UserName       string          `json:"userName"`
	AccountNumber  string          `json:"accountNumber"`
	AccountType    string          `json:"accountType"`
	MonnifyCustRef string          `json:"monnifyCustomerReference"`
	BankName       string          `json:"bankName"`
	Balance        decimal.Decimal `json:"balance"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
}
type CreateAccountRequest struct {
	UserId         uuid.UUID                 `json:"userId" validate:"required,uuid4"`
	Username       string                    `json:"userName" validate:"required"`
	Accounts       []monnify.ReservedAccount `json:"accounts"`
	AccountType    string                    `json:"accountType" validate:"required,oneof=savings current"`
	MonnifyCustRef string                    `json:"monnifyCustRef" validate:"required,min=5,max=50"`
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

type InternalTransferResp struct {
	FromAccountId uuid.UUID
	FromBalance   decimal.Decimal
	ToAccountId   uuid.UUID
	ToBalance     decimal.Decimal
	ToUserId      uuid.UUID
	FromUserId    uuid.UUID
}

type GetAccountWithUserInfo struct {
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
