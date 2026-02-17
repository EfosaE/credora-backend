package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/EfosaE/credora-backend/domain/simulator"
	"github.com/EfosaE/credora-backend/internal/response"
	"github.com/EfosaE/credora-backend/internal/validation"
	"github.com/EfosaE/credora-backend/service"
	"github.com/go-chi/render"
)

type SimulatorHandler struct {
	simSvc *service.SimulatorService
}

func NewSimulatorHandler(simSvc *service.SimulatorService) *SimulatorHandler {
	return &SimulatorHandler{simSvc: simSvc}
}

func (h *SimulatorHandler) SimulateTransferExt(w http.ResponseWriter, r *http.Request) {
	var req simulator.TransferRequest

	// Decode JSON body
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		fmt.Println("Error decoding JSON:", err)
		response.SendError(w, r, response.BadRequest(err, "Unable to parse request body. Please ensure it's valid JSON."))
		return
	}

	// Validate request
	if err := validation.SafeValidateStruct(validation.Validate, &req); err != nil {
		if strings.Contains(err.Error(), "internal validation error") {
			response.SendError(w, r, response.InternalServerError(err, "An internal validation error occurred. Please try again later."))
			return
		}
		errs := validation.ParseValidationErrors(err)
		response.SendError(w, r, response.BadRequest(errs, "Input validation failed. Please check your fields."))
		return
	}

	// Simulate external transfer
	err := h.simSvc.SendMoney(r.Context(), &req)
	if err != nil {
		fmt.Println("Error simulating external transfer:", err)
		response.SendError(w, r, response.InternalServerError(err, "An error occurred while simulating the transfer."))
		return
	}

	response.SendSuccess(w, r, response.OK(
		nil,
		nil,
		"Simulated transfer successfully triggered.",
	))
}
