// // handler/user_test.go
package handler

// import (
// 	"context"
// 	"encoding/json"
// 	"net/http"
// 	"net/http/httptest"
// 	"strings"
// 	"testing"

// 	"github.com/EfosaE/credora-backend/domain/email"
// 	"github.com/EfosaE/credora-backend/domain/monnify"
// 	"github.com/EfosaE/credora-backend/domain/user"
// 	"github.com/EfosaE/credora-backend/service"
// 	usersvc "github.com/EfosaE/credora-backend/service/user"
// 	"github.com/EfosaE/credora-backend/test"
// 	"github.com/EfosaE/credora-backend/test/mocks"
// 	"github.com/google/uuid"
// 	"github.com/stretchr/testify/assert"
// )

// func TestCreateUserHandler_Success(t *testing.T) {
// 	mockEmailAdapter := &mocks.MockEmailAdapter{
// 		SendEmailFunc: func(ctx context.Context, req email.SendEmailRequest) error {
// 			return nil
// 		},
// 	}
// 	mockUserRepo := &mocks.MockUserRepo{
// 		CreateFunc: func(ctx context.Context, req *user.CreateUserRequest) (*user.User, error) {
// 			return &user.User{
// 				ID:   uuid.New(),
// 				Name: req.Name,
// 			}, nil
// 		},
// 	}

// 	mockMonnifyRepo := &mocks.MockMonnifyRepo{
// 		CreateReservedAccountFunc: func(req *monnify.CreateCRAParams) (*monnify.CreateCRAResponse, error) {
// 			return &monnify.CreateCRAResponse{
// 				RequestSuccessful: true,
// 				ResponseMessage:   "Account created successfully",
// 				ResponseCode:      "0",
// 				ResponseBody: monnify.CreateCRAResponseBody{
// 					ContractCode:          "100693167467",
// 					AccountReference:      "REF123",
// 					AccountName:           "John Doe",
// 					CurrencyCode:          "NGN",
// 					CustomerEmail:         "john@example.com",
// 					CustomerName:          "John Doe",
// 					CollectionChannel:     "RESERVED_ACCOUNT",
// 					ReservationReference:  "ABC123456789",
// 					ReservedAccountType:   "GENERAL",
// 					Status:                "ACTIVE",
// 					CreatedOn:             "2024-11-25 07:35:17.566",
// 					Nin:                   "21212121212",
// 					RestrictPaymentSource: false,
// 					Accounts: []monnify.ReservedAccount{
// 						{
// 							BankCode:      "50515",
// 							BankName:      "Moniepoint Microfinance Bank",
// 							AccountNumber: "6839490147",
// 							AccountName:   "John Doe",
// 						},
// 					},
// 					IncomeSplitConfig: []monnify.IncomeSplitConfig{},
// 				},
// 			}, nil
// 		},
// 	}

// 	log := test.SetupTestLogger()
// 	mockMonnifySvc := service.NewMonnifyService(mockMonnifyRepo, log)
// 	mockEmailSvc := service.NewEmailService(mockEmailAdapter)
// 	// monnifyClient := test.SetupTestMonnifyClient()
// 	service := usersvc.NewUserService(mockUserRepo, log, mockMonnifySvc, mockEmailSvc)

// 	handler := NewUserHandler(service)

// 	body := `{"name": "Efosa", "email": "efosa@example.com", "password": "password123"}`
// 	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
// 	req.Header.Set("Content-Type", "application/json")
// 	w := httptest.NewRecorder()

// 	handler.CreateUserHandler(w, req)

// 	res := w.Result()
// 	defer res.Body.Close()

// 	assert.Equal(t, http.StatusCreated, res.StatusCode)
// 	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

// 	var resp map[string]any
// 	json.NewDecoder(res.Body).Decode(&resp)

// 	assert.Equal(t, "User created successfully", resp["message"])
// 	assert.NotNil(t, resp["data"])
// }
