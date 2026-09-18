package types

import (
	resourcetypes "github.com/universe-10th/echo-resources/types"
	resourcereflection "github.com/universe-10th/echo-resources/types/reflection"
)

// SortValidator validates sort fields against a MongoDB field mapping.
type SortValidator struct {
	mapping *resourcereflection.FieldsMapping
}

// NewSortValidator returns a sort validator for the supplied MongoDB field mapping.
func NewSortValidator(mapping *resourcereflection.FieldsMapping) SortValidator {
	return SortValidator{mapping: mapping}
}

// IsSortable reports whether field can be used in MongoDB sort expressions.
func (v SortValidator) IsSortable(field string, orderType resourcetypes.OrderType) bool {
	if !isValidOrderType(orderType) {
		return false
	}

	structField, ok := resourcereflection.StructFieldForJSON(v.mapping, field)
	if !ok {
		return false
	}

	return isScalar(structField.Type)
}

func isValidOrderType(orderType resourcetypes.OrderType) bool {
	switch orderType {
	case resourcetypes.Asc, resourcetypes.Desc:
		return true
	default:
		return false
	}
}

var _ resourcetypes.SortValidator = SortValidator{}
