package stubs

import "github.com/EfosaE/credora-backend/domain/monnify"

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
