package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/EfosaE/credora-backend/domain/simulator"
	"github.com/EfosaE/credora-backend/internal/response"
	"github.com/EfosaE/credora-backend/internal/utils"
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
		errs := utils.ParseValidationErrors(err)
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

	response.SendSuccess(w, r, response.OK(nil, "Simulated transfer successfully triggered."))
}

// func (h *AuthHandler) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
// 	var req user.LoginUserRequest

// 	// Decode JSON
// 	if err := render.DecodeJSON(r.Body, &req); err != nil {
// 		fmt.Println("Error decoding JSON:", err)
// 		response.SendError(w, r, response.BadRequest(err, "Invalid request body"))
// 		return
// 	}

// 	// Validate request
// 	// safevalidate strcut because for some reason, a panic in the validation package crashes my server so bad chi cant catch it.
// 	if err := validation.SafeValidateStruct(validation.Validate, &req); err != nil {
// 		if strings.Contains(err.Error(), "internal validation error") {
// 			response.SendError(w, r, response.InternalServerError(err, err.Error()))
// 			return
// 		}
// 		errs := utils.ParseValidationErrors(err)
// 		response.SendError(w, r, response.BadRequest(errs, "Validation Failed"))
// 		return
// 	}

// 	// Call service
// 	token, err := h.authService.Login(r.Context(), req.AccountNumber, req.Password)
// 	if err != nil {
// 		fmt.Println("Error during login:", err)
// 		response.SendError(w, r, response.BadRequest(nil, err.Error()))
// 		return
// 	}

// 	// Set JWT as HTTP-only cookie
// 	http.SetCookie(w, &http.Cookie{
// 		Name:     "jwt",
// 		Value:    token,
// 		Path:     "/",
// 		HttpOnly: true,
// 		SameSite: http.SameSiteLaxMode,
// 		Secure:   config.App.Env == "production", // Set to true in production with HTTPS
// 		MaxAge:   86400,
// 	})

// 	response.SendSuccess(w, r, response.OK(nil, "Login successful"))
// }
