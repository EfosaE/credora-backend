package handler

import (
	"net/http"

	"github.com/EfosaE/credora-backend/domain/auth"
	"github.com/EfosaE/credora-backend/internal/response"
	usersvc "github.com/EfosaE/credora-backend/service/user"
	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
)

type UserHandler struct {
	userService *usersvc.UserService
}

func NewUserHandler(userService *usersvc.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		http.Error(w, "invalid user_id in token", http.StatusUnauthorized)
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
		AccountNumber: claims["account_number"].(string),
	}

	response.SendSuccess(w, r, response.OK(user, "User info retrieved successfully"))
}
