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
	"github.com/EfosaE/credora-backend/internal/utils"

	"github.com/EfosaE/credora-backend/service"
	authsvc "github.com/EfosaE/credora-backend/service/auth"
	usersvc "github.com/EfosaE/credora-backend/service/user"
	"github.com/EfosaE/credora-backend/test"
	"github.com/EfosaE/credora-backend/test/fakes"
	"github.com/EfosaE/credora-backend/test/stubs"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateUserHandler_Success(t *testing.T) {
	// mockEmailAdapter := &fakes.MockEmailAdapter{
	// 	SendEmailFunc: func(ctx context.Context, req email.SendEmailRequest) error {
	// 		return nil
	// 	},
	// }
	mockUserRepo := &fakes.MockUserRepo{
		CreateFunc: func(ctx context.Context, req *user.CreateUserRequest) (*user.User, error) {
			return &user.User{
				ID:   uuid.New(),
				Name: req.Name,
			}, nil
		},
	}

	mockMonnifyRepo := &fakes.MockMonnifyRepo{
		CreateReservedAccountFunc: func(req *monnify.CreateCRAParams) (*monnify.CreateCRAResponse, error) {
			return stubs.StubCreateCRAResponse, nil
		},
	}

	log := test.SetupTestLogger()

	mockMonnifySvc := service.NewMonnifyService(mockMonnifyRepo, log)
	mockEventBus := &fakes.MockEventBus{}
	mockQueue := &fakes.MockQueueClient{}
	mockDeviceRepo := &fakes.MockDeviceTokenRepository{}

	userSvc := usersvc.NewUserService(mockUserRepo, log, mockEventBus, mockMonnifySvc, mockQueue, mockDeviceRepo)
	//This test doesnt really need the dependencies in authsvc hence nil
	authSvc := authsvc.NewAuthService(nil, nil, nil, nil, nil, nil, log)

	handler := NewAuthHandler(userSvc, authSvc)

	body := `{"name": "Efosa", "email": "efosa@example.com", "password": "password123", "nin":"35487696846", "phoneNumber":"09067353727"}`
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.RegisterUserHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusCreated, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	var resp map[string]any
	json.NewDecoder(res.Body).Decode(&resp)
	utils.PrintJSON(resp)

	assert.Equal(t, "The account details have been sent to your email, Login with it to verify your email", resp["message"])
	assert.NotNil(t, resp["data"])
}
