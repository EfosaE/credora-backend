package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/EfosaE/credora-backend/domain/auth"

	"github.com/EfosaE/credora-backend/internal/response"

	accountsvc "github.com/EfosaE/credora-backend/service/account"
	usersvc "github.com/EfosaE/credora-backend/service/user"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"

	"github.com/google/uuid"
)

type UserHandler struct {
	userService *usersvc.UserService
	acctService *accountsvc.AccountService
}

func NewUserHandler(userService *usersvc.UserService, acctService *accountsvc.AccountService) *UserHandler {
	return &UserHandler{userService: userService, acctService: acctService}
}

func (h *UserHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())

	userIDStr, ok := claims["userId"].(string)
	if !ok {
		http.Error(w, "invalid userId in token", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "invalid UUID format", http.StatusUnauthorized)
		return
	}

	user := auth.TokenPayload{
		UserID:        userID,
		Name:          claims["name"].(string),
		AccountNumber: claims["accountNumber"].(string),
	}

	response.SendSuccess(w, r, response.OK(
		response.Obj("user", user),
		nil,
		"User info retrieved successfully",
	))
}

func (h *UserHandler) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())

	userAcctNum, ok := claims["accountNumber"].(string)
	if !ok {
		http.Error(w, "invalid account number in token", http.StatusUnauthorized)
		return
	}

	// userID, err := uuid.Parse(userIDStr)
	// if err != nil {
	// 	http.Error(w, "invalid UUID format", http.StatusUnauthorized)
	// 	return
	// }

	user, err := h.acctService.FindUserByAccount(r.Context(), userAcctNum)

	if err != nil {
		fmt.Println("Error retrieving user balance:", err)
		response.SendError(w, r, response.BadRequest(nil, err.Error()))
		return
	}

	response.SendSuccess(w, r, response.OK(
		response.Obj("user", user),
		nil,
		"User info retrieved successfully",
	))
}

func (h *UserHandler) GetRecipientName(w http.ResponseWriter, r *http.Request) {
	acctNum := chi.URLParam(r, "acctNum")
	if acctNum == "" {
		response.SendError(w, r, response.BadRequest(
			errors.New("missing account number"),
			"Account number is required",
		))
		return
	}

	// Call service
	// user, token, err := h.authService.Login(r.Context(), req.AccountNumber, req.Password)
	user, err := h.acctService.FindUserByAccount(r.Context(), acctNum)
	if err != nil {
		fmt.Println("Error retrieving recipient name:", err)
		response.SendError(w, r, response.BadRequest(nil, err.Error()))
		return
	}

	response.SendSuccess(
		w,
		r,
		response.OK(
			response.ObjKV(response.KV{Key: "user", Value: user}),
			nil,
			"Recipient name retrieved successfully",
		),
	)

}
