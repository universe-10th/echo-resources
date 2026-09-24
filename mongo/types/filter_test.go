package types

import (
	"reflect"
	"testing"

	resourcetypes "github.com/universe-10th/echo-resources/types"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type filterProduct struct {
	Resource `bson:",inline"`
	Name     string  `bson:"product_name" json:"name"`
	Price    int     `bson:"price" json:"price"`
	Note     *string `bson:"note,omitempty" json:"note,omitempty"`
}

func TestFilterValidatorUsesDocumentFields(t *testing.T) {
	t.Parallel()

	mapping := NewFieldsMapping[bson.ObjectID, *filterProduct]()
	validator := NewFilterValidator(mapping)

	if mapping.ResourceType() != reflect.TypeOf(filterProduct{}) {
		t.Fatal("expected mapping to expose resource type")
	}
	if mapping.IDType() != reflect.TypeOf(bson.ObjectID{}) {
		t.Fatal("expected mapping to expose ID type")
	}

	if !validator.IsValidCmpFilter("id", bson.NewObjectID().Hex()) {
		t.Fatal("expected ObjectID hex string to be valid for id field")
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
	if !validator.IsNullCheckable("note") {
		t.Fatal("expected pointer field to be null-checkable")
	}
	if !validator.IsExistenceCheckable("name") {
		t.Fatal("expected MongoDB fields to support existence checks")
	}
	// if !validator.IsSortable("created_at") {
	// 	t.Fatal("expected embedded timestamp field to be sortable")
	// }
}

func TestFilterSerializerProducesBSONPredicate(t *testing.T) {
	t.Parallel()

	serializer := NewFilterSerializer(NewFieldsMapping[bson.ObjectID, *filterProduct]())

	got := serializer.Serialize(resourcetypes.FilterExpression{
		Operator: resourcetypes.FilterAnd,
		Expressions: []resourcetypes.FilterExpression{
			{Operator: resourcetypes.FilterContains, Field: "name", Value: "Ada+?"},
			{Operator: resourcetypes.FilterGTE, Field: "price", Value: float64(10)},
		},
	})
	want := bson.M{
		"$and": bson.A{
			bson.M{"product_name": bson.M{"$regex": `Ada\+\?`}},
			bson.M{"price": bson.M{"$gte": float64(10)}},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected BSON predicate\nwant: %#v\n got: %#v", want, got)
	}
}

func TestFilterSerializerProducesNonePredicate(t *testing.T) {
	t.Parallel()

	serializer := NewFilterSerializer(NewFieldsMapping[bson.ObjectID, *filterProduct]())

	got := serializer.Serialize(resourcetypes.FilterExpression{
		Operator: resourcetypes.FilterNone,
	})
	want := bson.M{"$expr": false}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected BSON predicate\nwant: %#v\n got: %#v", want, got)
	}
}
