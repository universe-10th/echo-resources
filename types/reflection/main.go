package reflection

import (
	"reflect"
	"strings"

	"github.com/universe-10th/echo-resources/types"
)

// MappingFunc stands for a function that, given an argument,
// inspects its type to extract some kind of mapping. Intended
// for struct->database fields.
type MappingFunc func(v any) map[string]string

// JSONToField maps JSON object keys to the exported struct
// field names they target. Used so the filter logic can have
// a conversion to fields (and then, later, to database fields).
// The type to inspect is retrieved from the given argument.
func JSONToField(v any) map[string]string {
	mapping := map[string]string{}

	valueType := indirectType(reflect.TypeOf(v))
	if valueType == nil || valueType.Kind() != reflect.Struct {
		return mapping
	}

	collectJSONToField(valueType, mapping)
	return mapping
}

func collectJSONToField(valueType reflect.Type, mapping map[string]string) {
	for i := range valueType.NumField() {
		field := valueType.Field(i)
		if field.PkgPath != "" {
			continue
		}

		fieldType := indirectType(field.Type)
		if field.Anonymous && field.Tag.Get("json") == "" && fieldType != nil && fieldType.Kind() == reflect.Struct {
			collectJSONToField(fieldType, mapping)
			continue
		}

		jsonName := field.Name
		if tag := field.Tag.Get("json"); tag != "" {
			tagName := strings.Split(tag, ",")[0]
			if tagName == "-" {
				continue
			}
			if tagName != "" {
				jsonName = tagName
			}
		}

		mapping[jsonName] = field.Name
	}
}

// FieldsMapping stands for the correspondence between fields in
// different scopes. This is: JSON <-> Field <-> Storage
type FieldsMapping struct {
	resourceType   reflect.Type
	idType         reflect.Type
	jsonToField    map[string]string
	fieldToJson    map[string]string
	fieldToStorage map[string]string
	storageToField map[string]string
}

// ResourceType returns the reflected resource type used to create this mapping.
func (fieldsMapping FieldsMapping) ResourceType() reflect.Type {
	return fieldsMapping.resourceType
}

// IDType returns the reflected resource ID type used to create this mapping.
func (fieldsMapping FieldsMapping) IDType() reflect.Type {
	return fieldsMapping.idType
}

// JSONToField converts a name from the JSON input into a name of
// an existing field. Empty string means no field was found.
func (fieldsMapping FieldsMapping) JSONToField(field string) string {
	return fieldsMapping.jsonToField[field]
}

// FieldToJSON converts a name of an existing field into its corresponding
// JSON field. Empty string means no field was found.
func (fieldsMapping FieldsMapping) FieldToJSON(field string) string {
	return fieldsMapping.fieldToJson[field]
}

// FieldToStorage converts a name from an existing field into its corresponding
// storage field. Empty string means no storage field was found.
func (fieldsMapping FieldsMapping) FieldToStorage(field string) string {
	return fieldsMapping.fieldToStorage[field]
}

// StorageToField converts a name of a storage field to its corresponding
// struct field. Empty string means no storage field was found.
func (fieldsMapping FieldsMapping) StorageToField(field string) string {
	return fieldsMapping.storageToField[field]
}

// FieldForJSON returns the struct field name mapped from a JSON field.
func FieldForJSON(mapping *FieldsMapping, jsonName string) string {
	if mapping == nil {
		return ""
	}

	return mapping.JSONToField(jsonName)
}

// StorageForJSON returns the storage field name mapped from a JSON field.
func StorageForJSON(mapping *FieldsMapping, jsonName string) string {
	fieldName := FieldForJSON(mapping, jsonName)
	if fieldName == "" {
		return ""
	}

	return mapping.FieldToStorage(fieldName)
}

// StructFieldForJSON returns the reflected struct field mapped from a JSON field.
func StructFieldForJSON(mapping *FieldsMapping, jsonName string) (reflect.StructField, bool) {
	fieldName := FieldForJSON(mapping, jsonName)
	if fieldName == "" {
		return reflect.StructField{}, false
	}

	return StructFieldByName(mapping.ResourceType(), fieldName)
}

// StructFieldByName returns an exported struct field by name, looking through anonymous embedded structs.
func StructFieldByName(valueType reflect.Type, fieldName string) (reflect.StructField, bool) {
	valueType = indirectType(valueType)
	if valueType == nil || valueType.Kind() != reflect.Struct {
		return reflect.StructField{}, false
	}

	if field, ok := valueType.FieldByName(fieldName); ok && field.PkgPath == "" {
		return field, true
	}

	for i := range valueType.NumField() {
		field := valueType.Field(i)
		if field.PkgPath != "" || !field.Anonymous {
			continue
		}

		fieldType := indirectType(field.Type)
		if fieldType == nil || fieldType.Kind() != reflect.Struct {
			continue
		}

		if nestedField, ok := StructFieldByName(fieldType, fieldName); ok {
			return nestedField, true
		}
	}

	return reflect.StructField{}, false
}

// IsIntegralNumber reports whether value is a numeric value without a fractional component.
func IsIntegralNumber(value any) bool {
	number, ok := NumericValue(value)
	return ok && number == float64(int64(number))
}

// NumericValue extracts common Go numeric values as float64.
func NumericValue(value any) (float64, bool) {
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

// AcceptsScalarValue reports whether value can be used for comparisons against valueType.
func AcceptsScalarValue(valueType reflect.Type, value any) bool {
	if valueType == nil {
		return false
	}

	switch valueType.Kind() {
	case reflect.String:
		_, ok := value.(string)
		return ok
	case reflect.Bool:
		_, ok := value.(bool)
		return ok
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return IsIntegralNumber(value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		number, ok := NumericValue(value)
		return ok && number >= 0 && number == float64(uint64(number))
	case reflect.Float32, reflect.Float64:
		_, ok := NumericValue(value)
		return ok
	default:
		return false
	}
}

// NewFieldsMapping creates a FieldsMapping instance from a generic resource
// type and a custom mapping function (intended per-storage-engine).
func NewFieldsMapping[IDT comparable, RT types.Resource[IDT]](customFieldToStorage MappingFunc) *FieldsMapping {
	var value RT
	jsonToField := JSONToField(value)
	fieldToJson := map[string]string{}
	for key, val := range jsonToField {
		fieldToJson[val] = key
	}
	fieldToStorage := customFieldToStorage(value)
	storageToField := map[string]string{}
	for key, val := range fieldToStorage {
		storageToField[val] = key
	}

	return &FieldsMapping{
		resourceType:   indirectType(reflect.TypeOf(value)),
		idType:         reflect.TypeOf((*IDT)(nil)).Elem(),
		jsonToField:    jsonToField,
		fieldToJson:    fieldToJson,
		fieldToStorage: fieldToStorage,
		storageToField: storageToField,
	}
}

func indirectType(valueType reflect.Type) reflect.Type {
	for valueType != nil && valueType.Kind() == reflect.Pointer {
		valueType = valueType.Elem()
	}

	return valueType
}
