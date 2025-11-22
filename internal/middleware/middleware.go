package custmiddleware

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/EfosaE/credora-backend/domain/operation"
	"github.com/EfosaE/credora-backend/internal/response"
	"github.com/EfosaE/credora-backend/internal/validation"
	"github.com/shopspring/decimal"
)

type ctxKey string

const ContextKeyInternalTransfer ctxKey = "internalTransferReq"

func IdempotencyMiddleware() func(next http.Handler) http.Handler {
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
			domainReq, err := dto.ToDomain()
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

			// ---- 4️⃣ Inject domain request into context ----
			ctx := context.WithValue(r.Context(), ContextKeyInternalTransfer, domainReq)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
