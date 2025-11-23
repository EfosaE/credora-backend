package custmiddleware

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/EfosaE/credora-backend/domain/operation"
	"github.com/EfosaE/credora-backend/internal/response"
	"github.com/EfosaE/credora-backend/internal/validation"
	accountsvc "github.com/EfosaE/credora-backend/service/account"
	idempotencysvc "github.com/EfosaE/credora-backend/service/idempotency"
	"github.com/go-chi/jwtauth/v5"
	"github.com/shopspring/decimal"
	"io"
	"net/http"
)

type ctxKey string

const ContextKeyInternalTransfer ctxKey = "internalTransferReq"

func InternalTransferMiddleware(acctService accountsvc.AccountService,
	idemSvc idempotencysvc.IdempotencyService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("Idempotency-Key")
			if key == "" {
				response.SendError(w, r, response.BadRequest(
					errors.New("missing idempotency key"),
					"Generate a UUID and set it in the Idempotency-Key header",
				))
				return
			}

			// Check if request was already processed
			if existing, _ := idemSvc.GetRecord(r.Context(), key); existing != nil {
				// Return the saved response directly
				response.SendSuccess(w, r, response.OK(existing, "This request has already been processed"))
				return
			}

			// ---- 2️⃣ Extract user from JWT ----
			_, claims, _ := jwtauth.FromContext(r.Context())
			acctNum := claims["account_number"].(string)

			// Read the body once
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				response.SendError(w, r, response.BadRequest(err, "Failed to read request body"))
				return
			}

			// Restore body for future reading
			r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

			// Decode JSON into DTO
			var dto operation.InternalTransferDTO
			if err := json.Unmarshal(bodyBytes, &dto); err != nil {
				response.SendError(w, r, response.BadRequest(err, "Invalid JSON"))
				return
			}

			// ---- 1️⃣ Struct validation ----
			if err := validation.SafeValidateStruct(validation.Validate, &dto); err != nil {
				errs := validation.ParseValidationErrors(err)
				response.SendError(w, r, response.BadRequest(errs, "Validation failed"))
				return
			}

			// ---- 2️⃣ Convert DTO → DOMAIN ----
			domainReq, err := dto.ToDomain(acctNum)
			if err != nil {
				response.SendError(w, r, response.BadRequest(err, "Invalid transfer fields"))
				return
			}

			domainReq.IdempotencyKey = key

			// ---- 3️⃣ Domain validations ----
			if domainReq.FromAcctNum == domainReq.ToAcctNum {
				response.SendError(w, r, response.BadRequest(
					operation.ErrInvalidTransfer,
					"Sender and receiver accounts cannot be the same",
				))
				return
			}

			if domainReq.Amount.LessThanOrEqual(decimal.Zero) {
				response.SendError(w, r, response.BadRequest(
					operation.ErrInvalidAmount,
					"Amount must be greater than zero",
				))
				return
			}

			// ---- 8️⃣ Lightweight DB validations (read-only!) ----
			fromAcct, err := acctService.FindAccountByAcctNum(r.Context(), domainReq.FromAcctNum)
			if err != nil {
				response.SendError(w, r, response.NotFound("Sender account not found"))
				return
			}

			if _, err := acctService.FindAccountByAcctNum(r.Context(), domainReq.ToAcctNum); err != nil {
				response.SendError(w, r, response.NotFound("Receiver account not found"))
				return
			}

			if fromAcct.Balance.LessThan(domainReq.Amount) {
				response.SendError(w, r, response.ValidationError(operation.ErrInsufficientFunds.Error()))
				return
			}

			// ---- 4️⃣ Inject domain request into context ----
			ctx := context.WithValue(r.Context(), ContextKeyInternalTransfer, domainReq)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
