package types

import (
	"reflect"
	"testing"
)

func TestRenderErrorIncludesErrorFields(t *testing.T) {
	t.Parallel()

	err := NotFoundError[int]{
		ElementName: "widget",
		Key:         10,
	}

	got := RenderError(err)
	want := map[string]any{
		"element_name": "widget",
		"key":          10,
		"code":         ErrNotFound,
		"detail":       "element not found",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected rendered error\nwant: %#v\n got: %#v", want, got)
	}
}

func TestRenderErrorIncludesEmptyErrorDetails(t *testing.T) {
	t.Parallel()

	got := RenderError(BadRequestError{})
	want := map[string]any{
		"code":   ErrBadRequest,
		"detail": "bad request",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected rendered error\nwant: %#v\n got: %#v", want, got)
	}
}

func TestRenderErrorAcceptsPointers(t *testing.T) {
	t.Parallel()

	err := &ValidationError{
		Errors: map[string]any{
			"name": "required",
		},
	}

	got := RenderError(err)
	want := map[string]any{
		"errors": map[string]any{
			"name": "required",
		},
		"code":   ErrInvalid,
		"detail": "invalid data",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected rendered error\nwant: %#v\n got: %#v", want, got)
	}
}
