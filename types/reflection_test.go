package types

import (
	"reflect"
	"testing"
)

type jsonToFieldTestValue struct {
	ID      string `json:"id"`
	Name    string `json:"name,omitempty"`
	Enabled bool
	Ignored string `json:"-"`
	hidden  string
}

func TestJSONToFieldMapsJSONNamesToStructFields(t *testing.T) {
	t.Parallel()

	got := JSONToField(jsonToFieldTestValue{})
	want := map[string]string{
		"id":      "ID",
		"name":    "Name",
		"Enabled": "Enabled",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected mapping\nwant: %#v\n got: %#v", want, got)
	}
}

func TestJSONToFieldAcceptsPointers(t *testing.T) {
	t.Parallel()

	got := JSONToField(&jsonToFieldTestValue{})
	want := map[string]string{
		"id":      "ID",
		"name":    "Name",
		"Enabled": "Enabled",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected mapping\nwant: %#v\n got: %#v", want, got)
	}
}

func TestJSONToFieldRejectsNonStructValues(t *testing.T) {
	t.Parallel()

	got := JSONToField("not a struct")
	if len(got) != 0 {
		t.Fatalf("expected empty mapping, got %#v", got)
	}
}

func TestJSONToFieldAcceptsNilPointers(t *testing.T) {
	t.Parallel()

	var value *jsonToFieldTestValue

	got := JSONToField(value)
	want := map[string]string{
		"id":      "ID",
		"name":    "Name",
		"Enabled": "Enabled",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected mapping\nwant: %#v\n got: %#v", want, got)
	}
}
