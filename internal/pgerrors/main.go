package pgerrors

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/EfosaE/credora-backend/internal/response"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// fieldDisplayNames maps database column names to user-friendly display names
// Only add entries here if you want a different display name than the column name
var fieldDisplayNames = map[string]string{
	"nin":          "NIN",
	"phone_number": "phone number",
	// Add more custom display names as needed
}

// extractColumnFromConstraint attempts to extract the column name from a constraint name
// Common patterns: "table_column_key", "table_column_idx", "table_column_unique", "table_pkey"
func extractColumnFromConstraint(constraintName, detail string) string {
	// Detail format: "Key (column_name)=(value) already exists."
	// Use this for ALL constraints, not just pkey
	if detail != "" {
		if start := strings.Index(detail, "("); start != -1 {
			if end := strings.Index(detail[start:], ")"); end != -1 {
				columnName := detail[start+1 : start+end]
				if commaIdx := strings.Index(columnName, ","); commaIdx != -1 {
					columnName = strings.TrimSpace(columnName[:commaIdx])
				}
				return columnName // returns "phone_number" ✅
			}
		}
	}

	// fallback: strip known suffixes and try constraint name
	name := constraintName
	for _, suffix := range []string{"_key", "_idx", "_unique", "_pkey"} {
		name = strings.TrimSuffix(name, suffix)
	}
	return name
}

// getDisplayName returns a user-friendly name for a field
func getDisplayName(columnName string) string {
	if displayName, ok := fieldDisplayNames[columnName]; ok {
		return displayName
	}
	// Return the column name as-is (keeping snake_case)
	return columnName
}

// HandlePGError inspects common Postgres + SQL errors.
// Returns true if it handled the error (meaning the caller should STOP).
func HandlePGError(w http.ResponseWriter, r *http.Request, err error) bool {
	if err == nil {
		return false
	}

	// ------------------------------------------
	// 1️⃣ Handle "no rows" error (sqlc / pgx / stdlib)
	// ------------------------------------------
	if errors.Is(err, pgx.ErrNoRows) {
		response.SendError(w, r, response.NotFound(
			"Resource not found",
		))
		return true
	}

	if errors.Is(err, pgx.ErrTxClosed) {
		response.SendError(w, r, response.InternalServerError(
			err,
			"Transaction is closed unexpectedly",
		))
		return true
	}

	// ------------------------------------------
	// 2️⃣ Handle Postgres-specific errors
	// ------------------------------------------
	pgErr, ok := err.(*pgconn.PgError)
	if !ok {
		return false // not a PostgreSQL error
	}

	switch pgErr.Code {

	// ------------------------------------------
	// UNIQUE VIOLATION
	// ------------------------------------------
	case "23505":
		columnName := extractColumnFromConstraint(pgErr.ConstraintName, pgErr.Detail)
		displayName := getDisplayName(columnName)

		msg := fmt.Sprintf("%s already exists", displayName)
		response.SendError(w, r, response.BadRequest(
			map[string]string{"field": columnName},
			msg,
		))
		return true

	// ------------------------------------------
	// FOREIGN KEY VIOLATION
	// ------------------------------------------
	case "23503":
		response.SendError(w, r, response.BadRequest(
			nil,
			"Related resource does not exist",
		))
		return true

	// ------------------------------------------
	// NOT NULL VIOLATION
	// ------------------------------------------
	case "23502":
		displayName := getDisplayName(pgErr.ColumnName)
		msg := fmt.Sprintf("Missing required field: %s", displayName)
		response.SendError(w, r, response.BadRequest(
			map[string]string{"field": pgErr.ColumnName},
			msg,
		))
		return true

	// ------------------------------------------
	// CHECK CONSTRAINT VIOLATION
	// ------------------------------------------
	case "23514":
		msg := fmt.Sprintf("Constraint failed: %s", pgErr.ConstraintName)
		response.SendError(w, r, response.BadRequest(nil, msg))
		return true

	// ------------------------------------------
	// EXCLUSION VIOLATION
	// ------------------------------------------
	case "23P01":
		msg := fmt.Sprintf("Exclusion constraint failed: %s", pgErr.ConstraintName)
		response.SendError(w, r, response.BadRequest(nil, msg))
		return true

	// ------------------------------------------
	// UNDEFINED COLUMN
	// e.g. bad query, typo in column name
	// ------------------------------------------
	case "42703":
		msg := fmt.Sprintf("Column does not exist: %s", pgErr.ColumnName)
		response.SendError(w, r, response.BadRequest(nil, msg))
		return true

	// ------------------------------------------
	// FALLBACK (unhandled Postgres error)
	// ------------------------------------------
	default:
		return false
	}
}
