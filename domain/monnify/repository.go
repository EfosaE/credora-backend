// infrastructure/monnify/repository.go
package monnify





type MonnifyRepository interface {
	Authenticate() error
	CreateReservedAccount(req  *CreateCRAParams) (*CreateCRAResponse, error)
	DeleteReservedAccount(accountRef string) (*CreateCRAResponse, error)
	// VerifyTransaction(reference string) (*TransactionStatus, error)
	// InitiateTransfer(req *PayoutRequest) (*PayoutResponse, error)
	ValidateWebhookSignature(body []byte, signature string) bool
}
