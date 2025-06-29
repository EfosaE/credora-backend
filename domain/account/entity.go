package account

import (
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID             uuid.UUID `json:"id"`
	UserId         string    `json:"user_id"`
	AccountNumber  string    `json:"account_number"`
	AccountType    string    `json:"account_type"`
	MonnifyCustRef string    `json:"monnify_cust_ref"`
	BankName       string    `json:"bank_name"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreateAccountRequest struct {
	UserId         uuid.UUID `json:"user_id" validate:"required,uuid4"`
	AccountNumber  string `json:"account_number" validate:"required,len=10,numeric"`
	AccountType    string `json:"account_type" validate:"required,oneof=savings current"`
	BankName       string `json:"bank_name" validate:"required,min=2,max=100"`
	MonnifyCustRef string `json:"monnify_cust_ref" validate:"required,min=5,max=50"`
}

// type CreateUserResponse struct {
// 	ID                   uuid.UUID                 `json:"id"`
// 	Name                 string                    `json:"name"`
// 	Email                string                    `json:"email"`
// 	AccountReference     string                    `json:"account_reference"`
// 	AccountName          string                    `json:"account_name"`
// 	Accounts             []monnify.ReservedAccount `json:"accounts"`
// 	ReservationReference string                    `json:"reservation_reference"`
// 	BankName             string                    `json:"bank_name,omitempty"`      // fallback if you just need one
// 	AccountNumber        string                    `json:"account_number,omitempty"` // fallback if you just need one
// 	Status               string                    `json:"status"`                   // Monnify reserved account status
// 	CreatedAt            time.Time                 `json:"created_at"`
// }
