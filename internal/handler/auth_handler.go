// handlers/user_handler.go
package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/EfosaE/credora-backend/domain/user"
	"github.com/EfosaE/credora-backend/internal/pgerrors"
	"github.com/EfosaE/credora-backend/internal/response"
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

	response.SendSuccess(
		w,
		r,
		response.Created(
			response.Obj("user", user),
			nil,
			"User created successfully",
		),
	)

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
		errs := validation.ParseValidationErrors(err)
		response.SendError(w, r, response.BadRequest(errs, "Validation Failed"))
		return
	}

	// Call service
	user, token, err := h.authService.Login(r.Context(), req.AccountNumber, req.Password)
	if err != nil {
		fmt.Println("Error during login:", err)
		response.SendError(w, r, response.BadRequest(nil, err.Error()))
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

	response.SendSuccess(
		w,
		r,
		response.OK(
			// Return an object with multiple key-value pairs using ObjKV e.g {"name": "Alice", "age": 30} nside the data field of the response
			response.ObjKV(response.KV{Key: "accessToken", Value: token}, response.KV{Key: "user", Value: user}),
			nil,
			"Login successful",
		),
	)

}

func (h *AuthHandler) ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req user.ResetPasswordRequest

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
	var req user.ValidatePasswordRequest

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
