package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
)

/**
 * Filters are a special topic here. They are parsed from a serialized format which
 * looks like this:
 *
 * field  : a valid camel_cased golang identifier. It will stick to a regex. By this
 *          point, we'll assume as solved that camel_cased field names are unique and
 *          they all satisfy: [a-z0-9][a-z)-9]*.
 *
 * cmp    : "$lt" | "$lte" | "$gt" | "$gte" | "$eq" | "$ne"
 * value  : A JSON-serialized value. It will be parsed against the field. For example,
 *          a string might actually represent a time.Time value if the involved field
 *          is a time.Time field.
 *
 * Filter : {"$none": true} <--- top-level only, matches no elements
 *        | {"$and": []Filter}
 *        | {"$or": []Filter}
 *        | {"$not": Filter}
 *        | {field: {cmp: value}} <--- comparing a field to a value
 *        | {field: {"$null": true|false}} <--- testing whether the field is null
 *        | {field: {"$exists": true|false}} <--- testing whether the field exists.
 *        | {field: {"$contains": string}} <--- testing for text containment
 *
 * This is NOT extensive in what the useful filters might imply, but these are the only
 * ones we will support today.
 *
 * So, first, I need a special class we can name FilterParser. It takes the filter as
 * a map[string]any (since this may come from user input, but always in the format I
 * explained above) and produces a VALID filter or raises an error. What a VALID filter
 * is... involves:
 *
 * - The referenced fields are valid.
 * - The values provided for comparisons are valid for the respective fields.
 * - The format of each allowed clause is valid.
 */

// The FilterSerializer interface has a method to serialize a parsed filter expression
// into an engine-specific query.
// The Query type is either a map (for MongoDB engine) or string (for GORM engine), and
// new engines might use their own types.
type FilterSerializer[Query any] interface {
	// Serialize produces query values to be used in the underlying database engine.
	Serialize(filter FilterExpression) Query
}

// The FilterValidator interface has methods to tell whether the fields and values involved
// in the parsing can be effectively parsed / identified as valid data.
type FilterValidator interface {
	// IsValidCmpFilter takes the name of a field and tells whether it is valid (for the
	// current filtering) and then takes a value: the value (which is a JSON-decoded value)
	// must be interpretable by the field. This means: if the field accepts the idea of
	// being populated by the value (this might involve decoding, e.g. a time.Time field
	// being populated from a formatted string), then this method returns true. Otherwise,
	// this method must return false.
	IsValidCmpFilter(filter string, value any) bool

	// IsNullCheckable takes the name of a field and tells whether it is valid (for
	// the current filtering) for a $null check (i.e. a nullable field).
	IsNullCheckable(filter string) bool

	// IsExistenceCheckable takes the name of a field and tells whether it is valid (for
	// the current filtering) for an $exists check (i.e. the field supports existence checks
	// in the underlying engine - typically false on GORM and typically true on MongoDB,
	// regardless of the field name, unless additional constraints are set).
	IsExistenceCheckable(filter string) bool

	// IsContainsCheckable takes the name of a field and tells whether it is valid (for
	// the current filtering) for a $contains check (i.e. a string field).
	IsContainsCheckable(field string) bool
}

// The FilterSource interface is an object that produces a valid instance of FilterSerializer
// and FilterValidator in the same place. Engines should provide tools to spawn both of them,
// accounting for appropriate field names mapping.
type FilterSource[Query any] interface {
	// Serializer spawns a filter serializer of the appropriate type.
	Serializer() FilterSerializer[Query]

	// Validator spawns a filter validator of the appropriate type.
	Validator() FilterValidator
}

// FilterOperator is the parsed operator for a filter expression.
type FilterOperator string

const (
	// FilterNone matches no elements. It is accepted only as a top-level filter.
	FilterNone FilterOperator = operatorNone

	// FilterAnd joins all child expressions with logical AND.
	FilterAnd FilterOperator = operatorAnd

	// FilterOr joins all child expressions with logical OR.
	FilterOr FilterOperator = operatorOr

	// FilterNot negates its single child expression.
	FilterNot FilterOperator = operatorNot

	// FilterLT compares a field with less-than semantics.
	FilterLT FilterOperator = operatorLT

	// FilterLTE compares a field with less-than-or-equal semantics.
	FilterLTE FilterOperator = operatorLTE

	// FilterGT compares a field with greater-than semantics.
	FilterGT FilterOperator = operatorGT

	// FilterGTE compares a field with greater-than-or-equal semantics.
	FilterGTE FilterOperator = operatorGTE

	// FilterEQ compares a field with equality semantics.
	FilterEQ FilterOperator = operatorEQ

	// FilterNE compares a field with non-equality semantics.
	FilterNE FilterOperator = operatorNE

	// FilterNull checks whether a field is null.
	FilterNull FilterOperator = operatorNull

	// FilterExists checks whether a field exists.
	FilterExists FilterOperator = operatorExists

	// FilterContains checks whether a field contains text.
	FilterContains FilterOperator = operatorContains
)

