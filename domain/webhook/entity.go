package webhook

import (
	"time"

	"github.com/shopspring/decimal"
)

type EventType string

const (
	EventSuccessfulTransaction  EventType = "SUCCESSFUL_TRANSACTION"
	EventCancelledTransaction   EventType = "CANCELLED_TRANSACTION"
	EventReversedTransaction    EventType = "TRANSACTION_REVERSED"
	EventSuccessfulDisbursement EventType = "SUCCESSFUL_DISBURSEMENT"
	EventFailedDisbursement     EventType = "FAILED_DISBURSEMENT"
)

type SuccessfulTransaction struct {
	Product                  Product            `json:"product"`
	TransactionReference     string             `json:"transactionReference"`
	PaymentReference         string             `json:"paymentReference"`
	PaidOnRaw                string             `json:"paidOn"` // Monnify time is not a valid time.Time that Go expects
	PaidOn                   time.Time          `json:"-"`
	PaymentDescription       string             `json:"paymentDescription"`
	Metadata                 map[string]any     `json:"metaData"`
	PaymentSourceInformation []PaymentSource    `json:"paymentSourceInformation"`
	DestinationAccountInfo   DestinationAccount `json:"destinationAccountInformation"`
	AmountPaid               decimal.Decimal    `json:"amountPaid"`
	TotalPayable             decimal.Decimal    `json:"totalPayable"`
	CardDetails              map[string]any     `json:"cardDetails"`
	PaymentMethod            string             `json:"paymentMethod"`
	Currency                 string             `json:"currency"`
	SettlementAmount         string             `json:"settlementAmount"`
	PaymentStatus            string             `json:"paymentStatus"`
	Customer                 Customer           `json:"customer"`
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

type InboundTransferPayload struct {
	TransactionDetails SuccessfulTransaction `json:"transactionDetails"`
}



type SuccessfulDisbursementData struct {
	Amount                   decimal.Decimal `json:"amount,omitempty"`
	TransactionReference     string          `json:"transactionReference,omitempty"`
	Fee                      decimal.Decimal `json:"fee,omitempty"`
	TransactionDescription   string          `json:"transactionDescription,omitempty"`
	DestinationAccountNumber string          `json:"destinationAccountNumber,omitempty"`
	SessionID                string          `json:"sessionId,omitempty"`
	CreatedOn                string          `json:"createdOn,omitempty"`
	DestinationAccountName   string          `json:"destinationAccountName,omitempty"`
	Reference                string          `json:"reference,omitempty"`
	DestinationBankCode      string          `json:"destinationBankCode,omitempty"`
	CompletedOn              string          `json:"completedOn,omitempty"`
	Narration                string          `json:"narration,omitempty"`
	Currency                 string          `json:"currency,omitempty"`
	DestinationBankName      string          `json:"destinationBankName,omitempty"`
	Status                   string          `json:"status,omitempty"`
}
