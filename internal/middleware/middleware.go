package custmiddleware

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"

	"encoding/json"

	"github.com/EfosaE/credora-backend/domain/operation"
	"github.com/EfosaE/credora-backend/infrastructure"
	"github.com/EfosaE/credora-backend/internal/response"
	"github.com/EfosaE/credora-backend/internal/validation"
	"github.com/shopspring/decimal"
)

type ctxKey string

const ContextKeyInternalTransfer ctxKey = "internalTransferReq"

func IdempotencyMiddleware(cache *infrastructure.IdempotencyCache) func(next http.Handler) http.Handler {
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

            // Read body once
            bodyBytes, err := io.ReadAll(r.Body)
            if err != nil {
                response.SendError(w, r, response.BadRequest(err, "Failed to read request body"))
                return
            }

            // Restore for future reading (if needed)
            r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

            // Decode into DTO (string fields)
            var dto operation.InternalTransferDTO
            if err := json.Unmarshal(bodyBytes, &dto); err != nil {
                response.SendError(w, r, response.BadRequest(err, "Invalid JSON"))
                return
            }

            // ---- 1️⃣ STRUCT VALIDATION ----
            if err := validation.SafeValidateStruct(validation.Validate, &dto); err != nil {
                errs := validation.ParseValidationErrors(err)
                response.SendError(w, r, response.BadRequest(errs, "Validation Failed"))
                return
            }

            // ---- 2️⃣ Convert DTO → DOMAIN ----
            domainReq, err := dto.ToDomain()
            if err != nil {
                response.SendError(w, r, response.BadRequest(err, "Invalid transfer fields"))
                return
            }

            domainReq.IdempotencyKey = key

            // ---- 3️⃣ DOMAIN VALIDATION ----
            if domainReq.FromAcctNum == domainReq.ToAcctNum {
                response.SendError(w, r, response.BadRequest(operation.ErrInvalidTransfer,
                    "Sender and receiver account cannot be the same"))
                return
            }

            if domainReq.Amount.LessThanOrEqual(decimal.Zero) {
                response.SendError(w, r, response.BadRequest(operation.ErrInvalidAmount,
                    "Amount must be greater than zero"))
                return
            }

            // ---- 4️⃣ Idempotency Check ----
            allowed, err := cache.TryAcquire(r.Context(), key)
            if err != nil {
                response.SendError(w, r, response.InternalServerError(err, "Idempotency check failed"))
                return
            }

            if !allowed {
                response.SendError(w, r, response.Conflict(operation.ErrDuplicateRequest,
                    "This request has already been processed"))
                return
            }

            // ---- 5️⃣ Inject into context ----
            ctx := context.WithValue(r.Context(), ContextKeyInternalTransfer, domainReq)
            r = r.WithContext(ctx)

            // ---- 6️⃣ Call handler ----
            next.ServeHTTP(w, r)
        })
    }
}

