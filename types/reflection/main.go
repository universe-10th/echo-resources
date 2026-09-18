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

	valueType := reflect.TypeOf(v)
	for valueType != nil && valueType.Kind() == reflect.Pointer {
		valueType = valueType.Elem()
	}

	if valueType == nil || valueType.Kind() != reflect.Struct {
		return mapping
	}

	for i := range valueType.NumField() {
		field := valueType.Field(i)
		if field.PkgPath != "" {
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

	return mapping
}

// FieldsMapping stands for the correspondence between fields in
// different scopes. This is: JSON <-> Field <-> Storage
type FieldsMapping struct {
	jsonToField    map[string]string
	fieldToJson    map[string]string
	fieldToStorage map[string]string
	storageToField map[string]string
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
		jsonToField:    jsonToField,
		fieldToJson:    fieldToJson,
		fieldToStorage: fieldToStorage,
		storageToField: storageToField,
	}
}
