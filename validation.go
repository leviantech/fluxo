package fluxo

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func validateStruct(s interface{}) error {
	if err := validate.Struct(s); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return fmt.Errorf("validation failed: %v", err)
		}

		var errorMessages []string
		for _, e := range validationErrors {
			errorMessages = append(errorMessages, formatValidationError(e))
		}

		return fmt.Errorf("validation failed: %s", strings.Join(errorMessages, "; "))
	}
	return nil
}

func formatValidationError(e validator.FieldError) string {
	field := e.Field()
	tag := e.Tag()
	param := e.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, param)
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", field, param)
	case "numeric":
		return fmt.Sprintf("%s must be numeric", field)
	case "alpha":
		return fmt.Sprintf("%s must contain only letters", field)
	case "alphanum":
		return fmt.Sprintf("%s must contain only letters and numbers", field)
	default:
		return fmt.Sprintf("%s failed validation for %s", field, tag)
	}
}
