// File: handler/operations_handler.go

package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/EfosaE/credora-backend/domain/operation"
	"github.com/EfosaE/credora-backend/domain/transaction"
	custmiddleware "github.com/EfosaE/credora-backend/internal/middleware"
	"github.com/EfosaE/credora-backend/internal/pgerrors"
	"github.com/EfosaE/credora-backend/internal/queues"
	"github.com/EfosaE/credora-backend/internal/response"
	idempotencysvc "github.com/EfosaE/credora-backend/service/idempotency"
	operationsvc "github.com/EfosaE/credora-backend/service/operation"
	"github.com/go-chi/chi/v5"
)

// OperationHandler handles operations like internal transfers
type OperationHandler struct {
	operationService   *operationsvc.OperationService
	idempotencyService *idempotencysvc.IdempotencyService
	queue              queues.Queue // It should depend on the Queue interface for unit testing
}

// NewOperationHandler creates a new handler
func NewOperationHandler(opService *operationsvc.OperationService, idempService *idempotencysvc.IdempotencyService, queue queues.Queue) *OperationHandler {
	return &OperationHandler{operationService: opService, idempotencyService: idempService, queue: queue}
}

// InternalTransfer handles POST /transfer
func (h *OperationHandler) InternalTransfer(w http.ResponseWriter, r *http.Request) {

	// Retrieve domain request injected by middleware
	req, ok := r.Context().Value(custmiddleware.ContextKeyInternalTransfer).(*operation.InternalTransferReq)
	if !ok || req == nil {
		response.SendError(w, r, response.BadRequest(
			errors.New("missing transfer request"),
			"Request was not parsed correctly",
		))
		return
	}
	// Wrap the req into InternalTransferTaskPayload
	taskPayload := queues.InternalTransferTaskPayload{
		Req:      req,
		QueuedAt: time.Now().UnixNano(), // capture enqueue timestamp
	}
	// Add the record first to the idempotency table with PENDING status
	if err := h.idempotencyService.AddToIdempotencyTable(r.Context(), req.IdempotencyKey, operation.OperationTypeInternalTransfer, &req, transaction.StatusPending); err != nil {

		if pgerrors.HandlePGError(w, r, err) {
			return
		}
		// fallback unexpected error
		response.SendError(w, r, response.InternalServerError(
			err,
			"Failed to add to idempotency table",
		))
		return

	}

	// Call the Enqueue for Internal trasnfer
	err := h.queue.EnqueueInternalTransfer(taskPayload)
	if err != nil {
		response.SendError(w, r, response.InternalServerError(
			err,
			"Failed to enqueue internal transfer",
		))
		return
	}

	response.SendSuccess(w, r, response.Accepted(
		response.ObjKV(response.KV{Key: "status", Value: "pending"}, response.KV{Key: "transferId", Value: req.IdempotencyKey}),
		nil,
		"Your request is being processed",
	))
}

func (h *OperationHandler) GetTransferStatus(w http.ResponseWriter, r *http.Request) {
	// Extract transaction / idempotency key from path
	trxID := chi.URLParam(r, "trxID")
	if trxID == "" {
		response.SendError(w, r, response.BadRequest(
			errors.New("missing transfer ID"),
			"Transfer ID is required",
		))
		return
	}

	// Query idempotency table for the status
	record, err := h.idempotencyService.GetRecord(r.Context(), trxID)
	if err != nil {
		if pgerrors.HandlePGError(w, r, err) {
			return
		}
		// fallback unexpected error
		response.SendError(w, r, response.InternalServerError(err, "Failed to retrieve transfer status"))
		return
	}

	response.SendSuccess(w, r, response.OK(
		response.Obj("status", &record),
		nil,
		"Transfer status retrieved successfully",
	))
}
