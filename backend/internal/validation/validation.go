package validation

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// Struct validates a struct using its `validate` tags and returns a
// human-readable error message, or nil if valid.
func Struct(s interface{}) error {
	if err := validate.Struct(s); err != nil {
		var ve validator.ValidationErrors
		if ok := asValidationErrors(err, &ve); ok && len(ve) > 0 {
			first := ve[0]
			return fmt.Errorf("field %q failed validation: %s", strings.ToLower(first.Field()), ruleMessage(first))
		}
		return err
	}
	return nil
}

func asValidationErrors(err error, target *validator.ValidationErrors) bool {
	if ve, ok := err.(validator.ValidationErrors); ok {
		*target = ve
		return true
	}
	return false
}

func ruleMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email"
	case "min":
		return "must be at least " + fe.Param() + " characters/value"
	case "max":
		return "must be at most " + fe.Param() + " characters/value"
	case "gt":
		return "must be greater than " + fe.Param()
	case "oneof":
		return "must be one of: " + fe.Param()
	default:
		return "is invalid (" + fe.Tag() + ")"
	}
}
