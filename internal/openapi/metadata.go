// internal/openapi/metadata.go
package openapi

import (
	"github.com/EfosaE/credora-backend/domain/auth"
	"github.com/EfosaE/credora-backend/domain/idempotency"
	"github.com/EfosaE/credora-backend/domain/operation"
	"github.com/EfosaE/credora-backend/domain/user"
	"github.com/EfosaE/credora-backend/internal/handler"
)

func RegisterOpenAPIRoutesMetaData() {

	// -- Health --
	register(&Doc{
		Summary:     "Liveness check",
		Description: "Returns 200 if the server is alive",
		Response:    handler.HealthData{},
	}, "GET", "/health/liveness")

	// -- Auth --
	register(&Doc{
		Summary:     "Register user",
		Description: "Creates a new user account",
		Request:     user.CreateUserRequest{},
		Response:    user.CreateUserResponse{},
	}, "POST", "/auth/register")

	register(&Doc{
		Summary:     "Login",
		Description: "Authenticates a user and returns a JWT token",
		Request:     auth.LoginUserRequest{},
		Response:    auth.LoginResponse{},
	}, "POST", "/auth/login")

	register(&Doc{
		Summary:     "Request password reset",
		Description: "Sends a password reset link to the user's email",
		Request:     auth.ResetPasswordRequest{},
	}, "POST", "/auth/reset-password")

	// -- Users --
	register(&Doc{
		Summary:     "Get user profile",
		Description: "Returns the authenticated user's account information",
		Response:    user.User{},
	}, "GET", "/users/info")

	register(&Doc{
		Summary:     "Internal transfer",
		Description: "Initiates a transfer between two internal Credora accounts, this is enqueued, get the result by calling the transferID value",
		Request:     operation.InternalTransferDTO{},
		Response:    operation.InternalTransferResponse{},
		Headers: []HeaderParam{
			{Name: "Idempotency-Key", Description: "Unique key to prevent duplicate requests", Required: true},
		},
	}, "POST", "/transfers/internal")

	register(&Doc{
		Summary:     "Get transfer status",
		Description: "Gets the result of an initiated transfer via the transferId",
		Response:    idempotency.IdempotencyData{},
	}, "GET", "/transfers/{trxID}/status")

	// ... etc
}
