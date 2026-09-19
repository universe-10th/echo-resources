package types

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// The SortSerializer interface has a method to serialize a parsed sort expression
// into an engine-specific list. The Query type is either a bson.D (for MongoDB engine)
// or string (for GORM engine), and new engines might use their own types.
type SortSerializer[Query any] interface {
	// Serialize produces query values to be used in the underlying database engine.
	Serialize(sort SortExpression) Query
}

// The SortValidator has methods to test whether a sort can be done for
// a specific field. As of today, this sort only includes Asc and Desc.
type SortValidator interface {
	// IsSortable takes the name of a field and tells whether it is valid (for the
	// current sorting) and it can be sorted. For SQL databases, most of the fields are
	// sortable (numbers, strings, dates, incremental IDs). For MongoDB databases, most
	// of the fields are sortable (compound fields are not).
	IsSortable(field string, orderType OrderType) bool
}

// The SortSource interface is an object that produces a valid instance of SortSerializer
// and SortValidator in the same place. Engines should provide tools to spawn both of them,
// accounting for appropriate field names mapping.
type SortSource[Query any] interface {
	// Serializer spawns a sort serializer of the appropriate type.
	Serializer() SortSerializer[Query]

	// Validator spawns a sort validator of the appropriate type.
	Validator() SortValidator
}

// OrderType describes how a list result should be sorted.
type OrderType uint8

const (
	// Asc sorts a field in ascending order. It is the default zero value.
	Asc OrderType = iota

	// Desc sorts a field in descending order.
	Desc
)

// Sort describes one field ordering rule for list operations.
type Sort struct {
	Field string
	Order OrderType
}

// SortExpression is the database-neutral DSL produced by SortParser.
type SortExpression struct {
	Sort []Sort
}

// SortParser parses and validates sort specifications.
type SortParser struct {
	validator SortValidator
}

// NewSortParser returns a parser that validates field ordering with validator.
func NewSortParser(validator SortValidator) SortParser {
	return SortParser{validator: validator}
}

// Parse parses a comma-separated sort specification.
//
// Fields sort ascending by default. A leading "-" sorts the field descending.
// For example, "foo,bar,-baz" parses as foo ASC, bar ASC, baz DESC.
func (p SortParser) Parse(sentence string) (SortExpression, error) {
	if p.validator == nil {
		return SortExpression{}, errors.New("sort validator is nil")
	}

	if sentence == "" {
		return SortExpression{Sort: []Sort{}}, nil
	}

	fields := strings.Split(sentence, ",")
	sort := make([]Sort, 0, len(fields))
	for index, rawField := range fields {
		field, orderType, err := parseSortField(rawField)
		if err != nil {
			return SortExpression{}, fmt.Errorf("sort field %d: %w", index, err)
		}

		if !p.validator.IsSortable(field, orderType) {
			return SortExpression{}, fmt.Errorf("sort field %q is not allowed", field)
		}

		sort = append(sort, Sort{Field: field, Order: orderType})
	}

	return SortExpression{Sort: sort}, nil
}

func parseSortField(rawField string) (string, OrderType, error) {
	if rawField == "" {
		return "", Asc, errors.New("field is empty")
	}

	orderType := Asc
	field := rawField
	if strings.HasPrefix(rawField, "-") {
		orderType = Desc
		field = strings.TrimPrefix(rawField, "-")
	}

	if !isValidSortField(field) {
		return "", Asc, fmt.Errorf("invalid field %q", field)
	}

	return field, orderType, nil
}

func isValidSortField(field string) bool {
	return sortFieldPattern.MatchString(field)
}

var sortFieldPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_]*$`)
