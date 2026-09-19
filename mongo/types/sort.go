package types

import (
	"reflect"
	"time"

	resourcetypes "github.com/universe-10th/echo-resources/types"
	resourcereflection "github.com/universe-10th/echo-resources/types/reflection"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// SortSerializer serializes parsed sorts into MongoDB BSON sort specifications.
type SortSerializer struct {
	mapping *resourcereflection.FieldsMapping
}

// NewSortSerializer returns a serializer for the supplied MongoDB field mapping.
func NewSortSerializer(mapping *resourcereflection.FieldsMapping) SortSerializer {
	return SortSerializer{mapping: mapping}
}

// Serialize serializes sort into a BSON sort specification.
func (s SortSerializer) Serialize(sort resourcetypes.SortExpression) bson.D {
	if len(sort.Sort) == 0 {
		return bson.D{}
	}

	result := make(bson.D, 0, len(sort.Sort))
	for _, item := range sort.Sort {
		storageName := resourcereflection.StorageForJSON(s.mapping, item.Field)
		if storageName == "" {
			return bson.D{}
		}

		direction := mongoSortDirection(item.Order)
		if direction == 0 {
			return bson.D{}
		}

		result = append(result, bson.E{Key: storageName, Value: direction})
	}

	return result
}

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

func mongoSortDirection(orderType resourcetypes.OrderType) int {
	switch orderType {
	case resourcetypes.Asc:
		return 1
	case resourcetypes.Desc:
		return -1
	default:
		return 0
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

var (
	_ resourcetypes.SortSerializer[bson.D] = SortSerializer{}
	_ resourcetypes.SortValidator          = SortValidator{}
)
