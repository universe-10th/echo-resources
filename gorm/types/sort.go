package types

import (
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	resourcetypes "github.com/universe-10th/echo-resources/types"
	resourcereflection "github.com/universe-10th/echo-resources/types/reflection"
	"gorm.io/gorm"
)

// SortSource derives GORM sort validation and serialization from a field mapping.
type SortSource struct {
	mapping    *resourcereflection.FieldsMapping
	serializer SortSerializer
	validator  SortValidator
}

// NewSortSource returns a SortSource for the supplied GORM field mapping.
func NewSortSource(mapping *resourcereflection.FieldsMapping) SortSource {
	return SortSource{
		mapping:    mapping,
		serializer: NewSortSerializer(mapping),
		validator:  NewSortValidator(mapping),
	}
}

// Serializer returns a GORM SQL order serializer.
func (s SortSource) Serializer() resourcetypes.SortSerializer[string] {
	return s.serializer
}

// Validator returns a GORM sort validator.
func (s SortSource) Validator() resourcetypes.SortValidator {
	return s.validator
}

// SortSerializer serializes parsed sorts into GORM SQL order fragments.
type SortSerializer struct {
	mapping *resourcereflection.FieldsMapping
}

// NewSortSerializer returns a serializer for the supplied GORM field mapping.
func NewSortSerializer(mapping *resourcereflection.FieldsMapping) SortSerializer {
	return SortSerializer{mapping: mapping}
}

// Serialize serializes sort into a SQL order fragment.
func (s SortSerializer) Serialize(sort resourcetypes.SortExpression) string {
	if len(sort.Sort) == 0 {
		return ""
	}

	parts := make([]string, 0, len(sort.Sort))
	for _, item := range sort.Sort {
		storageName := resourcereflection.StorageForJSON(s.mapping, item.Field)
		if storageName == "" {
			return ""
		}

		direction := sqlSortDirection(item.Order)
		if direction == "" {
			return ""
		}

		parts = append(parts, quoteIdentifier(storageName)+" "+direction)
	}

	return strings.Join(parts, ", ")
}

// SortValidator validates sort fields against a GORM field mapping.
type SortValidator struct {
	mapping *resourcereflection.FieldsMapping
}

// NewSortValidator returns a sort validator for the supplied GORM field mapping.
func NewSortValidator(mapping *resourcereflection.FieldsMapping) SortValidator {
	return SortValidator{mapping: mapping}
}

// IsSortable reports whether field can be used in SQL ORDER BY expressions.
func (v SortValidator) IsSortable(field string, orderType resourcetypes.OrderType) bool {
	if !isValidOrderType(orderType) {
		return false
	}

	structField, ok := resourcereflection.StructFieldForJSON(v.mapping, field)
	if !ok {
		return false
	}

	return isScalar(structField.Type) && !hasUnsortableGORMType(structField)
}

func isValidOrderType(orderType resourcetypes.OrderType) bool {
	switch orderType {
	case resourcetypes.Asc, resourcetypes.Desc:
		return true
	default:
		return false
	}
}

func sqlSortDirection(orderType resourcetypes.OrderType) string {
	switch orderType {
	case resourcetypes.Asc:
		return "ASC"
	case resourcetypes.Desc:
		return "DESC"
	default:
		return ""
	}
}

func isScalar(valueType reflect.Type) bool {
	valueType = dereferenceType(valueType)
	if valueType == nil {
		return false
	}
	if valueType == reflect.TypeOf(time.Time{}) || valueType == reflect.TypeOf(uuid.UUID{}) || valueType == reflect.TypeOf(gorm.DeletedAt{}) {
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

func hasUnsortableGORMType(field reflect.StructField) bool {
	tag := parseGORMTag(field.Tag.Get("gorm"))
	storageType := strings.ToLower(tag["type"])

	switch storageType {
	case "text", "tinytext", "mediumtext", "longtext",
		"blob", "tinyblob", "mediumblob", "longblob",
		"json", "jsonb":
		return true
	default:
		return false
	}
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

var (
	_ resourcetypes.SortSource[string]     = SortSource{}
	_ resourcetypes.SortSerializer[string] = SortSerializer{}
	_ resourcetypes.SortValidator          = SortValidator{}
)
