// File: handler/operations_handler.go

package handler

import (
	"errors"
	"net/http"

	"github.com/EfosaE/credora-backend/domain/operation"
	custmiddleware "github.com/EfosaE/credora-backend/internal/middleware"
	"github.com/EfosaE/credora-backend/internal/response"
	operationsvc "github.com/EfosaE/credora-backend/service/operation"
)

// OperationHandler handles operations like internal transfers
type OperationHandler struct {
	operationService *operationsvc.OperationService
}

// NewOperationHandler creates a new handler
func NewOperationHandler(opService *operationsvc.OperationService) *OperationHandler {
	return &OperationHandler{operationService: opService}
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

    // Service call
    if err := h.operationService.InternalTransfer(r.Context(), req); err != nil {
        response.SendError(w, r, response.BadRequest(err, "transfer failed: "+err.Error()))
        return
    }

    response.SendSuccess(w, r, response.OK(nil, "Transfer successful"))
}

