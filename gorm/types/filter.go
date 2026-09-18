package types

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	gormreflection "github.com/universe-10th/echo-resources/gorm/types/reflection"
	resourcetypes "github.com/universe-10th/echo-resources/types"
	"gorm.io/gorm"
)

// FilterSource derives GORM filter validation and serialization from a model struct.
type FilterSource struct {
	model any
}

// NewFilterSource returns a FilterSource for the supplied GORM model.
func NewFilterSource(model any) FilterSource {
	return FilterSource{model: model}
}

// Serializer returns a GORM SQL predicate serializer.
func (s FilterSource) Serializer() resourcetypes.FilterSerializer[string] {
	return NewFilterSerializer(s.model)
}

// Validator returns a GORM filter validator.
func (s FilterSource) Validator() resourcetypes.FilterValidator {
	return NewFilterValidator(s.model)
}

// FilterValidator validates filter fields and values against a GORM model struct.
type FilterValidator struct {
	fields map[string]filterField
}

// NewFilterValidator returns a validator for the supplied GORM model.
func NewFilterValidator(model any) FilterValidator {
	return FilterValidator{fields: collectFilterFields(model)}
}

// IsValidCmpFilter reports whether filter is a known field and value fits its Go type.
func (v FilterValidator) IsValidCmpFilter(filter string, value any) bool {
	field, ok := v.fields[filter]
	if !ok {
		return false
	}

	return acceptsValue(field.Type, value)
}

// IsNullCheckable reports whether filter can be checked for SQL NULL.
func (v FilterValidator) IsNullCheckable(filter string) bool {
	field, ok := v.fields[filter]
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
	field, ok := v.fields[filter]
	if !ok {
		return false
	}

	return dereferenceType(field.Type).Kind() == reflect.String
}

// IsSortable reports whether filter can be used in SQL ORDER BY expressions.
func (v FilterValidator) IsSortable(filter string) bool {
	field, ok := v.fields[filter]
	if !ok {
		return false
	}

	return isScalar(field.Type)
}

// FilterSerializer serializes parsed filters into GORM SQL predicate strings.
type FilterSerializer struct {
	fields map[string]filterField
}

// NewFilterSerializer returns a serializer for the supplied GORM model.
func NewFilterSerializer(model any) FilterSerializer {
	return FilterSerializer{fields: collectFilterFields(model)}
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
		field, ok := s.fields[filter.Field]
		if !ok {
			return ""
		}

		return fmt.Sprintf("%s %s %s", quoteIdentifier(field.StorageName), sqlOperator(filter.Operator), sqlLiteral(filter.Value))
	case resourcetypes.FilterNull:
		field, ok := s.fields[filter.Field]
		if !ok {
			return ""
		}

		if value, _ := filter.Value.(bool); value {
			return fmt.Sprintf("%s IS NULL", quoteIdentifier(field.StorageName))
		}

		return fmt.Sprintf("%s IS NOT NULL", quoteIdentifier(field.StorageName))
	case resourcetypes.FilterExists:
		field, ok := s.fields[filter.Field]
		if !ok {
			return ""
		}

		if value, _ := filter.Value.(bool); value {
			return fmt.Sprintf("%s IS NOT NULL", quoteIdentifier(field.StorageName))
		}

		return fmt.Sprintf("%s IS NULL", quoteIdentifier(field.StorageName))
	case resourcetypes.FilterContains:
		field, ok := s.fields[filter.Field]
		if !ok {
			return ""
		}

		return fmt.Sprintf(
			"%s LIKE %s ESCAPE '\\'",
			quoteIdentifier(field.StorageName),
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

type filterField struct {
	JSONName    string
	FieldName   string
	StorageName string
	Type        reflect.Type
}

func collectFilterFields(model any) map[string]filterField {
	fields := map[string]filterField{}

	modelType := indirectType(reflect.TypeOf(model))
	if modelType == nil || modelType.Kind() != reflect.Struct {
		return fields
	}

	fieldToStorage := gormreflection.FieldToStorage(model)
	collectFilterFieldsFromType(modelType, fieldToStorage, fields)
	return fields
}

func collectFilterFieldsFromType(modelType reflect.Type, fieldToStorage map[string]string, fields map[string]filterField) {
	for i := range modelType.NumField() {
		field := modelType.Field(i)
		if field.PkgPath != "" {
			continue
		}

		tagName, skip := jsonTagName(field)
		if skip {
			continue
		}

		fieldType := indirectType(field.Type)
		if field.Anonymous && tagName == "" && fieldType != nil && fieldType.Kind() == reflect.Struct {
			collectFilterFieldsFromType(fieldType, fieldToStorage, fields)
			continue
		}

		storageName := fieldToStorage[field.Name]
		if storageName == "" {
			continue
		}

		jsonName := tagName
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

	name, _, _ := strings.Cut(tag, ",")
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

var (
	_ resourcetypes.FilterSource[string]     = FilterSource{}
	_ resourcetypes.FilterValidator          = FilterValidator{}
	_ resourcetypes.FilterSerializer[string] = FilterSerializer{}
)
