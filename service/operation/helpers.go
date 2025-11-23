package operationsvc

import (
	"errors"
	"fmt"
	"github.com/EfosaE/credora-backend/domain/operation"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func wrapAccountLockError(account string, err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("account %s not found: %w", account, operation.ErrAccountNotFound)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return fmt.Errorf("postgres error (%s) locking account %s: %w",
			pgErr.Code, account, err)
	}

	return fmt.Errorf("lock account %s: %w", account, err)
}
