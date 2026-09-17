package types

// ErrorCode stands for the supported error codes, intended
// to tell about the type of error and to serve as HTTP
// response codes.
type ErrorCode uint16

const (
	// ErrNotFound is used when an element was not found.
	ErrNotFound ErrorCode = 404

	// ErrConflict is used when a constraint is violated
	// (e.g. a key already in use), when a singleton in
	// a specific scope already has a non-deleted value,
	// or when an object is not in the proper state (e.g.
	// trying to prune or restore a non-deleted object).
	ErrConflict ErrorCode = 409

	// ErrInvalid is used when the data being processed
	// is somehow not valid. This relates to field errors,
	// more than conflict errors.
	ErrInvalid ErrorCode = 422

	// ErrInternal depicts an arbitrary error. Its details
	// are never meant to reach the end user.
	ErrInternal ErrorCode = 500
)

type Error interface {
	error
	Code() ErrorCode
}

// NotFoundError stands for when an element does not exist.
type NotFoundError[IDT comparable] struct {
	ElementName string `json:"element_name"`
	Key         IDT    `json:"key"`
}

// Error implements the error interface in NotFoundError.
func (e NotFoundError[IDT]) Error() string {
	return "element not found"
}

// Code implements the error code for the Error interface in NotFoundError,
// returning ErrNotFound.
func (e NotFoundError[IDT]) Code() ErrorCode {
	return ErrNotFound
}

// NotDeletedError stands for when an element is not soft-deleted
// and a soft-deleted-requiring operation tries to take place on it.
type NotDeletedError struct{}

// Error implements the error interface in NotDeletedError.
func (e NotDeletedError) Error() string {
	return "element not deleted"
}

// Code implements the error code for the Error interface in NotDeletedError,
// returning ErrConflict.
func (e NotDeletedError) Code() ErrorCode {
	return ErrConflict
}

// AlreadyUsedError stands for when a duplicate key error occurs.
// Both the key fields and values array must have the same, non-zero, length.
type AlreadyUsedError struct {
	KeyFields []any `json:"key_fields"`
	Values    []any `json:"values"`
}

// Error implements the error interface in AlreadyUsedError.
func (e AlreadyUsedError) Error() string {
	return "already used"
}

// Code implements the error code for the Error interface in AlreadyUsedError,
// returning ErrConflict.
func (e AlreadyUsedError) Code() ErrorCode {
	return ErrConflict
}

// ValidationError stands for when fields are invalid on create / patch.
type ValidationError struct {
	Errors map[string]any `json:"errors"`
}

// Error implements the error interface in ValidationError.
func (e ValidationError) Error() string {
	return "invalid data"
}

// Code implements the error code for the Error interface in ValidationError.
func (e ValidationError) Code() ErrorCode {
	return ErrInvalid
}

// InternalError stands for when an internal / unexpected error occurs.
type InternalError struct{}

// Error implements the error interface in InternalError.
func (e InternalError) Error() string {
	return "internal error"
}

// Code implements the error code for the Error interface in InternalError.
func (e InternalError) Code() ErrorCode {
	return ErrInternal
}

var (
	_ Error = NotFoundError[string]{}
	_ Error = NotDeletedError{}
	_ Error = AlreadyUsedError{
		KeyFields: nil,
		Values:    nil,
	}
	_ Error = ValidationError{
		Errors: nil,
	}
	_ Error = InternalError{}
)
