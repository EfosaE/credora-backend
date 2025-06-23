// handlers/user_handler.go
package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/EfosaE/credora-backend/domain/user"
	"github.com/EfosaE/credora-backend/internal/response"
	"github.com/EfosaE/credora-backend/internal/utils"
	"github.com/EfosaE/credora-backend/internal/validation"
	"github.com/EfosaE/credora-backend/service"
	"github.com/go-chi/render"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
// 	idStr := chi.URLParam(r, "id")
// 	id, err := strconv.ParseInt(idStr, 10, 64)
// 	if err != nil {
// 		http.Error(w, "invalid ID", http.StatusBadRequest)
// 		return
// 	}

// 	user, err := h.userService.GetUserByID(id)
// 	if err != nil {
// 		http.Error(w, "user not found", http.StatusNotFound)
// 		return
// 	}

// 	json.NewEncoder(w).Encode(user)
// }

func (h *UserHandler) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var req user.CreateUserRequest

	// Decode JSON
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		fmt.Println("Error decoding JSON:", err)
		response.SendError(w, r, response.BadRequest(err, "Invalid request body"))
		return
	}

	// Validate request
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
		response.SendError(w, r, response.InternalServerError(err, "could not create user"))
		return
	}

	response.SendSuccess(w, r, response.Created(user, "User created successfully"))
}
