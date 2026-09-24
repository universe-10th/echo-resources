package types

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

type filterTestValidator struct {
	cmpFields       map[string]bool
	nullFields      map[string]bool
	existenceFields map[string]bool
	containsFields  map[string]bool
	sortableFields  map[string]bool
}

func (v filterTestValidator) IsValidCmpFilter(filter string, value any) bool {
	return v.cmpFields[filter]
}

func (v filterTestValidator) IsNullCheckable(filter string) bool {
	return v.nullFields[filter]
}

func (v filterTestValidator) IsExistenceCheckable(filter string) bool {
	return v.existenceFields[filter]
}

func (v filterTestValidator) IsContainsCheckable(field string) bool {
	return v.containsFields[field]
}

func (v filterTestValidator) IsSortable(field string) bool {
	return v.sortableFields[field]
}

func TestFilterParserParsesComparisonFilter(t *testing.T) {
	t.Parallel()

	filter := parseFilterForTest(t, `{"age":{"$gte":21}}`)

	want := FilterExpression{
		Operator: FilterGTE,
		Field:    "age",
		Value:    float64(21),
	}
	assertFilterExpression(t, filter, want)
}

func TestFilterParserParsesNoneFilter(t *testing.T) {
	t.Parallel()

	filter := parseFilterForTest(t, `{"$none":true}`)

	want := FilterExpression{Operator: FilterNone}
	assertFilterExpression(t, filter, want)
}

func TestFilterParserParsesLogicalFilter(t *testing.T) {
	t.Parallel()

	filter := parseFilterForTest(t, `{
		"$and": [
			{"age": {"$gte": 21}},
			{"$or": [
				{"name": {"$contains": "ada"}},
				{"deleted_at": {"$null": true}}
			]},
			{"$not": {"archived": {"$exists": true}}}
		]
	}`)

	want := FilterExpression{
		Operator: FilterAnd,
		Expressions: []FilterExpression{
			{Operator: FilterGTE, Field: "age", Value: float64(21)},
			{
				Operator: FilterOr,
				Expressions: []FilterExpression{
					{Operator: FilterContains, Field: "name", Value: "ada"},
					{Operator: FilterNull, Field: "deleted_at", Value: true},
				},
			},
			{
				Operator: FilterNot,
				Expressions: []FilterExpression{
					{Operator: FilterExists, Field: "archived", Value: true},
				},
			},
		},
	}
	assertFilterExpression(t, filter, want)
}

func TestFilterParserRejectsInvalidFilters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{name: "non object", input: `"age"`},
		{name: "multiple root clauses", input: `{"age":{"$eq":1},"name":{"$eq":"Ada"}}`},
		{name: "none false", input: `{"$none":false}`},
		{name: "none non bool", input: `{"$none":"true"}`},
		{name: "nested none in and", input: `{"$and":[{"$none":true}]}`},
		{name: "nested none in not", input: `{"$not":{"$none":true}}`},
		{name: "empty and", input: `{"$and":[]}`},
		{name: "not list", input: `{"$and":{"age":{"$eq":1}}}`},
		{name: "invalid field", input: `{"$age":{"$eq":1}}`},
		{name: "field without operation", input: `{"age":1}`},
		{name: "multiple operations", input: `{"age":{"$gte":18,"$lte":30}}`},
		{name: "unknown operation", input: `{"age":{"$between":[18,30]}}`},
		{name: "null non bool", input: `{"deleted_at":{"$null":"yes"}}`},
		{name: "exists non bool", input: `{"archived":{"$exists":"yes"}}`},
		{name: "contains non string", input: `{"name":{"$contains":42}}`},
		{name: "trailing json value", input: `{"age":{"$eq":1}} {"age":{"$eq":2}}`},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parser := NewFilterParser(defaultFilterTestValidator())
			_, err := parser.Parse(json.NewDecoder(strings.NewReader(tt.input)))
			if err == nil {
				t.Fatal("expected invalid filter to return an error")
			}
		})
	}
}

func TestFilterParserRejectsValidatorDeniedFilters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{name: "comparison", input: `{"age":{"$eq":21}}`},
		{name: "null", input: `{"deleted_at":{"$null":true}}`},
		{name: "exists", input: `{"archived":{"$exists":true}}`},
		{name: "contains", input: `{"name":{"$contains":"ada"}}`},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parser := NewFilterParser(filterTestValidator{})
			_, err := parser.Parse(json.NewDecoder(strings.NewReader(tt.input)))
			if err == nil {
				t.Fatal("expected validator-denied filter to return an error")
			}
		})
	}
}

