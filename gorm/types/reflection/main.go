package reflection

import (
	"reflect"
	"strings"
	"unicode"

	"github.com/universe-10th/echo-resources/types"
	"github.com/universe-10th/echo-resources/types/reflection"
)

// FieldToStorage maps exported struct field names to their GORM column names.
func FieldToStorage(v any) map[string]string {
	mapping := map[string]string{}

	valueType := indirectType(reflect.TypeOf(v))
	if valueType == nil || valueType.Kind() != reflect.Struct {
		return mapping
	}

	collectFieldToStorage(valueType, "", mapping)
	return mapping
}

func collectFieldToStorage(valueType reflect.Type, prefix string, mapping map[string]string) {
	for i := range valueType.NumField() {
		field := valueType.Field(i)
		if field.PkgPath != "" {
			continue
		}

		tag := parseGORMTag(field.Tag.Get("gorm"))
		if _, skip := tag["-"]; skip {
			continue
		}

		fieldType := indirectType(field.Type)
		embeddedPrefix := prefix + tag["embeddedprefix"]
		if field.Anonymous && fieldType != nil && fieldType.Kind() == reflect.Struct {
			collectFieldToStorage(fieldType, embeddedPrefix, mapping)
			continue
		}
		if _, embedded := tag["embedded"]; embedded && fieldType != nil && fieldType.Kind() == reflect.Struct {
			collectFieldToStorage(fieldType, embeddedPrefix, mapping)
			continue
		}

		column := tag["column"]
		if column == "" {
			column = toSnakeCase(field.Name)
		}

		mapping[field.Name] = prefix + column
	}
}

func indirectType(valueType reflect.Type) reflect.Type {
	for valueType != nil && valueType.Kind() == reflect.Pointer {
		valueType = valueType.Elem()
	}

	return valueType
}

func parseGORMTag(tag string) map[string]string {
	values := map[string]string{}
	for _, part := range strings.Split(tag, ";") {
		if part == "" {
			continue
		}

		key, value, ok := strings.Cut(part, ":")
		if !ok {
			values[strings.ToLower(part)] = ""
			continue
		}

		values[strings.ToLower(key)] = value
	}

	return values
}

func toSnakeCase(value string) string {
	runes := []rune(value)
	var builder strings.Builder
	for index, r := range runes {
		if index > 0 && unicode.IsUpper(r) {
			previous := runes[index-1]
			nextIsLower := index+1 < len(runes) && unicode.IsLower(runes[index+1])
			previousIsWordTail := unicode.IsLower(previous) || unicode.IsDigit(previous)
			if previousIsWordTail || nextIsLower {
				builder.WriteByte('_')
			}
		}
		if index > 0 && unicode.IsDigit(r) && !unicode.IsDigit(runes[index-1]) {
			builder.WriteByte('_')
		}
		builder.WriteRune(unicode.ToLower(r))
	}

	return builder.String()
}

// NewFieldsMapping creates a FieldsMapping instance from a generic resource
// type and a custom mapping function (intended for GORM).
func NewFieldsMapping[IDT comparable, RT types.Resource[IDT]]() *reflection.FieldsMapping {
	return reflection.NewFieldsMapping[IDT, RT](FieldToStorage)
}
