package validation

import (
	"errors"
	"fmt"
	"log"

	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New()
	RegisterCustomRules(Validate)
}

func SafeValidateStruct(v *validator.Validate, s any) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Validation panic recovered: %v", r)
			// Handle different types that might be recovered
			switch x := r.(type) {
			case string:
				err = errors.New("internal validation error: " + x)
			case error:
				err = fmt.Errorf("internal validation error: %w", x)
			default:
				err = fmt.Errorf("internal validation error: %v", x)
			}
		}
	}()
	return v.Struct(s)
}