package user

import (
	"time"

	"github.com/EfosaE/credora-backend/domain/monnify"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Nin       string    `json:"nin"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateUserRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Email       string `json:"email" validate:"required,email,max=255"`
	Nin         string `json:"nin" validate:"required,min=11,max=11"`
	Password    string `json:"password" validate:"required,min=8,max=100"`
	PhoneNumber string `json:"phone_number" validate:"required,min=11,max=15"`
}

type CreateUserResponse struct {
	ID                   uuid.UUID                 `json:"id"`
	Name                 string                    `json:"name"`
	Email                string                    `json:"email"`
	AccountReference     string                    `json:"account_reference"`
	AccountName          string                    `json:"account_name"`
	Accounts             []monnify.ReservedAccount `json:"accounts"`
	ReservationReference string                    `json:"reservation_reference"`
	BankName             string                    `json:"bank_name,omitempty"`      // fallback if you just need one
	AccountNumber        string                    `json:"account_number,omitempty"` // fallback if you just need one
	Status               string                    `json:"status"`                   // Monnify reserved account status
	CreatedAt            time.Time                `json:"created_at"`
}
