package types

import (
	"reflect"
	"regexp"
	"time"

	resourcetypes "github.com/universe-10th/rest-resources/types"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// FilterSource derives MongoDB filter validation and serialization from a field mapping.
type FilterSource struct {
	mapping    *resourcetypes.FieldsMapping
	serializer FilterSerializer
	validator  FilterValidator
}

// NewFilterSource returns a FilterSource for the supplied MongoDB field mapping.
func NewFilterSource(mapping *resourcetypes.FieldsMapping) FilterSource {
	return FilterSource{
		mapping:    mapping,
		serializer: NewFilterSerializer(mapping),
		validator:  NewFilterValidator(mapping),
	}
}

// Serializer returns a MongoDB BSON filter serializer.
func (s FilterSource) Serializer() resourcetypes.FilterSerializer[bson.M] {
	return s.serializer
}

// Validator returns a MongoDB filter validator.
func (s FilterSource) Validator() resourcetypes.FilterValidator {
	return s.validator
}

// FilterValidator validates filter fields and values against a MongoDB field mapping.
type FilterValidator struct {
	mapping *resourcetypes.FieldsMapping
}

// NewFilterValidator returns a validator for the supplied MongoDB field mapping.
func NewFilterValidator(mapping *resourcetypes.FieldsMapping) FilterValidator {
	return FilterValidator{mapping: mapping}
}

// IsValidCmpFilter reports whether filter is a known field and value fits its Go type.
func (v FilterValidator) IsValidCmpFilter(filter string, value any) bool {
	field, ok := resourcetypes.StructFieldForJSON(v.mapping, filter)
	if !ok {
		return false
	}

	return acceptsValue(field.Type, value)
}

// IsNullCheckable reports whether filter can be checked for null.
func (v FilterValidator) IsNullCheckable(filter string) bool {
	field, ok := resourcetypes.StructFieldForJSON(v.mapping, filter)
	if !ok {
		return false
	}

	return isNullable(field.Type)
}

// IsExistenceCheckable reports whether filter can be checked for existence.
func (v FilterValidator) IsExistenceCheckable(filter string) bool {
	return resourcetypes.FieldForJSON(v.mapping, filter) != ""
}

// IsContainsCheckable reports whether filter is a string field.
func (v FilterValidator) IsContainsCheckable(filter string) bool {
	field, ok := resourcetypes.StructFieldForJSON(v.mapping, filter)
	if !ok {
		return false
	}

	return dereferenceType(field.Type).Kind() == reflect.String
}

// FilterSerializer serializes parsed filters into MongoDB BSON filters.
type FilterSerializer struct {
	mapping *resourcetypes.FieldsMapping
}

// NewFilterSerializer returns a serializer for the supplied MongoDB field mapping.
func NewFilterSerializer(mapping *resourcetypes.FieldsMapping) FilterSerializer {
	return FilterSerializer{mapping: mapping}
}

// Serialize serializes filter into a bson.M predicate.
func (s FilterSerializer) Serialize(filter resourcetypes.FilterExpression) bson.M {
	switch filter.Operator {
	case resourcetypes.FilterNone:
		return bson.M{"$expr": false}
	case resourcetypes.FilterAnd:
		return s.serializeLogical("$and", filter.Expressions)
	case resourcetypes.FilterOr:
		return s.serializeLogical("$or", filter.Expressions)
	case resourcetypes.FilterNot:
		if len(filter.Expressions) != 1 {
			return bson.M{}
		}

		return bson.M{"$nor": bson.A{s.Serialize(filter.Expressions[0])}}
	case resourcetypes.FilterLT, resourcetypes.FilterLTE, resourcetypes.FilterGT,
		resourcetypes.FilterGTE, resourcetypes.FilterEQ, resourcetypes.FilterNE:
		storageName := resourcetypes.StorageForJSON(s.mapping, filter.Field)
		if storageName == "" {
			return bson.M{}
		}

		operator := mongoOperator(filter.Operator)
		if operator == "" {
			return bson.M{}
		}
		if operator == "$eq" {
			return bson.M{storageName: filter.Value}
		}

		return bson.M{storageName: bson.M{operator: filter.Value}}
	case resourcetypes.FilterNull:
		storageName := resourcetypes.StorageForJSON(s.mapping, filter.Field)
		if storageName == "" {
			return bson.M{}
		}

		if value, _ := filter.Value.(bool); value {
			return bson.M{storageName: nil}
		}

		return bson.M{storageName: bson.M{"$ne": nil}}
	case resourcetypes.FilterExists:
		storageName := resourcetypes.StorageForJSON(s.mapping, filter.Field)
		if storageName == "" {
			return bson.M{}
		}

		value, _ := filter.Value.(bool)
		return bson.M{storageName: bson.M{"$exists": value}}
	case resourcetypes.FilterContains:
		storageName := resourcetypes.StorageForJSON(s.mapping, filter.Field)
		if storageName == "" {
			return bson.M{}
		}

		value, _ := filter.Value.(string)
		return bson.M{storageName: bson.M{"$regex": regexp.QuoteMeta(value)}}
	default:
		return bson.M{}
	}
}

func (s FilterSerializer) serializeLogical(operator string, expressions []resourcetypes.FilterExpression) bson.M {
	if len(expressions) == 0 {
		return bson.M{}
	}

	parts := make(bson.A, 0, len(expressions))
	for _, expression := range expressions {
		parts = append(parts, s.Serialize(expression))
	}

	return bson.M{operator: parts}
}

func acceptsValue(valueType reflect.Type, value any) bool {
	if value == nil {
		return isNullable(valueType)
	}

	valueType = dereferenceType(valueType)
	if valueType == reflect.TypeOf(time.Time{}) {
		_, ok := value.(time.Time)
		if ok {
			return true
		}

		_, err := time.Parse(time.RFC3339, toString(value))
		return err == nil
	}
	if valueType == reflect.TypeOf(bson.ObjectID{}) {
		_, ok := value.(bson.ObjectID)
		if ok {
			return true
		}

		_, err := bson.ObjectIDFromHex(toString(value))
		return err == nil
	}

	return resourcetypes.AcceptsScalarValue(valueType, value)
}

func isNullable(valueType reflect.Type) bool {
	if valueType == nil {
		return false
	}

	switch valueType.Kind() {
	case reflect.Pointer, reflect.Interface, reflect.Map, reflect.Slice:
		return true
	default:
		return false
	}
}

func dereferenceType(valueType reflect.Type) reflect.Type {
	for valueType != nil && valueType.Kind() == reflect.Pointer {
		valueType = valueType.Elem()
	}

	return valueType
}

func mongoOperator(operator resourcetypes.FilterOperator) string {
	switch operator {
	case resourcetypes.FilterLT:
		return "$lt"
	case resourcetypes.FilterLTE:
		return "$lte"
	case resourcetypes.FilterGT:
		return "$gt"
	case resourcetypes.FilterGTE:
		return "$gte"
	case resourcetypes.FilterEQ:
		return "$eq"
	case resourcetypes.FilterNE:
		return "$ne"
	default:
		return ""
	}
}

func toString(value any) string {
	stringValue, ok := value.(string)
	if ok {
		return stringValue
	}

	return ""
}

var (
	_ resourcetypes.FilterSource[bson.M]     = FilterSource{}
	_ resourcetypes.FilterValidator          = FilterValidator{}
	_ resourcetypes.FilterSerializer[bson.M] = FilterSerializer{}
)
