package handler

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/EfosaE/credora-backend/domain/auth"
	"github.com/EfosaE/credora-backend/domain/transaction"
	"github.com/EfosaE/credora-backend/domain/user"

	"github.com/EfosaE/credora-backend/internal/response"

	accountsvc "github.com/EfosaE/credora-backend/service/account"
	transactionsvc "github.com/EfosaE/credora-backend/service/transaction"
	usersvc "github.com/EfosaE/credora-backend/service/user"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"

	"github.com/google/uuid"
)

type UserHandler struct {
	userService *usersvc.UserService
	acctService *accountsvc.AccountService
	trxSvc      *transactionsvc.TransactionService
}

func NewUserHandler(userService *usersvc.UserService, acctService *accountsvc.AccountService, trxSvc *transactionsvc.TransactionService) *UserHandler {
	return &UserHandler{userService: userService, acctService: acctService, trxSvc: trxSvc}
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

func (h *UserHandler) GetTransactionHistoryHandler(w http.ResponseWriter, r *http.Request) {
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

	// Read cursor from query param
	cursorStr := r.URL.Query().Get("cursor")
	limitStr := r.URL.Query().Get("limit")

	var cursor *transaction.Cursor

	if cursorStr != "" {
		createdAt, id, err := DecodeCursor(cursorStr)
		if err != nil {
			response.SendError(w, r, response.BadRequest(
				errors.New("invalid cursor"),
				"Cursor format is invalid",
			))
			return
		}

		cursor = &transaction.Cursor{
			CreatedAt: createdAt,
			ID:        id,
		}
	}

	if limitStr == "" {
		limitStr = "50" // default limit
	}

	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil {
		response.SendError(w, r, response.BadRequest(
			errors.New("invalid limit"),
			"Limit must be a valid integer",
		))
		return
	}

	// Call service
	txns, nextCursorStruct, err := h.trxSvc.GetUserTransactions(
		r.Context(),
		userID,
		cursor,
		int32(limit),
	)
	if err != nil {
		fmt.Println("Error retrieving transaction history:", err)
		response.SendError(w, r, response.BadRequest(nil, err.Error()))
		return
	}

	// Encode next cursor before sending
	// var nextCursor string
	// if nextCursorStruct != nil {
	// 	nextCursor = EncodeCursor(
	// 		nextCursorStruct.CreatedAt,
	// 		nextCursorStruct.ID,
	// 	)
	// }

	var nextCursor *string

	if nextCursorStruct != nil {
		encoded := EncodeCursor(
			nextCursorStruct.CreatedAt,
			nextCursorStruct.ID,
		)
		nextCursor = &encoded
	}
	response.SendSuccess(
		w,
		r,
		response.OK(
			response.ObjKV(
				response.KV{Key: "transactions", Value: txns},
				response.KV{Key: "nextCursor", Value: nextCursor},
			),
			nil,
			fmt.Sprintf("%d records retrieved successfully", len(*txns)),
		),
	)
}

func (h *UserHandler) RegisterDeviceToken(w http.ResponseWriter, r *http.Request) {
	var req user.RegisterDeviceTokenRequest

	// Decode JSON body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendError(w, r, response.BadRequest(nil, "invalid request body"))
		return
	}

	// Basic validation
	if req.Token == "" || req.Platform == "" {
		response.SendError(w, r, response.BadRequest(nil, "token and platform are required"))
		return
	}

	_, claims, _ := jwtauth.FromContext(r.Context())

	userIDStr, ok := claims["userId"].(string)
	if !ok {
		response.SendError(w, r, response.Unauthorized("invalid userId in token"))
		return
	}

	userID, iErr := uuid.Parse(userIDStr)
	if iErr != nil {
		response.SendError(w, r, response.Unauthorized("invalid uuid format"))
		return
	}

	// Call service layer
	_, err := h.userService.RegisterDeviceTokenToUserID(
		r.Context(),
		userID,
		req.Token,
		req.Platform,
	)
	if err != nil {
		response.SendError(w, r, response.BadRequest(nil, err.Error()))
		return
	}

	response.SendSuccess(w, r, response.OK(
		nil,
		nil,
		"Device token registered successfully",
	))
}

func (h *UserHandler) GetUserByEmailHandler(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")
	if email == "" {
		response.SendError(w, r, response.BadRequest(nil, "email is required"))
		return
	}

	user, err := h.userService.GetUserByEmail(r.Context(), email)
	if err != nil {
		response.SendError(w, r, response.BadRequest(nil, err.Error()))
		return
	}

	response.SendSuccess(w, r, response.OK(
		response.Obj("user", user),
		nil,
		"User info retrieved successfully",
	))
}

// HELPERS
func EncodeCursor(createdAt time.Time, id int64) string {
	raw := fmt.Sprintf("%d|%d", createdAt.UnixNano(), id)
	return base64.StdEncoding.EncodeToString([]byte(raw))
}

func DecodeCursor(cursor string) (time.Time, int64, error) {
	bytes, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, 0, err
	}

	parts := strings.Split(string(bytes), "|")
	if len(parts) != 2 {
		return time.Time{}, 0, errors.New("invalid cursor format")
	}

	unixNano, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return time.Time{}, 0, err
	}

	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return time.Time{}, 0, err
	}

	return time.Unix(0, unixNano), id, nil
}
