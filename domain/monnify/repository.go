// infrastructure/monnify/repository.go
package monnify

import (
	"fmt"
	"time"
)





type MonnifyRepository interface {
	Authenticate() error
	CreateReservedAccount(req  *CreateCRAParams) (*CreateCRAResponse, error)
	DeleteReservedAccount(accountRef string) (*CreateCRAResponse, error)
	// VerifyTransaction(reference string) (*TransactionStatus, error)
	// InitiateTransfer(req *PayoutRequest) (*PayoutResponse, error)
	ValidateWebhookSignature(body []byte, signature string) bool
}


func ParseMonnifyTime(s string) (time.Time, error) {
    layouts := []string{
        "2006-01-02 15:04:05.0",  // common format
        "2006-01-02 15:04:05",    // sometimes missing .0
    }

    for _, layout := range layouts {
        t, err := time.Parse(layout, s)
        if err == nil {
            return t, nil
        }
    }

    return time.Time{}, fmt.Errorf("invalid Monnify time: %s", s)
}
