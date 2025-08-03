// pkg/pgerrors/handlers.go or anywhere reusable
package pgerrors

import (
	// "errors"
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
// func HandleUniqueViolation(w http.ResponseWriter, r *http.Request, err error) bool {
// 	var pgErr *pgconn.PgError
// 	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
// 		field, ok := constraintToField[pgErr.ConstraintName]
// 		if !ok {
// 			field = "field"
// 		}
// 		msg := fmt.Sprintf("%s already exists", field)
// 		response.SendError(w, r, response.BadRequest(map[string]string{"field": field}, msg))
// 		return true
// 	}
// 	return false
// }

func HandlePGError(w http.ResponseWriter, r *http.Request, err error) bool {
	pgErr, ok := err.(*pgconn.PgError)
	if !ok {
		return false // Not a Postgres error
	}

	switch pgErr.Code {
	case "23505": // unique_violation
		field, ok := constraintToField[pgErr.ConstraintName]
		if !ok {
			field = "field"
		}
		msg := fmt.Sprintf("%s already exists", field)
		response.SendError(w, r, response.BadRequest(map[string]string{"field": field}, msg))
		return true

	case "23503": // foreign_key_violation
		msg := "Related resource not found"
		response.SendError(w, r, response.BadRequest(nil, msg))
		return true

	case "23502": // not_null_violation
		msg := fmt.Sprintf("Missing required field: %s", pgErr.ColumnName)
		response.SendError(w, r, response.BadRequest(map[string]string{"field": pgErr.ColumnName}, msg))
		return true

	// Add more codes if needed (e.g., check_violation, exclusion_violation)
	default:
		return false // unhandled pg error
	}
}