// FilterExpression is the database-neutral DSL produced by FilterParser.
//
// FilterNone has no field, value, or nested expressions. For logical operators,
// Expressions contains the nested filters. FilterNot always contains exactly one
// nested expression. For field operators, Field contains the camel-cased field name
// and Value contains the JSON-decoded comparison/check value.
type FilterExpression struct {
	Operator    FilterOperator
	Field       string
	Value       any
	Expressions []FilterExpression
}

// Restrict applies another filter to this filter in-place using AND semantics.
//
// A FilterNone receiver already matches no elements, so further restrictions do
// not change it. A FilterNone restriction turns this filter into an empty
// FilterNone expression. When either side is an AND expression, Restrict keeps
// the resulting tree flat by merging or appending AND children instead of
// creating nested AND expressions.
func (filter *FilterExpression) Restrict(restriction *FilterExpression) {
	if filter == nil || restriction == nil || restriction.Operator == "" {
		return
	}

	if filter.Operator == FilterNone {
		return
	}

	if restriction.Operator == FilterNone {
		*filter = FilterExpression{Operator: FilterNone}
		return
	}

	if filter.Operator == "" {
		*filter = *restriction
		return
	}

	if filter.Operator == FilterAnd && restriction.Operator == FilterAnd {
		filter.Expressions = append(filter.Expressions, restriction.Expressions...)
		return
	}

	if filter.Operator == FilterAnd {
		filter.Expressions = append(filter.Expressions, *restriction)
		return
	}

	if restriction.Operator == FilterAnd {
		current := *filter
		*filter = *restriction
		filter.Expressions = append(filter.Expressions, current)
		return
	}

	*filter = FilterExpression{
		Operator:    FilterAnd,
		Expressions: []FilterExpression{*filter, *restriction},
	}
}

// FilterParser parses and validates serialized JSON filter specifications.
type FilterParser struct {
	validator FilterValidator
}

// NewFilterParser returns a parser that validates field operations with validator.
func NewFilterParser(validator FilterValidator) FilterParser {
	return FilterParser{validator: validator}
}

// Parse decodes one JSON filter specification and returns its validated filter expression.
func (p FilterParser) Parse(decoder *json.Decoder) (FilterExpression, error) {
	if decoder == nil {
		return FilterExpression{}, errors.New("filter decoder is nil")
	}

	if p.validator == nil {
		return FilterExpression{}, errors.New("filter validator is nil")
	}

	var raw any
	if err := decoder.Decode(&raw); err != nil {
		return FilterExpression{}, fmt.Errorf("decode filter: %w", err)
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return FilterExpression{}, errors.New("filter must contain exactly one JSON value")
		}

		return FilterExpression{}, fmt.Errorf("decode filter: %w", err)
	}

	return p.parseFilter(raw, true)
}

func (p FilterParser) parseFilter(raw any, allowNone bool) (FilterExpression, error) {
	object, ok := raw.(map[string]any)
	if !ok {
		return FilterExpression{}, fmt.Errorf("filter must be an object, got %T", raw)
	}

	if len(object) != 1 {
		return FilterExpression{}, fmt.Errorf("filter object must contain exactly one clause, got %d", len(object))
	}

	for key, value := range object {
		switch key {
		case operatorNone:
			return p.parseNone(value, allowNone)
		case operatorAnd, operatorOr:
			return p.parseLogicalList(key, value)
		case operatorNot:
			return p.parseLogicalNot(value)
		default:
			return p.parseFieldFilter(key, value)
		}
	}

	return FilterExpression{}, errors.New("filter object is empty")
}

func (p FilterParser) parseNone(raw any, allowNone bool) (FilterExpression, error) {
	if !allowNone {
		return FilterExpression{}, fmt.Errorf("%s filter is only allowed at the top level", operatorNone)
	}

	value, ok := raw.(bool)
	if !ok || !value {
		return FilterExpression{}, fmt.Errorf("%s filter must be true", operatorNone)
	}

	return FilterExpression{Operator: FilterNone}, nil
}

