package idempotency

import (
	"encoding/json"

	"github.com/EfosaE/credora-backend/domain/transaction"
)

type IdempotencyData struct {
	IdemKey       string                        `json:"idemKey"`
	Status        transaction.TransactionStatus `json:"status"`
	OperationType string                        `json:"operationType"`
	Payload       json.RawMessage               `json:"payload"`
}

// // IsValid checks if the status is one of the allowed values
// func (s IdempotencyStatus) IsValid() bool {
// 	switch s {
// 	case StatusPending, StatusProcessing, StatusSuccess, StatusFailed:
// 		return true
// 	default:
// 		return false
// 	}
// }
