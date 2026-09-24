package utils

// FlagIndex constrains the unsigned integer types that can be used as flag
// positions in a Flags value.
type FlagIndex interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uint
}

// Flags is a compact set of enum-like flag positions.
//
// Each T value is interpreted as a zero-based bit index, not as a bit mask.
// For example, value 3 maps to bit 1<<3.
type Flags[T FlagIndex] uint64

// NewFlags creates a Flags value containing all supplied flag positions.
func NewFlags[T FlagIndex](values ...T) Flags[T] {
	var f Flags[T]

	for _, value := range values {
		f |= 1 << uint64(value)
	}

	return f
}

// Has reports whether value is present in the flag set.
func (f Flags[T]) Has(value T) bool {
	mask := Flags[T](1 << uint64(value))
	return f&mask != 0
}

// HasAll reports whether every supplied value is present in the flag set.
func (f Flags[T]) HasAll(values ...T) bool {
	for _, value := range values {
		if !f.Has(value) {
			return false
		}
	}

	return true
}

// HasAny reports whether at least one supplied value is present in the flag set.
func (f Flags[T]) HasAny(values ...T) bool {
	for _, value := range values {
		if f.Has(value) {
			return true
		}
	}

	return false
}

// Add inserts all supplied values into the flag set.
func (f *Flags[T]) Add(values ...T) {
	for _, value := range values {
		*f |= 1 << uint64(value)
	}
}

// Remove deletes all supplied values from the flag set.
func (f *Flags[T]) Remove(values ...T) {
	for _, value := range values {
		*f &^= 1 << uint64(value)
	}
}

// Toggle flips whether value is present in the flag set.
func (f *Flags[T]) Toggle(value T) {
	*f ^= 1 << uint64(value)
}

// Empty reports whether the flag set contains no values.
func (f Flags[T]) Empty() bool {
	return f == 0
}
