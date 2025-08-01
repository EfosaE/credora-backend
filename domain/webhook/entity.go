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
	Product                  Product                `json:"product"`
	TransactionReference     string                 `json:"transactionReference"`
	PaymentReference         string                 `json:"paymentReference"`
	PaidOn                   time.Time              `json:"paidOn"`
	PaymentDescription       string                 `json:"paymentDescription"`
	Metadata                 map[string]interface{} `json:"metaData"`
	PaymentSourceInformation []PaymentSource        `json:"paymentSourceInformation"`
	DestinationAccountInfo   DestinationAccount     `json:"destinationAccountInformation"`
	AmountPaid               decimal.Decimal        `json:"amountPaid"`
	TotalPayable             decimal.Decimal        `json:"totalPayable"`
	CardDetails              map[string]interface{} `json:"cardDetails"` // define explicitly if needed
	PaymentMethod            string                 `json:"paymentMethod"`
	Currency                 string                 `json:"currency"`
	SettlementAmount         string                 `json:"settlementAmount"`
	PaymentStatus            string                 `json:"paymentStatus"`
	Customer                 Customer               `json:"customer"`
}

type Product struct {
	Reference string `json:"reference"`
	Type      string `json:"type"`
}

type PaymentSource struct {
	BankCode      string          `json:"bankCode"`
	AmountPaid    decimal.Decimal `json:"amountPaid"`
	AccountName   string          `json:"accountName"`
	SessionID     string          `json:"sessionId"`
	AccountNumber string          `json:"accountNumber"`
}

type DestinationAccount struct {
	BankCode      string `json:"bankCode"`
	BankName      string `json:"bankName"`
	AccountNumber string `json:"accountNumber"`
}

type Customer struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
