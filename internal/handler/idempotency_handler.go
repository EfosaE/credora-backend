// handlers/user_handler.go
package handler

import (
	"fmt"
	"net/http"

	// "github.com/EfosaE/credora-backend/domain/idempotency"
	// "github.com/EfosaE/credora-backend/domain/operation"
	// "github.com/EfosaE/credora-backend/domain/transaction"
	"github.com/EfosaE/credora-backend/domain/idempotency"
	"github.com/EfosaE/credora-backend/domain/operation"
	"github.com/EfosaE/credora-backend/domain/transaction"
	"github.com/EfosaE/credora-backend/internal/response"
	idempotencysvc "github.com/EfosaE/credora-backend/service/idempotency"
)

type IdempotencyHandler struct {
	idempSvc *idempotencysvc.IdempotencyService
}

func NewIdempotencyHandler(idempSvc *idempotencysvc.IdempotencyService) *IdempotencyHandler {
	return &IdempotencyHandler{idempSvc: idempSvc}
}


func (h *IdempotencyHandler) CreateRecordHandler(w http.ResponseWriter, r *http.Request) {

	payload := &idempotency.IdempotencyData{
		IdemKey:       "Monnify:23456",
		Status:        transaction.StatusSuccess,
		OperationType: string(operation.OperationTypeWebhookInboundTransfer),
	}

	// Call the Monnify service to delete the reserved account
	err := h.idempSvc.AddToIdempotencyTable(r.Context(), payload.IdemKey, operation.OperationType(payload.OperationType), payload, payload.Status)
	// err := h.idempSvc.MarkProcessed(r.Context(), "Monnify:23457")
	if err != nil {
		fmt.Println("Error occurred in adding record", err)
		response.SendError(w, r, response.InternalServerError(err, "could not add webhook payload"))
		return
	}

	// Success
	response.SendSuccess(w, r, response.OK(
		nil,
		nil,
		"Test key added successfully",
	))
}

// check
func (h *IdempotencyHandler) CheckRecordHandler(w http.ResponseWriter, r *http.Request) {
	exists, err := h.idempSvc.Exists(r.Context(), "Monnify:23456")

	if err != nil {
		fmt.Println("Error occurred in checking record", err)
		response.SendError(w, r, response.InternalServerError(err, "could not check idempotency payload"))
		return
	}

	// Success
	response.SendSuccess(w, r, response.OK(
		map[string]any{"exists": exists},
		nil,
		"",
	))
}
