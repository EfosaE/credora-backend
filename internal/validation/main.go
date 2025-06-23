package validation

import (
	"errors"
	"log"

	"github.com/go-playground/validator/v10"
)

var Validate = validator.New()


func SafeValidateStruct(v *validator.Validate, s any) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Validation panic recovered: %v", r)
			err = errors.New("internal validation error: " + r.(string))
		}
	}()
	return v.Struct(s)
}
