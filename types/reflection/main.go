package reflection

import (
	"reflect"
	"strings"
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

	value := reflect.ValueOf(v)
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return mapping
		}

		value = value.Elem()
	}

	if value.Kind() != reflect.Struct {
		return mapping
	}

	valueType := value.Type()
	for i := range value.NumField() {
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
