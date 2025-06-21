package utils
import (
    "fmt"
    "strings"

    "github.com/go-playground/validator/v10"
)

func ParseValidationErrors(err error) map[string]string {
    validationErrors := make(map[string]string)

    if errs, ok := err.(validator.ValidationErrors); ok {
        for _, e := range errs {
            field := strings.ToLower(e.Field())
            var msg string

            switch e.Tag() {
            case "required":
                msg = fmt.Sprintf("%s is required", field)
            case "email":
                msg = fmt.Sprintf("%s must be a valid email address", field)
            case "min":
                msg = fmt.Sprintf("%s must be at least %s characters", field, e.Param())
            case "max":
                msg = fmt.Sprintf("%s must be at most %s characters", field, e.Param())
            default:
                msg = fmt.Sprintf("%s is invalid", field)
            }

            validationErrors[field] = msg
        }
    }

    return validationErrors
}
