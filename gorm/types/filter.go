package types

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	resourcetypes "github.com/universe-10th/echo-resources/types"
	resourcereflection "github.com/universe-10th/echo-resources/types/reflection"
	"gorm.io/gorm"
)

// FilterSource derives GORM filter validation and serialization from a field mapping.
type FilterSource struct {
	mapping    *resourcereflection.FieldsMapping
	serializer FilterSerializer
	validator  FilterValidator
}

// NewFilterSource returns a FilterSource for the supplied GORM field mapping.
func NewFilterSource(mapping *resourcereflection.FieldsMapping) FilterSource {
	return FilterSource{
		mapping:    mapping,
		serializer: NewFilterSerializer(mapping),
		validator:  NewFilterValidator(mapping),
	}
}

// Serializer returns a GORM SQL predicate serializer.
func (s FilterSource) Serializer() resourcetypes.FilterSerializer[string] {
	return s.serializer
}

// Validator returns a GORM filter validator.
func (s FilterSource) Validator() resourcetypes.FilterValidator {
	return s.validator
}

// FilterValidator validates filter fields and values against a GORM field mapping.
type FilterValidator struct {
	mapping *resourcereflection.FieldsMapping
}

// NewFilterValidator returns a validator for the supplied GORM field mapping.
func NewFilterValidator(mapping *resourcereflection.FieldsMapping) FilterValidator {
	return FilterValidator{mapping: mapping}
}

// IsValidCmpFilter reports whether filter is a known field and value fits its Go type.
func (v FilterValidator) IsValidCmpFilter(filter string, value any) bool {
	field, ok := resourcereflection.StructFieldForJSON(v.mapping, filter)
	if !ok {
		return false
	}

	return acceptsValue(field.Type, value)
}

// IsNullCheckable reports whether filter can be checked for SQL NULL.
func (v FilterValidator) IsNullCheckable(filter string) bool {
	field, ok := resourcereflection.StructFieldForJSON(v.mapping, filter)
	if !ok {
		return false
	}

	return isNullable(field.Type)
}

// IsExistenceCheckable reports whether filter supports an existence check.
func (v FilterValidator) IsExistenceCheckable(string) bool {
	return false
}

// IsContainsCheckable reports whether filter is a string field.
func (v FilterValidator) IsContainsCheckable(filter string) bool {
	field, ok := resourcereflection.StructFieldForJSON(v.mapping, filter)
	if !ok {
		return false
	}

	return dereferenceType(field.Type).Kind() == reflect.String
}

// IsSortable reports whether filter can be used in SQL ORDER BY expressions.
func (v FilterValidator) IsSortable(filter string) bool {
	field, ok := resourcereflection.StructFieldForJSON(v.mapping, filter)
	if !ok {
		return false
	}

	return isScalar(field.Type) && !hasUnsortableGORMType(field)
}

// FilterSerializer serializes parsed filters into GORM SQL predicate strings.
type FilterSerializer struct {
	mapping *resourcereflection.FieldsMapping
}

// NewFilterSerializer returns a serializer for the supplied GORM field mapping.
func NewFilterSerializer(mapping *resourcereflection.FieldsMapping) FilterSerializer {
	return FilterSerializer{mapping: mapping}
}

