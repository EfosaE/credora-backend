// handler/user_test.go
package handler

// import (
// 	"context"
// 	"encoding/json"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/EfosaE/credora-backend/domain/user"
// 	"github.com/EfosaE/credora-backend/test/mocks"
// )

// func TestCreateUserHandler_Success(t *testing.T) {
// 	mockService := &mocks.MockUserService{
// 		CreateUserFunc: func(ctx context.Context, req *user.CreateUserRequest) (*user.User, error) {
// 			return &user.User{
// 				ID:    "user-id-123",
// 				Name:  req.Name,
// 				Email: req.Email,
// 			}, nil
// 		},
// 	}

// 	handler := handler.NewUserHandler(mockService)

// 	body := `{"name": "Efosa", "email": "efosa@example.com", "password": "password123"}`
// 	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
// 	req.Header.Set("Content-Type", "application/json")
// 	w := httptest.NewRecorder()

// 	handler.CreateUserHandler(w, req)

// 	res := w.Result()
// 	defer res.Body.Close()

// 	assert.Equal(t, http.StatusCreated, res.StatusCode)
// 	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

// 	var resp map[string]interface{}
// 	json.NewDecoder(res.Body).Decode(&resp)

// 	assert.Equal(t, "User created successfully", resp["message"])
// 	assert.NotNil(t, resp["data"])
// }
