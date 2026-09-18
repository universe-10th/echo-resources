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

	got, code := RenderError(err)
	want := map[string]any{
		"element_name": "widget",
		"key":          10,
		"detail":       "element not found",
	}

	if code != uint16(ErrNotFound) {
		t.Fatalf("expected status code %d, got %d", ErrNotFound, code)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected rendered error\nwant: %#v\n got: %#v", want, got)
	}
}

func TestRenderErrorIncludesEmptyErrorDetails(t *testing.T) {
	t.Parallel()

	got, code := RenderError(BadRequestError{})
	want := map[string]any{
		"detail": "bad request",
	}

	if code != uint16(ErrBadRequest) {
		t.Fatalf("expected status code %d, got %d", ErrBadRequest, code)
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

	got, code := RenderError(err)
	want := map[string]any{
		"errors": map[string]any{
			"name": "required",
		},
		"detail": "invalid data",
	}

	if code != uint16(ErrInvalid) {
		t.Fatalf("expected status code %d, got %d", ErrInvalid, code)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected rendered error\nwant: %#v\n got: %#v", want, got)
	}
}
