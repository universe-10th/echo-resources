package types

import (
	"reflect"
	"testing"

	resourcetypes "github.com/universe-10th/rest-resources/types"
)

type filterProduct struct {
	Resource[int]
	Name  string  `gorm:"column:product_name" json:"name"`
	Body  string  `gorm:"type:text" json:"body"`
	Price int     `json:"price"`
	Note  *string `json:"note,omitempty"`
}

func TestFilterValidatorUsesModelFields(t *testing.T) {
	t.Parallel()

	mapping := NewFieldsMapping[int, *filterProduct]()
	validator := NewFilterValidator(mapping)

	if mapping.ResourceType() != reflect.TypeOf(filterProduct{}) {
		t.Fatal("expected mapping to expose resource type")
	}
	if mapping.IDType() != reflect.TypeOf(int(0)) {
		t.Fatal("expected mapping to expose ID type")
	}

	if !validator.IsValidCmpFilter("price", float64(10)) {
		t.Fatal("expected integral JSON number to be valid for int field")
	}
	if validator.IsValidCmpFilter("price", 10.5) {
		t.Fatal("expected fractional JSON number to be invalid for int field")
	}
	if !validator.IsContainsCheckable("name") {
		t.Fatal("expected string field to be contains-checkable")
	}
	if validator.IsContainsCheckable("price") {
		t.Fatal("expected int field to not be contains-checkable")
	}
	if !validator.IsNullCheckable("note") {
		t.Fatal("expected pointer field to be null-checkable")
	}
	if validator.IsExistenceCheckable("name") {
		t.Fatal("expected GORM fields to not support existence checks")
	}
	// if !validator.IsSortable("created_at") {
	// 	t.Fatal("expected embedded timestamp field to be sortable")
	// }
	// if validator.IsSortable("body") {
	// 	t.Fatal("expected text storage field to not be sortable")
	// }
}

func TestFilterSerializerProducesSQLPredicate(t *testing.T) {
	t.Parallel()

	serializer := NewFilterSerializer(NewFieldsMapping[int, *filterProduct]())

	got := serializer.Serialize(resourcetypes.FilterExpression{
		Operator: resourcetypes.FilterAnd,
		Expressions: []resourcetypes.FilterExpression{
			{Operator: resourcetypes.FilterEQ, Field: "name", Value: "Ada"},
			{Operator: resourcetypes.FilterGTE, Field: "price", Value: float64(10)},
		},
	})
	want := `("product_name" = 'Ada') AND ("price" >= 10)`

	if got != want {
		t.Fatalf("unexpected SQL predicate\nwant: %s\n got: %s", want, got)
	}
}

func TestFilterSerializerProducesNonePredicate(t *testing.T) {
	t.Parallel()

	serializer := NewFilterSerializer(NewFieldsMapping[int, *filterProduct]())

	got := serializer.Serialize(resourcetypes.FilterExpression{
		Operator: resourcetypes.FilterNone,
	})
	want := "1 = 0"

	if got != want {
		t.Fatalf("unexpected SQL predicate\nwant: %s\n got: %s", want, got)
	}
}
