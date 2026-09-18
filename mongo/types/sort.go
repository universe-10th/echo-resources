package types

import (
	"reflect"
	"time"

	resourcetypes "github.com/universe-10th/echo-resources/types"
	resourcereflection "github.com/universe-10th/echo-resources/types/reflection"
	"go.mongodb.org/mongo-driver/v2/bson"
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

func isScalar(valueType reflect.Type) bool {
	valueType = dereferenceType(valueType)
	if valueType == nil {
		return false
	}
	if valueType == reflect.TypeOf(time.Time{}) || valueType == reflect.TypeOf(bson.ObjectID{}) {
		return true
	}

	switch valueType.Kind() {
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

var _ resourcetypes.SortValidator = SortValidator{}
