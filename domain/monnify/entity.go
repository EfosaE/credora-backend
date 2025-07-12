package monnify

type MonnifyResp struct {
	RequestSuccessful bool   `json:"requestSuccessful"`
	ResponseMessage   string `json:"responseMessage"`
	ResponseCode      string `json:"responseCode"`
}

type MonnifyAuthResponse struct {
	MonnifyResp
	ResponseBody    struct {
		AccessToken string `json:"accessToken"`
		ExpiresIn   int    `json:"expiresIn"`
	} `json:"responseBody"`
}


// CRA = Customer Reserved Account
type CreateCRAParams struct {
	AccountReference     string `json:"accountReference"` // Unique to the customer (can be user ID)
	AccountName          string `json:"accountName"`      // Same as customerName
	CurrencyCode         string `json:"currencyCode"`     // Usually "NGN"
	ContractCode         string `json:"contractCode"`     // From your Monnify dashboard
	CustomerEmail        string `json:"customerEmail"`
	Nin                  string `json:"nin"`
	CustomerName         string `json:"customerName"`
	GetAllAvailableBanks bool   `json:"getAllAvailableBanks"`
}

// CRA = Customer Reserved Account
type CreateCRAResponse struct {
	MonnifyResp
	ResponseBody      CreateCRAResponseBody `json:"responseBody"`
}

// CRA = Customer Reserved Account
type CreateCRAResponseBody struct {
	ContractCode          string              `json:"contractCode"`
	AccountReference      string              `json:"accountReference"`
	AccountName           string              `json:"accountName"`
	CurrencyCode          string              `json:"currencyCode"`
	CustomerEmail         string              `json:"customerEmail"`
	CustomerName          string              `json:"customerName"`
	Accounts              []ReservedAccount   `json:"accounts"`
	CollectionChannel     string              `json:"collectionChannel"`
	ReservationReference  string              `json:"reservationReference"`
	ReservedAccountType   string              `json:"reservedAccountType"`
	Status                string              `json:"status"`
	CreatedOn             string              `json:"createdOn"` // consider using `time.Time` if you parse it
	IncomeSplitConfig     []IncomeSplitConfig `json:"incomeSplitConfig"`
	Nin                   string              `json:"nin"` //Nin or BVN, I chose NIn
	RestrictPaymentSource bool                `json:"restrictPaymentSource"`
}

type ReservedAccount struct {
	BankCode      string `json:"bankCode"`
	BankName      string `json:"bankName"`
	AccountNumber string `json:"accountNumber"`
	AccountName   string `json:"accountName"`
}
type IncomeSplitConfig struct {
	// Empty array in your example.
	// You can leave this empty or define structure based on Monnify docs if needed later.
}

type MonnifyConfig struct {
	ApiKey       string
	SecretKey    string
	ContractCode string
	BaseURL      string
	Token        string
}
