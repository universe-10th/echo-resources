package reflection

import (
	"reflect"
	"sync"

	"github.com/universe-10th/echo-resources/types"
	"github.com/universe-10th/echo-resources/types/reflection"
	"gorm.io/gorm/schema"
)

// FieldToStorage maps exported struct field names to their GORM column names.
func FieldToStorage(v any) map[string]string {
	mapping := map[string]string{}

	valueType := indirectType(reflect.TypeOf(v))
	if valueType == nil || valueType.Kind() != reflect.Struct {
		return mapping
	}

	parsed, err := schema.Parse(reflect.New(valueType).Interface(), &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		return mapping
	}

	for _, field := range parsed.Fields {
		if field.DBName == "" {
			continue
		}

		mapping[field.Name] = field.DBName
	}

	return mapping
}

func indirectType(valueType reflect.Type) reflect.Type {
	for valueType != nil && valueType.Kind() == reflect.Pointer {
		valueType = valueType.Elem()
	}

	return valueType
}

// NewFieldsMapping creates a FieldsMapping instance from a generic resource
// type and a custom mapping function (intended for GORM).
func NewFieldsMapping[IDT comparable, RT types.Resource[IDT]]() *reflection.FieldsMapping {
	return reflection.NewFieldsMapping[IDT, RT](FieldToStorage)
}
