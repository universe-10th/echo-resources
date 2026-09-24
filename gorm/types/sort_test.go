package types

import (
	"testing"

	resourcetypes "github.com/universe-10th/echo-resources/types"
)

func TestSortSourceProvidesSerializerAndValidator(t *testing.T) {
	t.Parallel()

	source := NewSortSource(NewFieldsMapping[int, *filterProduct]())

	if !source.Validator().IsSortable("price", resourcetypes.Asc) {
		t.Fatal("expected source validator to allow scalar field")
	}

	got := source.Serializer().Serialize(resourcetypes.SortExpression{
		Sort: []resourcetypes.Sort{
			{Field: "name", Order: resourcetypes.Asc},
		},
	})
	want := `"product_name" ASC`
	if got != want {
		t.Fatalf("unexpected SQL order fragment\nwant: %s\n got: %s", want, got)
	}
}

func TestSortValidatorUsesModelFields(t *testing.T) {
	t.Parallel()

	validator := NewSortValidator(NewFieldsMapping[int, *filterProduct]())

	if !validator.IsSortable("created_at", resourcetypes.Asc) {
		t.Fatal("expected embedded timestamp field to be sortable")
	}
	if !validator.IsSortable("price", resourcetypes.Desc) {
		t.Fatal("expected scalar field to be sortable")
	}
	if validator.IsSortable("body", resourcetypes.Asc) {
		t.Fatal("expected text storage field to not be sortable")
	}
	if validator.IsSortable("unknown", resourcetypes.Asc) {
		t.Fatal("expected unknown field to not be sortable")
	}
	if validator.IsSortable("price", resourcetypes.OrderType(99)) {
		t.Fatal("expected unknown order type to not be sortable")
	}
}

func TestSortSerializerProducesSQLOrderFragment(t *testing.T) {
	t.Parallel()

	serializer := NewSortSerializer(NewFieldsMapping[int, *filterProduct]())

	got := serializer.Serialize(resourcetypes.SortExpression{
		Sort: []resourcetypes.Sort{
			{Field: "name", Order: resourcetypes.Asc},
			{Field: "price", Order: resourcetypes.Desc},
		},
	})
	want := `"product_name" ASC, "price" DESC`

	if got != want {
		t.Fatalf("unexpected SQL order fragment\nwant: %s\n got: %s", want, got)
	}
}

func TestSortSerializerReturnsEmptySQLOrderFragmentForInvalidSort(t *testing.T) {
	t.Parallel()

	serializer := NewSortSerializer(NewFieldsMapping[int, *filterProduct]())

	got := serializer.Serialize(resourcetypes.SortExpression{
		Sort: []resourcetypes.Sort{
			{Field: "unknown", Order: resourcetypes.Asc},
		},
	})
	if got != "" {
		t.Fatalf("expected empty SQL order fragment, got %q", got)
	}

	got = serializer.Serialize(resourcetypes.SortExpression{
		Sort: []resourcetypes.Sort{
			{Field: "price", Order: resourcetypes.OrderType(99)},
		},
	})
	if got != "" {
		t.Fatalf("expected empty SQL order fragment, got %q", got)
	}
}
