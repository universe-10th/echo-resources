package types

import (
	"reflect"
	"testing"

	resourcetypes "github.com/universe-10th/echo-resources/types"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestSortSourceProvidesSerializerAndValidator(t *testing.T) {
	t.Parallel()

	source := NewSortSource(NewFieldsMapping[bson.ObjectID, *filterProduct]())

	if !source.Validator().IsSortable("price", resourcetypes.Asc) {
		t.Fatal("expected source validator to allow scalar field")
	}

	got := source.Serializer().Serialize(resourcetypes.SortExpression{
		Sort: []resourcetypes.Sort{
			{Field: "name", Order: resourcetypes.Asc},
		},
	})
	want := bson.D{{Key: "product_name", Value: 1}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected BSON sort\nwant: %#v\n got: %#v", want, got)
	}
}

func TestSortValidatorUsesDocumentFields(t *testing.T) {
	t.Parallel()

	validator := NewSortValidator(NewFieldsMapping[bson.ObjectID, *filterProduct]())

	if !validator.IsSortable("created_at", resourcetypes.Asc) {
		t.Fatal("expected embedded timestamp field to be sortable")
	}
	if !validator.IsSortable("price", resourcetypes.Desc) {
		t.Fatal("expected scalar field to be sortable")
	}
	if validator.IsSortable("unknown", resourcetypes.Asc) {
		t.Fatal("expected unknown field to not be sortable")
	}
	if validator.IsSortable("price", resourcetypes.OrderType(99)) {
		t.Fatal("expected unknown order type to not be sortable")
	}
}

func TestSortSerializerProducesBSONSort(t *testing.T) {
	t.Parallel()

	serializer := NewSortSerializer(NewFieldsMapping[bson.ObjectID, *filterProduct]())

	got := serializer.Serialize(resourcetypes.SortExpression{
		Sort: []resourcetypes.Sort{
			{Field: "name", Order: resourcetypes.Asc},
			{Field: "price", Order: resourcetypes.Desc},
		},
	})
	want := bson.D{
		{Key: "product_name", Value: 1},
		{Key: "price", Value: -1},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected BSON sort\nwant: %#v\n got: %#v", want, got)
	}
}

func TestSortSerializerReturnsEmptyBSONSortForInvalidSort(t *testing.T) {
	t.Parallel()

	serializer := NewSortSerializer(NewFieldsMapping[bson.ObjectID, *filterProduct]())

	got := serializer.Serialize(resourcetypes.SortExpression{
		Sort: []resourcetypes.Sort{
			{Field: "unknown", Order: resourcetypes.Asc},
		},
	})
	if len(got) != 0 {
		t.Fatalf("expected empty BSON sort, got %#v", got)
	}

	got = serializer.Serialize(resourcetypes.SortExpression{
		Sort: []resourcetypes.Sort{
			{Field: "price", Order: resourcetypes.OrderType(99)},
		},
	})
	if len(got) != 0 {
		t.Fatalf("expected empty BSON sort, got %#v", got)
	}
}
