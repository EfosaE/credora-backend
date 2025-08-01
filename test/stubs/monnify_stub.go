package stubs

import (
	"math/rand"
	"time"

	"github.com/EfosaE/credora-backend/domain/monnify"
	"github.com/EfosaE/credora-backend/domain/simulator"
	"github.com/EfosaE/credora-backend/domain/webhook"
	"github.com/EfosaE/credora-backend/internal/utils"
	"github.com/shopspring/decimal"
)

var StubCreateCRAResponse = &monnify.CreateCRAResponse{
	MonnifyResp: monnify.MonnifyResp{
		RequestSuccessful: true,
		ResponseMessage:   "Account created successfully",
		ResponseCode:      "0",
	},
	ResponseBody: monnify.CreateCRAResponseBody{
		ContractCode:          "100693167467",
		AccountReference:      "REF123",
		AccountName:           "John Doe",
		CurrencyCode:          "NGN",
		CustomerEmail:         "john@example.com",
		CustomerName:          "John Doe",
		CollectionChannel:     "RESERVED_ACCOUNT",
		ReservationReference:  "ABC123456789",
		ReservedAccountType:   "GENERAL",
		Status:                "ACTIVE",
		CreatedOn:             "2024-11-25 07:35:17.566",
		Nin:                   "21212121212",
		RestrictPaymentSource: false,
		Accounts: []monnify.ReservedAccount{
			{
				BankCode:      "50515",
				BankName:      "Moniepoint Microfinance Bank",
				AccountNumber: "6839490147",
				AccountName:   "John Doe",
			},
		},
		IncomeSplitConfig: []monnify.IncomeSplitConfig{},
	},
}

var StubAuthenticateResponse = &monnify.MonnifyAuthResponse{
	MonnifyResp: monnify.MonnifyResp{
		RequestSuccessful: true,
		ResponseMessage:   "Success",
		ResponseCode:      "0",
	},
	ResponseBody: struct {
		AccessToken string `json:"accessToken"`
		ExpiresIn   int    `json:"expiresIn"`
	}{
		AccessToken: "mocked-token",
		ExpiresIn:   3567,
	},
}

// BuildSimulatedSuccessEvent builds a stub SuccessfulTransactionEvent dynamically.
func BuildSimulatedSuccessEventWbHk(req simulator.TransferRequest) *monnify.SuccessfulTransactionEvent {
	channelCode := rand.Intn(100)
	trnRef := utils.GenerateMonnifyReference(channelCode)

	return &monnify.SuccessfulTransactionEvent{
		EventType: webhook.EventSuccessfulTransaction,
		EventData: webhook.SuccessfulTransaction{
			Product: webhook.Product{
				Reference: "1636106097661",
				Type:      "RESERVED_ACCOUNT",
			},
			TransactionReference: trnRef,
			PaymentReference:     trnRef,
			PaidOn:               time.Date(2021, 11, 17, 11, 28, 42, 615000000, time.UTC), // parsed timestamp
			PaymentDescription:   "Adm",
			Metadata:             map[string]interface{}{},
			PaymentSourceInformation: []webhook.PaymentSource{
				{
					BankCode:      req.SenderBankCode,
					AmountPaid:    decimal.NewFromFloat(req.Amount),
					AccountName:   req.SenderName,
					SessionID:     req.SenderName,
					AccountNumber: req.SenderAccount,
				},
			},
			DestinationAccountInfo: webhook.DestinationAccount{
				BankCode:      "232",
				BankName:      req.RecipientBankName,
				AccountNumber: req.RecipientAccount,
			},
			AmountPaid:       decimal.NewFromFloat(req.Amount),
			TotalPayable:     decimal.NewFromFloat(req.Amount),
			CardDetails:      map[string]any{},
			PaymentMethod:    "ACCOUNT_TRANSFER",
			Currency:         "NGN",
			SettlementAmount: utils.CalculateSettlement(decimal.NewFromFloat(req.Amount)),
			PaymentStatus:    "PAID",
			Customer: webhook.Customer{ 
				Name:  "John Doe",
				Email: "test@tester.com",
			},
		},
	}
}