func TestFilterParserRejectsNilDependencies(t *testing.T) {
	t.Parallel()

	parser := NewFilterParser(nil)
	_, err := parser.Parse(json.NewDecoder(strings.NewReader(`{"age":{"$eq":21}}`)))
	if err == nil {
		t.Fatal("expected nil validator to return an error")
	}

	parser = NewFilterParser(defaultFilterTestValidator())
	_, err = parser.Parse(nil)
	if err == nil {
		t.Fatal("expected nil decoder to return an error")
	}
}

func TestFilterSerializerInterface(t *testing.T) {
	t.Parallel()

	var _ FilterSerializer[string] = filterTestSerializer{}
}

func TestFilterExpressionRestrict(t *testing.T) {
	t.Parallel()

	age := FilterExpression{Operator: FilterGTE, Field: "age", Value: 21}
	name := FilterExpression{Operator: FilterContains, Field: "name", Value: "ada"}
	archived := FilterExpression{Operator: FilterExists, Field: "archived", Value: false}
	deleted := FilterExpression{Operator: FilterNull, Field: "deleted_at", Value: true}

	tests := []struct {
		name        string
		current     FilterExpression
		restriction *FilterExpression
		want        FilterExpression
	}{
		{
			name:        "current none remains none",
			current:     FilterExpression{Operator: FilterNone},
			restriction: &age,
			want:        FilterExpression{Operator: FilterNone},
		},
		{
			name:        "restriction none overwrites current",
			current:     age,
			restriction: &FilterExpression{Operator: FilterNone, Field: "ignored", Value: true},
			want:        FilterExpression{Operator: FilterNone},
		},
		{
			name: "and restriction merges into current and",
			current: FilterExpression{
				Operator:    FilterAnd,
				Expressions: []FilterExpression{age},
			},
			restriction: &FilterExpression{
				Operator:    FilterAnd,
				Expressions: []FilterExpression{name, archived},
			},
			want: FilterExpression{
				Operator:    FilterAnd,
				Expressions: []FilterExpression{age, name, archived},
			},
		},
		{
			name: "non-and restriction appends to current and",
			current: FilterExpression{
				Operator:    FilterAnd,
				Expressions: []FilterExpression{age},
			},
			restriction: &name,
			want: FilterExpression{
				Operator:    FilterAnd,
				Expressions: []FilterExpression{age, name},
			},
		},
		{
			name:    "and restriction becomes base and appends current",
			current: age,
			restriction: &FilterExpression{
				Operator:    FilterAnd,
				Expressions: []FilterExpression{name, archived},
			},
			want: FilterExpression{
				Operator:    FilterAnd,
				Expressions: []FilterExpression{name, archived, age},
			},
		},
		{
			name:        "two non-and filters are wrapped in and",
			current:     age,
			restriction: &name,
			want: FilterExpression{
				Operator:    FilterAnd,
				Expressions: []FilterExpression{age, name},
			},
		},
		{
			name:        "empty current becomes restriction",
			current:     FilterExpression{},
			restriction: &deleted,
			want:        deleted,
		},
		{
			name:        "nil restriction is no-op",
			current:     age,
			restriction: nil,
			want:        age,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.current.Restrict(tt.restriction)
			assertFilterExpression(t, tt.current, tt.want)
		})
	}
}

type filterTestSerializer struct{}

func (filterTestSerializer) Serialize(filter FilterExpression) string {
	return string(filter.Operator)
}

func parseFilterForTest(t *testing.T, input string) FilterExpression {
	t.Helper()

	parser := NewFilterParser(defaultFilterTestValidator())
	filter, err := parser.Parse(json.NewDecoder(strings.NewReader(input)))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	return filter
}

func assertFilterExpression(t *testing.T, got FilterExpression, want FilterExpression) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected filter expression\nwant: %#v\n got: %#v", want, got)
	}
}

func defaultFilterTestValidator() filterTestValidator {
	return filterTestValidator{
		cmpFields: map[string]bool{
			"age": true,
		},
		nullFields: map[string]bool{
			"deleted_at": true,
		},
		existenceFields: map[string]bool{
			"archived": true,
		},
		containsFields: map[string]bool{
			"name": true,
		},
		sortableFields: map[string]bool{
			"age": true,
		},
	}
}