// Serialize serializes filter into a SQL predicate string.
func (s FilterSerializer) Serialize(filter resourcetypes.FilterExpression) string {
	switch filter.Operator {
	case resourcetypes.FilterAnd:
		return s.serializeLogical("AND", filter.Expressions)
	case resourcetypes.FilterOr:
		return s.serializeLogical("OR", filter.Expressions)
	case resourcetypes.FilterNot:
		if len(filter.Expressions) != 1 {
			return ""
		}

		return fmt.Sprintf("NOT (%s)", s.Serialize(filter.Expressions[0]))
	case resourcetypes.FilterLT, resourcetypes.FilterLTE, resourcetypes.FilterGT,
		resourcetypes.FilterGTE, resourcetypes.FilterEQ, resourcetypes.FilterNE:
		storageName := resourcereflection.StorageForJSON(s.mapping, filter.Field)
		if storageName == "" {
			return ""
		}

		return fmt.Sprintf("%s %s %s", quoteIdentifier(storageName), sqlOperator(filter.Operator), sqlLiteral(filter.Value))
	case resourcetypes.FilterNull:
		storageName := resourcereflection.StorageForJSON(s.mapping, filter.Field)
		if storageName == "" {
			return ""
		}

		if value, _ := filter.Value.(bool); value {
			return fmt.Sprintf("%s IS NULL", quoteIdentifier(storageName))
		}

		return fmt.Sprintf("%s IS NOT NULL", quoteIdentifier(storageName))
	case resourcetypes.FilterExists:
		storageName := resourcereflection.StorageForJSON(s.mapping, filter.Field)
		if storageName == "" {
			return ""
		}

		if value, _ := filter.Value.(bool); value {
			return fmt.Sprintf("%s IS NOT NULL", quoteIdentifier(storageName))
		}

		return fmt.Sprintf("%s IS NULL", quoteIdentifier(storageName))
	case resourcetypes.FilterContains:
		storageName := resourcereflection.StorageForJSON(s.mapping, filter.Field)
		if storageName == "" {
			return ""
		}

		return fmt.Sprintf(
			"%s LIKE %s ESCAPE '\\'",
			quoteIdentifier(storageName),
			sqlLiteral("%"+escapeLike(fmt.Sprint(filter.Value))+"%"),
		)
	default:
		return ""
	}
}

func (s FilterSerializer) serializeLogical(operator string, expressions []resourcetypes.FilterExpression) string {
	if len(expressions) == 0 {
		return ""
	}

	parts := make([]string, 0, len(expressions))
	for _, expression := range expressions {
		serialized := s.Serialize(expression)
		if serialized == "" {
			return ""
		}

		parts = append(parts, fmt.Sprintf("(%s)", serialized))
	}

	return strings.Join(parts, " "+operator+" ")
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

		_, err := time.Parse(time.RFC3339, fmt.Sprint(value))
		return err == nil
	}
	if valueType == reflect.TypeOf(uuid.UUID{}) {
		_, ok := value.(uuid.UUID)
		if ok {
			return true
		}

		_, err := uuid.Parse(fmt.Sprint(value))
		return err == nil
	}
	if valueType == reflect.TypeOf(gorm.DeletedAt{}) {
		return acceptsValue(reflect.TypeOf(time.Time{}), value)
	}

	return resourcereflection.AcceptsScalarValue(valueType, value)
}

func isNullable(valueType reflect.Type) bool {
	if valueType == nil {
		return false
	}
	if valueType == reflect.TypeOf(gorm.DeletedAt{}) {
		return true
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

func dereferenceType(valueType reflect.Type) reflect.Type {
	for valueType != nil && valueType.Kind() == reflect.Pointer {
		valueType = valueType.Elem()
	}

	return valueType
}

func sqlOperator(operator resourcetypes.FilterOperator) string {
	switch operator {
	case resourcetypes.FilterLT:
		return "<"
	case resourcetypes.FilterLTE:
		return "<="
	case resourcetypes.FilterGT:
		return ">"
	case resourcetypes.FilterGTE:
		return ">="
	case resourcetypes.FilterEQ:
		return "="
	case resourcetypes.FilterNE:
		return "<>"
	default:
		return ""
	}
}

func sqlLiteral(value any) string {
	switch typed := value.(type) {
	case nil:
		return "NULL"
	case string:
		return "'" + strings.ReplaceAll(typed, "'", "''") + "'"
	case bool:
		if typed {
			return "TRUE"
		}

		return "FALSE"
	case time.Time:
		return sqlLiteral(typed.Format(time.RFC3339Nano))
	case fmt.Stringer:
		return sqlLiteral(typed.String())
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(typed), 'f', -1, 32)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprint(typed)
	default:
		return sqlLiteral(fmt.Sprint(typed))
	}
}

func quoteIdentifier(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}

func escapeLike(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `%`, `\%`)
	value = strings.ReplaceAll(value, `_`, `\_`)
	return value
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
	_ resourcetypes.FilterSource[string]     = FilterSource{}
	_ resourcetypes.FilterValidator          = FilterValidator{}
	_ resourcetypes.FilterSerializer[string] = FilterSerializer{}
)
