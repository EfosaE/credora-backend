//infrastructure/monnify/repository.go
package monnify



type MonnifyRepository interface {
	Authenticate() error
	CreateReservedAccount(req  *CreateCustomerRequest) (*CreateCustomerResponse, error)
	// VerifyTransaction(reference string) (*TransactionStatus, error)
	// InitiateTransfer(req *PayoutRequest) (*PayoutResponse, error)
	ValidateWebhookSignature(body []byte, signature string) bool
}
