// File: handler/operations_handler.go

package handler

import (
	"encoding/json"
	"net/http"

	"github.com/EfosaE/credora-backend/domain/operation"
	"github.com/EfosaE/credora-backend/internal/response"
	"github.com/EfosaE/credora-backend/internal/validation"
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
	var dto operation.InternalTransferDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate DTO
	if err := validation.SafeValidateStruct(validation.Validate, &dto); err != nil {
		errs := validation.ParseValidationErrors(err)
		response.SendError(w, r, response.BadRequest(errs, "Validation Failed"))
		return
	}
	// Build operation request
	// Convert
	req, err := dto.ToDomain()
	if err != nil {
		response.SendError(w, r, response.BadRequest(err, "Invalid transfer data"))
		return
	}

	// Call service
	if err := h.operationService.InternalTransfer(r.Context(), req); err != nil {
		response.SendError(w, r, response.BadRequest(err, "transfer failed: "+err.Error()))
		return
	}

	response.SendSuccess(w, r, response.OK(nil, "Transfer successful"))
}