func (p FilterParser) parseLogicalList(operator string, raw any) (FilterExpression, error) {
	values, ok := raw.([]any)
	if !ok {
		return FilterExpression{}, fmt.Errorf("%s filter must be an array", operator)
	}

	if len(values) == 0 {
		return FilterExpression{}, fmt.Errorf("%s filter must contain at least one nested filter", operator)
	}

	expressions := make([]FilterExpression, 0, len(values))
	for index, value := range values {
		filter, err := p.parseFilter(value, false)
		if err != nil {
			return FilterExpression{}, fmt.Errorf("%s filter item %d: %w", operator, index, err)
		}

		expressions = append(expressions, filter)
	}

	return FilterExpression{Operator: FilterOperator(operator), Expressions: expressions}, nil
}

func (p FilterParser) parseLogicalNot(raw any) (FilterExpression, error) {
	filter, err := p.parseFilter(raw, false)
	if err != nil {
		return FilterExpression{}, fmt.Errorf("%s filter: %w", operatorNot, err)
	}

	return FilterExpression{Operator: FilterNot, Expressions: []FilterExpression{filter}}, nil
}

func (p FilterParser) parseFieldFilter(field string, raw any) (FilterExpression, error) {
	if !isValidFilterField(field) {
		return FilterExpression{}, fmt.Errorf("invalid filter field %q", field)
	}

	operation, ok := raw.(map[string]any)
	if !ok {
		return FilterExpression{}, fmt.Errorf("filter field %q must contain an operation object", field)
	}

	if len(operation) != 1 {
		return FilterExpression{}, fmt.Errorf("filter field %q must contain exactly one operation, got %d", field, len(operation))
	}

	for operator, value := range operation {
		switch {
		case isComparisonOperator(operator):
			if !p.validator.IsValidCmpFilter(field, value) {
				return FilterExpression{}, fmt.Errorf("comparison filter %q on field %q is not allowed", operator, field)
			}

			return FilterExpression{Operator: FilterOperator(operator), Field: field, Value: value}, nil
		case operator == operatorNull:
			boolValue, ok := value.(bool)
			if !ok {
				return FilterExpression{}, fmt.Errorf("%s filter on field %q must be boolean", operatorNull, field)
			}

			if !p.validator.IsNullCheckable(field) {
				return FilterExpression{}, fmt.Errorf("%s filter on field %q is not allowed", operatorNull, field)
			}

			return FilterExpression{Operator: FilterNull, Field: field, Value: boolValue}, nil
		case operator == operatorExists:
			boolValue, ok := value.(bool)
			if !ok {
				return FilterExpression{}, fmt.Errorf("%s filter on field %q must be boolean", operatorExists, field)
			}

			if !p.validator.IsExistenceCheckable(field) {
				return FilterExpression{}, fmt.Errorf("%s filter on field %q is not allowed", operatorExists, field)
			}

			return FilterExpression{Operator: FilterExists, Field: field, Value: boolValue}, nil
		case operator == operatorContains:
			stringValue, ok := value.(string)
			if !ok {
				return FilterExpression{}, fmt.Errorf("%s filter on field %q must be string", operatorContains, field)
			}

			if !p.validator.IsContainsCheckable(field) {
				return FilterExpression{}, fmt.Errorf("%s filter on field %q is not allowed", operatorContains, field)
			}

			return FilterExpression{Operator: FilterContains, Field: field, Value: stringValue}, nil
		default:
			return FilterExpression{}, fmt.Errorf("unsupported filter operator %q on field %q", operator, field)
		}
	}

	return FilterExpression{}, fmt.Errorf("filter field %q operation object is empty", field)
}

func isValidFilterField(field string) bool {
	return filterFieldPattern.MatchString(field)
}

func isComparisonOperator(operator string) bool {
	switch operator {
	case operatorLT, operatorLTE, operatorGT, operatorGTE, operatorEQ, operatorNE:
		return true
	default:
		return false
	}
}

const (
	operatorNone     = "$none"
	operatorAnd      = "$and"
	operatorOr       = "$or"
	operatorNot      = "$not"
	operatorLT       = "$lt"
	operatorLTE      = "$lte"
	operatorGT       = "$gt"
	operatorGTE      = "$gte"
	operatorEQ       = "$eq"
	operatorNE       = "$ne"
	operatorNull     = "$null"
	operatorExists   = "$exists"
	operatorContains = "$contains"
)

var filterFieldPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_]*$`)
