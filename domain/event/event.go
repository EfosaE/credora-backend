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
	FromAcctId     uuid.UUID
	ToAcctId       uuid.UUID
	FromAcctNum    string
	ToAcctNum      string
	ToAcctUserId   uuid.UUID
	FromAcctUserId uuid.UUID
	Amount         decimal.Decimal
	OccurredAt     time.Time
	TransactionRef string
}

type EventMessage struct {
	EventType string //user.created
	Data      string //The payload is stringified
}
