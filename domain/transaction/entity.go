package transaction

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ============================================================================
// ENUMS
// ============================================================================

type TransactionStatus string
type TransactionType string

const (
	TransactionTypeCredit TransactionType = "CREDIT"
	TransactionTypeDebit  TransactionType = "DEBIT"
)

const (
	StatusPending TransactionStatus = "PENDING"
	StatusSuccess TransactionStatus = "SUCCESS"
	StatusFailed  TransactionStatus = "FAILED"
)

// ============================================================================
// MODELS
// ============================================================================

// Transaction - A single ledger entry (credit or debit)
type Transaction struct {
	ID          uuid.UUID         `json:"id"`
	AccountID   uuid.UUID         `json:"account_id"`
	Amount      decimal.Decimal   `json:"amount"`
	Type        TransactionType   `json:"type"` // Added: CREDIT or DEBIT
	Status      TransactionStatus `json:"status"`
	Description string            `json:"description"`
	Reference   string            `json:"reference"`
	Channel     string            `json:"channel"`
	Meta        json.RawMessage   `json:"meta,omitempty"` // Changed from []byte
	CreatedAt   time.Time         `json:"created_at"`
}

// ============================================================================
// INPUT TYPES
// ============================================================================

// NewTransactionInput - For creating a single transaction
type NewTransactionInput struct {
	AccountID       uuid.UUID         // The account being debited/credited
	CounterpartyID  *uuid.UUID        // Optional: the other account (for internal transfers)
	Amount          decimal.Decimal
	Type            TransactionType   // CREDIT or DEBIT
	Status          TransactionStatus
	Description     string
	Reference       string
	Channel         string
	Meta            json.RawMessage
}

// AddMoneyInput - For depositing money
type AddMoneyInput struct {
	AccountID uuid.UUID
	Amount    decimal.Decimal
	Channel   string
	Reference string          // Payment reference
	Meta      json.RawMessage // Payment details
}

// InternalTransferInput - For transfers between accounts
type InternalTransferInput struct {
	FromAccountID uuid.UUID
	ToAccountID   uuid.UUID
	Amount        decimal.Decimal
	Description   string
	Channel       string
	Reference     string
	Meta          json.RawMessage
}