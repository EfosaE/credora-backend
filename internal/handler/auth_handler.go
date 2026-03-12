// handlers/user_handler.go
package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/EfosaE/credora-backend/domain/auth"
	domainerr "github.com/EfosaE/credora-backend/domain/domianerrors"
	"github.com/EfosaE/credora-backend/domain/user"
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
		errs := validation.ParseValidationErrors(err)
		response.SendError(w, r, response.BadRequest(errs, "Validation Failed"))
		return
	}

	// Call service
	user, err := h.userService.CreateUser(r.Context(), &req)
	if err != nil {
		if pgerrors.HandlePGError(w, r, err) {
			return // already responded
		}
		response.SendError(w, r, response.InternalServerError(err, "could not create user"))
		return
	}

	data, _ := utils.StructToMap(user)
	response.SendSuccess(
		w,
		r,
		response.Created(
			data,
			nil,
			"The account details have been sent to your email, Login with it to verify your email",
		),
	)

}

func (h *AuthHandler) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	var req auth.LoginUserRequest

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
		errs := validation.ParseValidationErrors(err)
		response.SendError(w, r, response.BadRequest(errs, "Validation Failed"))
		return
	}

	// Call service
	user, token, err := h.authService.Login(r.Context(), req.Identifier, req.Password)
	if err != nil {

		switch err {

		case domainerr.ErrInvalidCredentials:
			response.SendError(
				w,
				r,
				response.Unauthorized(err.Error()),
			)

		case domainerr.ErrAccountNotActivated:
			response.SendError(
				w,
				r,
				response.Forbidden(err.Error()),
			)

		default:
			response.SendError(
				w,
				r,
				response.InternalServerError(err, "Login failed"),
			)
		}

		return
	}

	// // Set JWT as HTTP-only cookie
	// http.SetCookie(w, &http.Cookie{
	// 	Name:     "jwt",
	// 	Value:    token,
	// 	Path:     "/",
	// 	HttpOnly: true,
	// 	SameSite: http.SameSiteLaxMode,
	// 	Secure:   config.App.Env == "production", // Set to true in production with HTTPS
	// 	MaxAge:   86400,
	// })

	data, err := utils.StructToMap(auth.LoginResponse{
		AccessToken: token,
		User:        *user,
	})
	if err != nil {
		response.SendError(w, r, response.InternalServerError(err, err.Error()))
		return
	}

	response.SendSuccess(w, r, response.OK(data, nil, "Login successful"))

}

func (h *AuthHandler) ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req auth.ResetPasswordRequest

	// Decode JSON
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		response.SendError(w, r, response.BadRequest(err, "Invalid request body"))
		return
	}

	// Validate request
	if err := validation.SafeValidateStruct(validation.Validate, &req); err != nil {
		if strings.Contains(err.Error(), "internal validation error") {
			response.SendError(w, r, response.InternalServerError(err, err.Error()))
			return
		}
		errs := validation.ParseValidationErrors(err)
		response.SendError(w, r, response.BadRequest(errs, "Validation failed"))
		return
	}

	// Call service
	if err := h.authService.RequestPasswordReset(r.Context(), req.Email); err != nil {
		fmt.Println("Error during password reset:", err)
		response.SendError(w, r, response.InternalServerError(err, "Failed to process password reset request"))
		return
	}

	// Always return the same response
	response.SendSuccess(
		w,
		r,
		response.OK(
			nil,
			nil,
			"A password reset link has been sent to your email.",
		),
	)
}

func (h *AuthHandler) ValidatePasswordRequestHandler(w http.ResponseWriter, r *http.Request) {
	var req auth.ValidatePasswordRequest

	// Decode JSON
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		response.SendError(w, r, response.BadRequest(err, "Invalid request body"))
		return
	}

	// Validate request
	if err := validation.SafeValidateStruct(validation.Validate, &req); err != nil {
		if strings.Contains(err.Error(), "internal validation error") {
			response.SendError(w, r, response.InternalServerError(err, err.Error()))
			return
		}
		errs := validation.ParseValidationErrors(err)
		response.SendError(w, r, response.BadRequest(errs, "Validation failed"))
		return
	}

	// Call service
	if err := h.authService.ValidatePasswordResetRequest(r.Context(), req.Email, req.PasswordResetToken, req.NewPassword); err != nil {
		fmt.Println("Error during password reset validation:", err)
		response.SendError(w, r, response.BadRequest(err, "Failed to validate password reset request"))
		return
	}

	// if err := h.authService.RequestPasswordReset(r.Context(), req.Email); err != nil {
	// 	fmt.Println("Error during password reset:", err)
	// 	response.SendError(w, r, response.InternalServerError(err, "Failed to process password reset request"))
	// 	return
	// }

	response.SendSuccess(
		w,
		r,
		response.OK(
			nil,
			nil,
			"Your password has been reset, login with your new password.",
		),
	)
}
