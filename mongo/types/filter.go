package types

import (
	"reflect"
	"regexp"
	"time"

	mongoreflection "github.com/universe-10th/echo-resources/mongo/types/reflection"
	resourcetypes "github.com/universe-10th/echo-resources/types"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// FilterSource derives MongoDB filter validation and serialization from a document struct.
type FilterSource struct {
	document any
}

// NewFilterSource returns a FilterSource for the supplied MongoDB document.
func NewFilterSource(document any) FilterSource {
	return FilterSource{document: document}
}

// Serializer returns a MongoDB BSON filter serializer.
func (s FilterSource) Serializer() resourcetypes.FilterSerializer[bson.M] {
	return NewFilterSerializer(s.document)
}

// Validator returns a MongoDB filter validator.
func (s FilterSource) Validator() resourcetypes.FilterValidator {
	return NewFilterValidator(s.document)
}

// FilterValidator validates filter fields and values against a MongoDB document struct.
type FilterValidator struct {
	fields map[string]filterField
}

// NewFilterValidator returns a validator for the supplied MongoDB document.
func NewFilterValidator(document any) FilterValidator {
	return FilterValidator{fields: collectFilterFields(document)}
}

// IsValidCmpFilter reports whether filter is a known field and value fits its Go type.
func (v FilterValidator) IsValidCmpFilter(filter string, value any) bool {
	field, ok := v.fields[filter]
	if !ok {
		return false
	}

	return acceptsValue(field.Type, value)
}

// IsNullCheckable reports whether filter can be checked for null.
func (v FilterValidator) IsNullCheckable(filter string) bool {
	field, ok := v.fields[filter]
	if !ok {
		return false
	}

	return isNullable(field.Type)
}

// IsExistenceCheckable reports whether filter can be checked for existence.
func (v FilterValidator) IsExistenceCheckable(filter string) bool {
	_, ok := v.fields[filter]
	return ok
}

// IsContainsCheckable reports whether filter is a string field.
func (v FilterValidator) IsContainsCheckable(filter string) bool {
	field, ok := v.fields[filter]
	if !ok {
		return false
	}

	return dereferenceType(field.Type).Kind() == reflect.String
}

// IsSortable reports whether filter can be used in MongoDB sort expressions.
func (v FilterValidator) IsSortable(filter string) bool {
	field, ok := v.fields[filter]
	if !ok {
		return false
	}

	return isScalar(field.Type)
}

// FilterSerializer serializes parsed filters into MongoDB BSON filters.
type FilterSerializer struct {
	fields map[string]filterField
}

// NewFilterSerializer returns a serializer for the supplied MongoDB document.
func NewFilterSerializer(document any) FilterSerializer {
	return FilterSerializer{fields: collectFilterFields(document)}
}

// Serialize serializes filter into a bson.M predicate.
func (s FilterSerializer) Serialize(filter resourcetypes.FilterExpression) bson.M {
	switch filter.Operator {
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
		field, ok := s.fields[filter.Field]
		if !ok {
			return bson.M{}
		}

		operator := mongoOperator(filter.Operator)
		if operator == "" {
			return bson.M{}
		}
		if operator == "$eq" {
			return bson.M{field.StorageName: filter.Value}
		}

		return bson.M{field.StorageName: bson.M{operator: filter.Value}}
	case resourcetypes.FilterNull:
		field, ok := s.fields[filter.Field]
		if !ok {
			return bson.M{}
		}

		if value, _ := filter.Value.(bool); value {
			return bson.M{field.StorageName: nil}
		}

		return bson.M{field.StorageName: bson.M{"$ne": nil}}
	case resourcetypes.FilterExists:
		field, ok := s.fields[filter.Field]
		if !ok {
			return bson.M{}
		}

		value, _ := filter.Value.(bool)
		return bson.M{field.StorageName: bson.M{"$exists": value}}
	case resourcetypes.FilterContains:
		field, ok := s.fields[filter.Field]
		if !ok {
			return bson.M{}
		}

		value, _ := filter.Value.(string)
		return bson.M{field.StorageName: bson.M{"$regex": regexp.QuoteMeta(value)}}
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

type filterField struct {
	JSONName    string
	FieldName   string
	StorageName string
	Type        reflect.Type
}

func collectFilterFields(document any) map[string]filterField {
	fields := map[string]filterField{}

	documentType := indirectType(reflect.TypeOf(document))
	if documentType == nil || documentType.Kind() != reflect.Struct {
		return fields
	}

	fieldToStorage := mongoreflection.FieldToStorage(document)
	collectFilterFieldsFromType(documentType, fieldToStorage, fields)
	return fields
}

func collectFilterFieldsFromType(documentType reflect.Type, fieldToStorage map[string]string, fields map[string]filterField) {
	for i := range documentType.NumField() {
		field := documentType.Field(i)
		if field.PkgPath != "" {
			continue
		}

		jsonName, skip := jsonTagName(field)
		if skip {
			continue
		}

		fieldType := indirectType(field.Type)
		if field.Anonymous && jsonName == "" && fieldType != nil && fieldType.Kind() == reflect.Struct {
			collectFilterFieldsFromType(fieldType, fieldToStorage, fields)
			continue
		}

		storageName := fieldToStorage[field.Name]
		if storageName == "" {
			continue
		}

		if jsonName == "" {
			jsonName = field.Name
		}

		fields[jsonName] = filterField{
			JSONName:    jsonName,
			FieldName:   field.Name,
			StorageName: storageName,
			Type:        field.Type,
		}
	}
}

func jsonTagName(field reflect.StructField) (string, bool) {
	tag := field.Tag.Get("json")
	if tag == "" {
		return "", false
	}

	name, _, _ := stringsCut(tag, ",")
	return name, name == "-"
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

	switch valueType.Kind() {
	case reflect.String:
		_, ok := value.(string)
		return ok
	case reflect.Bool:
		_, ok := value.(bool)
		return ok
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return isIntegralNumber(value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		number, ok := numericValue(value)
		return ok && number >= 0 && number == float64(uint64(number))
	case reflect.Float32, reflect.Float64:
		_, ok := numericValue(value)
		return ok
	default:
		return false
	}
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

func dereferenceType(valueType reflect.Type) reflect.Type {
	for valueType != nil && valueType.Kind() == reflect.Pointer {
		valueType = valueType.Elem()
	}

	return valueType
}

func indirectType(valueType reflect.Type) reflect.Type {
	return dereferenceType(valueType)
}

func isIntegralNumber(value any) bool {
	number, ok := numericValue(value)
	return ok && number == float64(int64(number))
}

func numericValue(value any) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int8:
		return float64(typed), true
	case int16:
		return float64(typed), true
	case int32:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case uint:
		return float64(typed), true
	case uint8:
		return float64(typed), true
	case uint16:
		return float64(typed), true
	case uint32:
		return float64(typed), true
	case uint64:
		return float64(typed), true
	default:
		return 0, false
	}
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

func stringsCut(value string, separator string) (string, string, bool) {
	for i := 0; i+len(separator) <= len(value); i++ {
		if value[i:i+len(separator)] == separator {
			return value[:i], value[i+len(separator):], true
		}
	}

	return value, "", false
}

var (
	_ resourcetypes.FilterSource[bson.M]     = FilterSource{}
	_ resourcetypes.FilterValidator          = FilterValidator{}
	_ resourcetypes.FilterSerializer[bson.M] = FilterSerializer{}
)
