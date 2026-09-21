package utils

import (
	"errors"
	"reflect"
	"testing"

	"github.com/universe-10th/echo-resources/types"
)

type validationTestResource struct {
	Username string `validate:"min=5,excludesall= "`
	Email    string `validate:"email"`
}

func TestValidateReturnsNilWhenValid(t *testing.T) {
	t.Parallel()

	err := Validate(validationTestResource{
		Username: "valid",
		Email:    "user@example.com",
	})

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestValidateReturnsValidationError(t *testing.T) {
	t.Parallel()

	err := Validate(validationTestResource{
		Username: "a b",
		Email:    "invalid",
	})

	var validationError types.ValidationError
	if !errors.As(err, &validationError) {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	want := map[string]any{
		"Username": []map[string]string{
			{"rule": "min", "param": "5"},
		},
		"Email": []map[string]string{
			{"rule": "email"},
		},
	}

	if !reflect.DeepEqual(validationError.Errors, want) {
		t.Fatalf("unexpected validation errors\nwant: %#v\n got: %#v", want, validationError.Errors)
	}
}
