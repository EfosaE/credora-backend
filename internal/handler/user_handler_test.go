// handler/user_test.go
package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/EfosaE/credora-backend/domain/user"
	"github.com/EfosaE/credora-backend/domain/monnify"
	"github.com/EfosaE/credora-backend/service"
	"github.com/EfosaE/credora-backend/test"
	"github.com/EfosaE/credora-backend/test/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateUserHandler_Success(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepo{
		CreateFunc: func(ctx context.Context, req *user.CreateUserRequest) (*user.User, error) {
			return &user.User{
				ID:   uuid.New(),
				Name: req.Name,
			}, nil
		},
	}

	mockMonnifyRepo := &mocks.MockMonnifyRepo{
	CreateReservedAccountFunc: func(req *monnify.CreateCustomerRequest) (*monnify.CreateCustomerResponse, error) {
		return &monnify.CreateCustomerResponse{
			RequestSuccessful: true,
			ResponseMessage:   "Account created successfully",
			ResponseBody: monnify.CustomerResponseBody{
				AccountReference: "REF123",
				AccountName:      "John Doe",
				AccountNumber:    "1234567890",
				BankName:         "Wema Bank",
			},
		}, nil
	},
}

	log := test.SetupTestLogger()
	mockMonnifySvc := service.NewMonnifyService(mockMonnifyRepo, log)
	// monnifyClient := test.SetupTestMonnifyClient()
	service := service.NewUserService(mockUserRepo, log, mockMonnifySvc)

	handler := NewUserHandler(service)

	body := `{"name": "Efosa", "email": "efosa@example.com", "password": "password123"}`
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateUserHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusCreated, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	var resp map[string]interface{}
	json.NewDecoder(res.Body).Decode(&resp)

	assert.Equal(t, "User created successfully", resp["message"])
	assert.NotNil(t, resp["data"])
}
