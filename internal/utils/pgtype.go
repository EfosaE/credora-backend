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

// ConvertUUID converts a google uuid.UUID to pgtype.UUID
func ConvertUUID(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: u,
		Valid: true,
	}
}

// func PgNumericToDecimal(n pgtype.Numeric) (decimal.Decimal, error) {
// 	if !n.Valid {
// 		return decimal.Zero, fmt.Errorf("numeric is null")
// 	}
// 	return decimal.NewFromString(n.Int.String())
// }

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
