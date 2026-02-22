package event

import (
	// "github.com/EfosaE/credora-backend/domain/user"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const (
	EventUserCreated             = "user.created"
	EventInternalTransferSuccess = "transfer.success"
	StreamUserEvents             = "stream:users"
	StreamTransferEvents         = "stream:transfers"
	// StreamNotificationEvents = "stream:notifications"
)

type UserCreatedEvent struct {
	UserID        uuid.UUID `json:"userId"`
	AccountNumber string    `json:"accountNumber"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	BankName      string    `json:"bankName"`
}

type InternalTransferEvent struct {
	TransferID    string //Should be the transaction reference
	FromAcctNum   string
	RecipientName string
	SenderName    string
	ToAcctNum     string
	ToAcctUserId  uuid.UUID
	Amount        decimal.Decimal
	OccurredAt    time.Time
}

type EventMessage struct {
	EventType string //user.created
	Data      string //The payload is stringified
}
