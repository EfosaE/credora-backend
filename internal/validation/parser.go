package validation

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

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
		fieldName := extractFieldName(e)

		var msg string

		switch e.Tag() {

		case "required":
			msg = fmt.Sprintf("%s is required", fieldName)

		case "email":
			msg = fmt.Sprintf("%s must be a valid email address", fieldName)

		case "min":
			msg = fmt.Sprintf("%s must be at least %s characters long", fieldName, e.Param())

		case "max":
			msg = fmt.Sprintf("%s must be at most %s characters long", fieldName, e.Param())

		case "len":
			msg = fmt.Sprintf("%s must be exactly %s characters long", fieldName, e.Param())

		case "oneof":
			msg = fmt.Sprintf("%s must be one of: %s", fieldName, e.Param())

		case "gt":
			msg = fmt.Sprintf("%s must be greater than %s", fieldName, e.Param())

		case "gte":
			msg = fmt.Sprintf("%s must be greater than or equal to %s", fieldName, e.Param())

		case "lt":
			msg = fmt.Sprintf("%s must be less than %s", fieldName, e.Param())

		case "lte":
			msg = fmt.Sprintf("%s must be less than or equal to %s", fieldName, e.Param())

		case "uuid4":
			msg = fmt.Sprintf("%s must be a valid UUID", fieldName)

		case "decimal":
			msg = fmt.Sprintf("%s must be a valid decimal value", fieldName)

		default:
			msg = fmt.Sprintf("%s is invalid", fieldName)
		}

		validationErrors[fieldName] = msg
	}

	return validationErrors
}

// extractFieldName gets alias or json tag automatically.
// Fallback = camelCase struct field name.
func extractFieldName(e validator.FieldError) string {
	field := e.StructField()

	structType := e.Type()
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}

	if structType.Kind() == reflect.Struct {
		if sf, ok := structType.FieldByName(field); ok {

			// 1. alias tag has highest priority
			if alias := sf.Tag.Get("alias"); alias != "" {
				return alias
			}

			// 2. use json tag if available
			if jsonTag := sf.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
				return strings.Split(jsonTag, ",")[0]
			}
		}
	}

	// 3. fallback: use camelCase struct name
	return toCamelCase(field)
}

// Converts PascalCase → camelCase (UserId → userId)
func toCamelCase(str string) string {
	if str == "" {
		return str
	}

	runes := []rune(str)
	runes[0] = unicode.ToLower(runes[0])

	return string(runes)
}
