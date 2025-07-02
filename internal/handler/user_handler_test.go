// // handler/user_test.go
package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	// "github.com/EfosaE/credora-backend/domain/email"
	"github.com/EfosaE/credora-backend/domain/monnify"
	"github.com/EfosaE/credora-backend/domain/user"
	"github.com/EfosaE/credora-backend/service"
	usersvc "github.com/EfosaE/credora-backend/service/user"
	"github.com/EfosaE/credora-backend/test"
	"github.com/EfosaE/credora-backend/test/mocks"
	"github.com/EfosaE/credora-backend/test/stubs"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateUserHandler_Success(t *testing.T) {
	// mockEmailAdapter := &mocks.MockEmailAdapter{
	// 	SendEmailFunc: func(ctx context.Context, req email.SendEmailRequest) error {
	// 		return nil
	// 	},
	// }
	mockUserRepo := &mocks.MockUserRepo{
		CreateFunc: func(ctx context.Context, req *user.CreateUserRequest) (*user.User, error) {
			return &user.User{
				ID:   uuid.New(),
				Name: req.Name,
			}, nil
		},
	}

	mockMonnifyRepo := &mocks.MockMonnifyRepo{
		CreateReservedAccountFunc: func(req *monnify.CreateCRAParams) (*monnify.CreateCRAResponse, error) {
			return stubs.StubCreateCRAResponse, nil
		},
	}

	log := test.SetupTestLogger()
	mockMonnifySvc := service.NewMonnifyService(mockMonnifyRepo, log)
	mockEventBus := &mocks.MockEventBus{}
	// mockEmailSvc := service.NewEmailService(mockEmailAdapter)
	// monnifyClient := test.SetupTestMonnifyClient()
	service := usersvc.NewUserService(mockUserRepo, log, mockEventBus, mockMonnifySvc)

	handler := NewUserHandler(service)

	body := `{"name": "Efosa", "email": "efosa@example.com", "password": "password123", "nin":"35487696846", "phone_number":"09067353727"}`
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateUserHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusCreated, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	var resp map[string]any
	json.NewDecoder(res.Body).Decode(&resp)

	assert.Equal(t, "User created successfully", resp["message"])
	assert.NotNil(t, resp["data"])
}
