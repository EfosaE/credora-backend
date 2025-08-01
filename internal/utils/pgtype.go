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

func PgNumericToDecimal(n pgtype.Numeric) (decimal.Decimal, error) {
	if !n.Valid {
		return decimal.Zero, fmt.Errorf("numeric is null")
	}
	return decimal.NewFromString(n.Int.String())
}


func DecimalToPgNumeric(d decimal.Decimal) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	err := n.Scan(d.String())
	return n, err
}
