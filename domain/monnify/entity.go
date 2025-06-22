package monnify

type CreateCustomerRequest struct {
	AccountReference string `json:"accountReference"` // Unique to the customer (can be user ID)
	AccountName      string `json:"accountName"`      // Same as customerName
	CurrencyCode     string `json:"currencyCode"`     // Usually "NGN"
	ContractCode     string `json:"contractCode"`     // From your Monnify dashboard
	CustomerEmail    string `json:"customerEmail"`
	BVN              string `json:"bvn,omitempty"` // Optional
	CustomerName     string `json:"customerName"`
}

type CreateCustomerResponse struct {
	RequestSuccessful bool                 `json:"requestSuccessful"`
	ResponseMessage   string               `json:"responseMessage"`
	ResponseCode      string               `json:"responseCode"`
	ResponseBody      CustomerResponseBody `json:"responseBody"`
}

type CustomerResponseBody struct {
	AccountReference     string `json:"accountReference"`
	AccountName          string `json:"accountName"`
	AccountNumber        string `json:"accountNumber"`
	BankName             string `json:"bankName"`
	BankCode             string `json:"bankCode"`
	CustomerEmail        string `json:"customerEmail"`
	CustomerName         string `json:"customerName"`
	ReservationReference string `json:"reservationReference"`
}

type MonnifyAuthResponse struct {
	ResponseMessage string `json:"responseMessage"`
	ResponseCode    string `json:"responseCode"`
	ResponseBody    struct {
		AccessToken string `json:"accessToken"`
		ExpiresIn   int    `json:"expiresIn"`
	} `json:"responseBody"`
}

type MonnifyConfig struct {
	ApiKey       string
	SecretKey    string
	ContractCode string
	BaseURL      string
	Token        string
}
