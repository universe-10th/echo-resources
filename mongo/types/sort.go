package types

import (
	"reflect"
	"time"

	resourcetypes "github.com/universe-10th/echo-resources/types"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// SortSource derives MongoDB sort validation and serialization from a field mapping.
type SortSource struct {
	mapping    *resourcetypes.FieldsMapping
	serializer SortSerializer
	validator  SortValidator
}

// NewSortSource returns a SortSource for the supplied MongoDB field mapping.
func NewSortSource(mapping *resourcetypes.FieldsMapping) SortSource {
	return SortSource{
		mapping:    mapping,
		serializer: NewSortSerializer(mapping),
		validator:  NewSortValidator(mapping),
	}
}

// Serializer returns a MongoDB BSON sort serializer.
func (s SortSource) Serializer() resourcetypes.SortSerializer[bson.D] {
	return s.serializer
}

// Validator returns a MongoDB sort validator.
func (s SortSource) Validator() resourcetypes.SortValidator {
	return s.validator
}

// SortSerializer serializes parsed sorts into MongoDB BSON sort specifications.
type SortSerializer struct {
	mapping *resourcetypes.FieldsMapping
}

// NewSortSerializer returns a serializer for the supplied MongoDB field mapping.
func NewSortSerializer(mapping *resourcetypes.FieldsMapping) SortSerializer {
	return SortSerializer{mapping: mapping}
}

// Serialize serializes sort into a BSON sort specification.
func (s SortSerializer) Serialize(sort resourcetypes.SortExpression) bson.D {
	if len(sort.Sort) == 0 {
		return bson.D{}
	}

	result := make(bson.D, 0, len(sort.Sort))
	for _, item := range sort.Sort {
		storageName := resourcetypes.StorageForJSON(s.mapping, item.Field)
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
	mapping *resourcetypes.FieldsMapping
}

// NewSortValidator returns a sort validator for the supplied MongoDB field mapping.
func NewSortValidator(mapping *resourcetypes.FieldsMapping) SortValidator {
	return SortValidator{mapping: mapping}
}

// IsSortable reports whether field can be used in MongoDB sort expressions.
func (v SortValidator) IsSortable(field string, orderType resourcetypes.OrderType) bool {
	if !isValidOrderType(orderType) {
		return false
	}

	structField, ok := resourcetypes.StructFieldForJSON(v.mapping, field)
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
	_ resourcetypes.SortSource[bson.D]     = SortSource{}
	_ resourcetypes.SortSerializer[bson.D] = SortSerializer{}
	_ resourcetypes.SortValidator          = SortValidator{}
)
