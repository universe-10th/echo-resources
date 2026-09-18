package reflection

import (
	"reflect"
	"strings"

	"github.com/universe-10th/echo-resources/types"
	"github.com/universe-10th/echo-resources/types/reflection"
)

// FieldToStorage maps exported struct field names to their MongoDB BSON field names.
func FieldToStorage(v any) map[string]string {
	mapping := map[string]string{}

	valueType := indirectType(reflect.TypeOf(v))
	if valueType == nil || valueType.Kind() != reflect.Struct {
		return mapping
	}

	collectFieldToStorage(valueType, mapping)
	return mapping
}

func collectFieldToStorage(valueType reflect.Type, mapping map[string]string) {
	for i := range valueType.NumField() {
		field := valueType.Field(i)
		if field.PkgPath != "" {
			continue
		}

		name, options := parseBSONTag(field.Tag.Get("bson"))
		if name == "-" {
			continue
		}

		fieldType := indirectType(field.Type)
		if (field.Anonymous || options["inline"]) && fieldType != nil && fieldType.Kind() == reflect.Struct {
			collectFieldToStorage(fieldType, mapping)
			continue
		}

		if name == "" {
			name = strings.ToLower(field.Name)
		}

		mapping[field.Name] = name
	}
}

func indirectType(valueType reflect.Type) reflect.Type {
	for valueType != nil && valueType.Kind() == reflect.Pointer {
		valueType = valueType.Elem()
	}

	return valueType
}

func parseBSONTag(tag string) (string, map[string]bool) {
	options := map[string]bool{}
	if tag == "" {
		return "", options
	}

	parts := strings.Split(tag, ",")
	for _, option := range parts[1:] {
		if option != "" {
			options[option] = true
		}
	}

	return parts[0], options
}

// NewFieldsMapping creates a FieldsMapping instance from a generic resource
// type and a custom mapping function (intended for MongoDB).
func NewFieldsMapping[IDT comparable, RT types.Resource[IDT]]() reflection.FieldsMapping {
	return reflection.NewFieldsMapping[IDT, RT](FieldToStorage)
}
