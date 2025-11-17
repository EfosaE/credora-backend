package validation

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ParseValidationErrors returns user-friendly messages
// and ensures all error keys use json/alias tags instead
// of Go struct field names.
func ParseValidationErrors(err error) map[string]string {
	validationErrors := make(map[string]string)

	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		return validationErrors
	}

	for _, e := range errs {
		jsonField := extractFieldName(e)

		var msg string

		switch e.Tag() {

		// Common validation rules
		case "required":
			msg = fmt.Sprintf("%s is required", jsonField)

		case "email":
			msg = fmt.Sprintf("%s must be a valid email address", jsonField)

		case "min":
			msg = fmt.Sprintf("%s must be at least %s characters long", jsonField, e.Param())

		case "max":
			msg = fmt.Sprintf("%s must be at most %s characters long", jsonField, e.Param())

		case "len":
			msg = fmt.Sprintf("%s must be exactly %s characters long", jsonField, e.Param())

		case "oneof":
			msg = fmt.Sprintf("%s must be one of: %s", jsonField, e.Param())

		// Numeric comparisons
		case "gt":
			msg = fmt.Sprintf("%s must be greater than %s", jsonField, e.Param())

		case "gte":
			msg = fmt.Sprintf("%s must be greater than or equal to %s", jsonField, e.Param())

		case "lt":
			msg = fmt.Sprintf("%s must be less than %s", jsonField, e.Param())

		case "lte":
			msg = fmt.Sprintf("%s must be less than or equal to %s", jsonField, e.Param())

		// Custom rules
		case "uuid4":
			msg = fmt.Sprintf("%s must be a valid UUID", jsonField)

		case "decimal":
			msg = fmt.Sprintf("%s must be a valid decimal value", jsonField)

		default:
			msg = fmt.Sprintf("%s is invalid", jsonField)
		}

		validationErrors[jsonField] = msg
	}

	return validationErrors
}

// extractFieldName gets JSON tag or alias tag automatically.
// Fallback = struct field name.
func extractFieldName(e validator.FieldError) string {
	field := e.StructField() // Go struct field
	
	// Get the actual struct type being validated
	structType := e.Type()
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}
	
	// Try to get the struct field from the type
	if structType.Kind() == reflect.Struct {
		if sf, ok := structType.FieldByName(field); ok {
			// 1. alias tag has highest priority
			if alias := sf.Tag.Get("alias"); alias != "" {
				return alias
			}

			// 2. use json tag if available
			if jsonTag := sf.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
				return strings.Split(jsonTag, ",")[0] // first part only
			}
		}
	}

	// 3. fallback: use snake_case struct name
	return toSnakeCase(field)
}

// Converts CamelCase → snake_case (FirstName → first_name)
func toSnakeCase(str string) string {
	var out []rune
	for i, r := range str {
		if i > 0 && r >= 'A' && r <= 'Z' {
			out = append(out, '_')
		}
		out = append(out, r)
	}
	return strings.ToLower(string(out))
}