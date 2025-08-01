package transaction

import "github.com/shopspring/decimal"

type TransactionStatus string

const (
	StatusPending TransactionStatus = "PENDING"
	StatusSuccess TransactionStatus = "SUCCESS"
	StatusFailed  TransactionStatus = "FAILED"
)

type Transaction struct {
	Status TransactionStatus
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
