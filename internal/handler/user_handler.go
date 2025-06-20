// handlers/user_handler.go
package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/EfosaE/credora-backend/domain/user"
	"github.com/EfosaE/credora-backend/service"
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

    // Decode JSON body into the struct
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Optionally: log the struct
    fmt.Printf("%+v\n", req)

    // Respond
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"message": "User created successfully"})
}

