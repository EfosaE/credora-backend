package custmiddleware

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	domainerr "github.com/EfosaE/credora-backend/domain/domianerrors"
	"github.com/EfosaE/credora-backend/domain/operation"
	"github.com/EfosaE/credora-backend/internal/response"
	"github.com/EfosaE/credora-backend/internal/utils"
	"github.com/EfosaE/credora-backend/internal/validation"
	accountsvc "github.com/EfosaE/credora-backend/service/account"
	idempotencysvc "github.com/EfosaE/credora-backend/service/idempotency"
	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ctxKey string

const ContextKeyInternalTransfer ctxKey = "internalTransferReq"

// genericInvalidRequestErr is returned whenever we want to avoid leaking
// whether an account exists or belongs to a different user.
var genericInvalidRequestErr = errors.New("invalid request")

func InternalTransferMiddleware(acctService accountsvc.AccountService,
	idemSvc idempotencysvc.IdempotencyService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// ---- Idempotency key presence check ----
			key := r.Header.Get("Idempotency-Key")
			if key == "" {
				response.SendError(w, r, response.BadRequest(
					errors.New("missing idempotency key"),
					"Generate a UUID and set it in the Idempotency-Key header",
				))
				return
			}

			// ---- JWT claims — safe extraction with nil/type guards ----
			_, claims, _ := jwtauth.FromContext(r.Context())

			raw, ok := claims["userId"]
			if !ok {
				response.SendError(w, r, response.Unauthorized(
					domainerr.ErrInvalidCredentials.Error(),
				))
				return
			}

			userIdStr, ok := raw.(string)
			if !ok {
				response.SendError(w, r, response.Unauthorized(
					domainerr.ErrInvalidCredentials.Error(),
				))
				return
			}

			userID, err := uuid.Parse(userIdStr)
			if err != nil {
				response.SendError(w, r, response.BadRequest(err, "Invalid user ID"))
				return
			}

			// ---- Scope idempotency key to the authenticated user ----
			// Prevents cross-user key collisions and activity probing.
			scopedKey := fmt.Sprintf("%s:%s", userIdStr, key)

			// ---- Idempotency check — error is no longer silently swallowed ----
			existing, err := idemSvc.GetRecord(r.Context(), scopedKey)
			if err != nil {
				fmt.Println(err)
				response.SendError(w, r, response.InternalServerError(
					errors.New("unable to process request"),
					"An internal error occurred",
				))
				return
			}
			if existing != nil {
				response.SendSuccess(w, r, response.OK(
					response.Obj("status", existing),
					nil,
					"This request has already been processed",
				))
				return
			}

			// ---- Read and restore body ----
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				response.SendError(w, r, response.BadRequest(err, "Failed to read request body"))
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

			// ---- Decode JSON into DTO ----
			var dto operation.InternalTransferDTO
			if err := json.Unmarshal(bodyBytes, &dto); err != nil {
				response.SendError(w, r, response.BadRequest(err, "Invalid JSON"))
				return
			}

			// ---- 1️⃣ Struct validation ----
			if err := validation.SafeValidateStruct(validation.Validate, &dto); err != nil {
				if strings.Contains(err.Error(), "internal validation error") {
					response.SendError(w, r, response.InternalServerError(err, "An internal error occurred"))
					return
				}
				errs := validation.ParseValidationErrors(err)
				response.SendError(w, r, response.BadRequest(errs, "Validation failed"))
				return
			}
			// if err := validation.SafeValidateStruct(validation.Validate, &dto); err != nil {
			// 	errs := validation.ParseValidationErrors(err)
			// 	response.SendError(w, r, response.BadRequest(errs, "Validation failed"))
			// 	return
			// }

			utils.PrintJSON(dto)

			// ---- 2️⃣ Convert DTO → Domain ----
			domainReq, err := dto.ToDomain()
			if err != nil {
				response.SendError(w, r, response.BadRequest(err, "Invalid transfer fields"))
				return
			}

			domainReq.IdempotencyKey = scopedKey // store the scoped key, not the raw header value

			// ---- 3️⃣ Domain validations (pure, no DB) ----
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

			// ---- 4️⃣ DB validations — single query for sender (ownership + balance) ----
			//
			// FindAccountByOwner replaces the previous two separate calls:
			//   acctService.AccountBelongsToUser(...)   — ownership check
			//   acctService.FindAccountByAcctNum(...)   — fetch balance
			//
			// Both operations are collapsed into one DB round-trip.
			// On "not found" OR "wrong owner" we return the same generic error
			// to prevent timing/enumeration attacks.
			// utils.PrintJSON(domainReq)
			fromAcct, err := acctService.FindAccountByOwner(r.Context(), userID, domainReq.FromAcctNum)
			if err != nil {
				if errors.Is(err, domainerr.ErrAccountNotFound) || errors.Is(err, domainerr.ErrAccountNotActivated) {
					// Uniform response — attacker cannot distinguish "doesn't exist"
					// from "belongs to someone else".
					response.SendError(w, r, response.BadRequest(
						genericInvalidRequestErr,
						"Invalid request",
					))
					return
				}
				response.SendError(w, r, response.InternalServerError(
					errors.New("unable to process request"),
					"An internal error occurred",
				))
				return
			}
			utils.PrintJSON(fromAcct)
			// ---- 5️⃣ Fetch receiver account (retained for downstream use) ----
			_, tErr := acctService.FindAccountByAcctNum(r.Context(), domainReq.ToAcctNum)
			if tErr != nil {
				// Use the same generic message to avoid leaking account existence.
				response.SendError(w, r, response.BadRequest(
					genericInvalidRequestErr,
					"Invalid request",
				))
				return
			}

			// ---- 6️⃣ Balance check ----
			if fromAcct.Balance.LessThan(domainReq.Amount) {
				response.SendError(w, r, response.ValidationError(operation.ErrInsufficientFunds.Error()))
				return
			}

			// ---- 7️⃣ Inject domain request + resolved accounts into context ----
			// Passing both accounts avoids redundant DB fetches in the handler.
			ctx := context.WithValue(r.Context(), ContextKeyInternalTransfer, domainReq)
			// ctx = context.WithValue(ctx, ContextKeyFromAccount, fromAcct)
			// ctx = context.WithValue(ctx, ContextKeyToAccount, toAcct)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
