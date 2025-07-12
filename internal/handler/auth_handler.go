// handlers/user_handler.go
package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/EfosaE/credora-backend/domain/user"
	"github.com/EfosaE/credora-backend/internal/config"
	"github.com/EfosaE/credora-backend/internal/pgerrors"
	"github.com/EfosaE/credora-backend/internal/response"
	"github.com/EfosaE/credora-backend/internal/utils"
	"github.com/EfosaE/credora-backend/internal/validation"

	authsvc "github.com/EfosaE/credora-backend/service/auth"
	usersvc "github.com/EfosaE/credora-backend/service/user"
	"github.com/go-chi/render"
)

type AuthHandler struct {
	userService *usersvc.UserService
	authService *authsvc.AuthService
}

func NewAuthHandler(userService *usersvc.UserService, authService *authsvc.AuthService) *AuthHandler {
	return &AuthHandler{userService: userService, authService: authService}
}

func (h *AuthHandler) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	var req user.CreateUserRequest

	// Decode JSON
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		fmt.Println("Error decoding JSON:", err)
		response.SendError(w, r, response.BadRequest(err, "Invalid request body"))
		return
	}

	// Validate request
	// safevalidate strcut because for some reason, a panic in the validation package crashes my server so bad chi cant catch it.
	if err := validation.SafeValidateStruct(validation.Validate, &req); err != nil {
		if strings.Contains(err.Error(), "internal validation error") {
			response.SendError(w, r, response.InternalServerError(err, err.Error()))
			return
		}
		errs := utils.ParseValidationErrors(err)
		response.SendError(w, r, response.BadRequest(errs, "Validation Failed"))
		return
	}

	// Call service
	user, err := h.userService.CreateUser(r.Context(), &req)
	if err != nil {
		if pgerrors.HandleUniqueViolation(w, r, err) {
			return // already responded
		}
		response.SendError(w, r, response.InternalServerError(err, "could not create user"))
		return
	}

	response.SendSuccess(w, r, response.Created(user, "User created successfully"))
}

func (h *AuthHandler) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	var req user.LoginUserRequest

	// Decode JSON
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		fmt.Println("Error decoding JSON:", err)
		response.SendError(w, r, response.BadRequest(err, "Invalid request body"))
		return
	}

	// Validate request
	// safevalidate strcut because for some reason, a panic in the validation package crashes my server so bad chi cant catch it.
	if err := validation.SafeValidateStruct(validation.Validate, &req); err != nil {
		if strings.Contains(err.Error(), "internal validation error") {
			response.SendError(w, r, response.InternalServerError(err, err.Error()))
			return
		}
		errs := utils.ParseValidationErrors(err)
		response.SendError(w, r, response.BadRequest(errs, "Validation Failed"))
		return
	}

	// Call service
	token, err := h.authService.Login(r.Context(), req.AccountNumber, req.Password)
	if err != nil {
		fmt.Println("Error during login:", err)
		response.SendError(w, r, response.BadRequest(nil, err.Error()))
		return
	}

	// Set JWT as HTTP-only cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   config.App.Env == "production", // Set to true in production with HTTPS
		MaxAge:   86400,
	})

	response.SendSuccess(w, r, response.OK(nil, "Login successful"))
}
