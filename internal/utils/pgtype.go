package utils

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

func ToPgText(s string) pgtype.Text {
	return pgtype.Text{
		String: s,
		Valid:  s != "",
	}
}

// ToPgUUID converts a google/uuid.UUID to pgtype.UUID
func ToPgUUID(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: u,
		Valid: true,
	}
}

// ToPgUUID converts a google/uuid.UUID to pgtype.UUID
func ToPgNullableUUID(u *uuid.UUID) pgtype.UUID {
	if u == nil {
		return pgtype.UUID{
			Valid: false,
		}
	}

	return pgtype.UUID{
		Bytes: *u,
		Valid: true,
	}
}

// FromPgUUID converts a pgtype.UUID to google/uuid.UUID
func FromPgUUID(pg pgtype.UUID) uuid.UUID {
	return uuid.UUID(pg.Bytes)
}

// Coverts pgtype.Numeric to string
func PgNumericToString(n pgtype.Numeric) string {
	v, err := n.Value()
	if err != nil {
		panic(err)
	}
	return v.(string)
}

func DecimalToPgNumeric(d decimal.Decimal) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	err := n.Scan(d.String())
	return n, err
}
func PgNumericToDecimal(n pgtype.Numeric) (decimal.Decimal, error) {
	if !n.Valid {
		return decimal.Zero, fmt.Errorf("numeric is null")
	}

	val, err := n.Value()
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to extract value: %w", err)
	}

	switch v := val.(type) {
	case string:
		return decimal.NewFromString(v)
	default:
		return decimal.Zero, fmt.Errorf("unexpected pgtype.Numeric value type: %T", val)
	}
}

func MustPgNumericToDecimal(n pgtype.Numeric) decimal.Decimal {
	if !n.Valid {
		panic("pgtype.Numeric is null")
	}
	val, err := n.Value()
	if err != nil {
		panic(fmt.Sprintf("failed to extract numeric value: %v", err))
	}
	switch v := val.(type) {
	case string:
		dec, err := decimal.NewFromString(v)
		if err != nil {
			panic(fmt.Sprintf("invalid decimal string: %v", err))
		}
		return dec
	default:
		panic(fmt.Sprintf("unexpected pgtype.Numeric type: %T", val))
	}
}

func MustDecimalToPgNumeric(d decimal.Decimal) pgtype.Numeric {
	var n pgtype.Numeric
	if err := n.Scan(d.String()); err != nil {
		panic(fmt.Sprintf("failed to convert decimal to pgtype.Numeric: %v", err))
	}
	return n
}

func NewPgNumericFromString(s string) pgtype.Numeric {
	var n pgtype.Numeric
	if err := n.Scan(s); err != nil {
		panic(fmt.Sprintf("failed to convert string to pgtype.Numeric: %v", err))
	}
	return n
}
