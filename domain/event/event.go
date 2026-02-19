package event

import (
	// "github.com/EfosaE/credora-backend/domain/user"
	"github.com/google/uuid"
)

type UserCreatedEvent struct {
	UserID        uuid.UUID `json:"userId"`
	AccountNumber string    `json:"accountNumber"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	BankName      string    `json:"bankName"`
}
