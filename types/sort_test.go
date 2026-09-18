package types

import (
	"reflect"
	"testing"
)

type sortTestValidator struct {
	fields map[string]map[OrderType]bool
}

func (v sortTestValidator) IsSortable(field string, orderType OrderType) bool {
	orders, ok := v.fields[field]
	if !ok {
		return false
	}

	return orders[orderType]
}

func TestSortParserParsesSortSentence(t *testing.T) {
	t.Parallel()

	parser := NewSortParser(defaultSortTestValidator())
	got, err := parser.Parse("foo,bar,-baz")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	want := SortExpression{
		Sort: []Sort{
			{Field: "foo", Order: Asc},
			{Field: "bar", Order: Asc},
			{Field: "baz", Order: Desc},
		},
	}
	assertSortExpression(t, got, want)
}

func TestSortParserParsesEmptySortSentence(t *testing.T) {
	t.Parallel()

	parser := NewSortParser(defaultSortTestValidator())
	got, err := parser.Parse("")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	want := SortExpression{Sort: []Sort{}}
	assertSortExpression(t, got, want)
}

func TestSortParserRejectsInvalidSortSentences(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{name: "empty field", input: "foo,,bar"},
		{name: "trailing comma", input: "foo,"},
		{name: "missing descending field", input: "-"},
		{name: "invalid field", input: "$foo"},
		{name: "invalid descending field", input: "--foo"},
		{name: "space", input: "foo, bar"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parser := NewSortParser(defaultSortTestValidator())
			_, err := parser.Parse(tt.input)
			if err == nil {
				t.Fatal("expected invalid sort sentence to return an error")
			}
		})
	}
}

func TestSortParserRejectsValidatorDeniedSorts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{name: "ascending", input: "denied"},
		{name: "descending", input: "-foo"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parser := NewSortParser(sortTestValidator{
				fields: map[string]map[OrderType]bool{
					"foo": {
						Asc: true,
					},
				},
			})
			_, err := parser.Parse(tt.input)
			if err == nil {
				t.Fatal("expected validator-denied sort to return an error")
			}
		})
	}
}

func TestSortParserRejectsNilDependencies(t *testing.T) {
	t.Parallel()

	parser := NewSortParser(nil)
	_, err := parser.Parse("foo")
	if err == nil {
		t.Fatal("expected nil validator to return an error")
	}
}

func assertSortExpression(t *testing.T, got SortExpression, want SortExpression) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected sort expression\nwant: %#v\n got: %#v", want, got)
	}
}

func defaultSortTestValidator() sortTestValidator {
	return sortTestValidator{
		fields: map[string]map[OrderType]bool{
			"foo": {
				Asc:  true,
				Desc: true,
			},
			"bar": {
				Asc: true,
			},
			"baz": {
				Desc: true,
			},
		},
	}
}
