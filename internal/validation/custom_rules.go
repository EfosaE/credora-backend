// internal/validation/custom_rules.go
package validation

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func RegisterCustomRules(v *validator.Validate) {
	// UUID validator
	v.RegisterValidation("uuid4", func(fl validator.FieldLevel) bool {
		_, err := uuid.Parse(fl.Field().String())
		return err == nil
	})

	// Decimal validator
	v.RegisterValidation("decimal", func(fl validator.FieldLevel) bool {
		_, err := decimal.NewFromString(fl.Field().String())
		return err == nil
	})
}
