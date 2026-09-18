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
