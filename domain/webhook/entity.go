package webhook

import (
	"time"

	"github.com/shopspring/decimal"
)

type EventType string

const (
	EventSuccessfulTransaction EventType = "SUCCESSFUL_TRANSACTION"
	// EventFailedTransaction     EventType = "FAILED_TRANSACTION"
	EventCancelledTransaction EventType = "CANCELLED_TRANSACTION"
	EventReversedTransaction  EventType = "TRANSACTION_REVERSED"
	EventDisbursementSuccess  EventType = "DISBURSEMENT_SUCCESSFUL"
	EventDisbursementFailure  EventType = "DISBURSEMENT_FAILED"
)

type SuccessfulTransaction struct {
	Product                  Product            `json:"product"`
	TransactionReference     string             `json:"transaction_reference"`
	PaymentReference         string             `json:"payment_reference"`
	PaidOn                   time.Time          `json:"paid_on"`
	PaymentDescription       string             `json:"payment_description"`
	Metadata                 map[string]any     `json:"metadata"`
	PaymentSourceInformation []PaymentSource    `json:"payment_source_information"`
	DestinationAccountInfo   DestinationAccount `json:"destination_account_information"`
	AmountPaid               decimal.Decimal    `json:"amount_paid"`
	TotalPayable             decimal.Decimal    `json:"total_payable"`
	CardDetails              map[string]any     `json:"card_details"` // define explicitly if needed
	PaymentMethod            string             `json:"payment_method"`
	Currency                 string             `json:"currency"`
	SettlementAmount         string             `json:"settlement_amount"`
	PaymentStatus            string             `json:"payment_status"`
	Customer                 Customer           `json:"customer"`
}

type Product struct {
	Reference string `json:"reference"`
	Type      string `json:"type"`
}

type PaymentSource struct {
	BankCode      string          `json:"bank_code"`
	AmountPaid    decimal.Decimal `json:"amount_paid"`
	AccountName   string          `json:"account_name"`
	SessionID     string          `json:"session_id"`
	AccountNumber string          `json:"account_number"`
}

type DestinationAccount struct {
	BankCode      string `json:"bank_code"`
	BankName      string `json:"bank_name"`
	AccountNumber string `json:"account_number"`
}

type Customer struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
