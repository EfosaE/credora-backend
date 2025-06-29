package account

import (
	"context"

	// "github.com/google/uuid"
)

// AccountRepository defines the methods that the sqlc account repository should implement.
type AccountRepository interface {
	CreateAcct(ctx context.Context, req *CreateAccountRequest) (*Account, error)
	
}
