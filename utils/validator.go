package utils

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/universe-10th/echo-resources/types"
)

var validate = validator.New()

// Validate validates value using go-playground/validator tags.
func Validate(value any) error {
	if err := validate.Struct(value); err != nil {
		var validationErrors validator.ValidationErrors
		if !errors.As(err, &validationErrors) {
			return err
		}

		return types.ValidationError{Errors: collectValidationErrors(validationErrors)}
	}

	return nil
}

func collectValidationErrors(validationErrors validator.ValidationErrors) map[string]any {
	errorsByField := make(map[string]any, len(validationErrors))

	for _, fieldError := range validationErrors {
		fieldName := fieldError.StructField()
		rules, _ := errorsByField[fieldName].([]map[string]string)

		rule := map[string]string{
			"rule": fieldError.Tag(),
		}
		if param := fieldError.Param(); param != "" {
			rule["param"] = param
		}

		errorsByField[fieldName] = append(rules, rule)
	}

	return errorsByField
}
