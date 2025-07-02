package event

import (
	// "github.com/EfosaE/credora-backend/domain/user"
	"github.com/google/uuid"
)

type UserCreatedEvent struct {
	UserID        uuid.UUID `json:"user_id"`
	AccountNumber string    `json:"account_number"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	BankName      string    `json:"bank_name"`
}
