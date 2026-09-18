package types

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

// The SortValidator has methods to test whether a sort can be done for
// a specific field. As of today, this sort only includes Asc and Desc.
type SortValidator interface {
	// IsSortable takes the name of a field and tells whether it is valid (for the
	// current sorting) and it can be sorted. For SQL databases, most of the fields are
	// sortable (numbers, strings, dates, incremental IDs). For MongoDB databases, most
	// of the fields are sortable (compound fields are not).
	IsSortable(field string, orderType OrderType) bool
}

// SortExpression is the database-neutral DSL produced by SortParser.
type SortExpression struct {
	Sort []Sort
}

// SortParser parses and validates serialized JSON sort specifications.
type SortParser struct {
	validator SortValidator
}
