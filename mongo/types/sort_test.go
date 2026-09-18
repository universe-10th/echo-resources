package types

import (
	"testing"

	mongoreflection "github.com/universe-10th/echo-resources/mongo/types/reflection"
	resourcetypes "github.com/universe-10th/echo-resources/types"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestSortValidatorUsesDocumentFields(t *testing.T) {
	t.Parallel()

	validator := NewSortValidator(mongoreflection.NewFieldsMapping[bson.ObjectID, filterProduct]())

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
