// pkg/pgerrors/handlers.go or anywhere reusable
package pgerrors

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/EfosaE/credora-backend/internal/response"
	"github.com/jackc/pgx/v5/pgconn"
)

var constraintToField = map[string]string{
	"users_email_key":        "email",
	"users_phone_number_key": "phone_number",
	"users_nin_key":          "nin",
}

// HandleUniqueViolation inspects the error and sends a 400 if it's a unique constraint error.
// Returns true if it handled the error, false if not.
func HandleUniqueViolation(w http.ResponseWriter, r *http.Request, err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		field, ok := constraintToField[pgErr.ConstraintName]
		if !ok {
			field = "field"
		}
		msg := fmt.Sprintf("%s already exists", field)
		response.SendError(w, r, response.BadRequest(map[string]string{"field": field}, msg))
		return true
	}
	return false
}
