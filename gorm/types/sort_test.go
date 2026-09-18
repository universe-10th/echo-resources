package types

import (
	"testing"

	gormreflection "github.com/universe-10th/echo-resources/gorm/types/reflection"
	resourcetypes "github.com/universe-10th/echo-resources/types"
)

func TestSortValidatorUsesModelFields(t *testing.T) {
	t.Parallel()

	validator := NewSortValidator(gormreflection.NewFieldsMapping[int, filterProduct]())

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
