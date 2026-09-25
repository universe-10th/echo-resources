package types

import (
	"reflect"
	"strings"
)

// ErrorCode stands for the supported error codes, intended
// to tell about the type of error and to serve as HTTP
// response codes.
type ErrorCode uint16

const (
	// ErrBadRequest is used when the JSON parsing did not
	// succeed, or when Content-Type is not application/json.
	ErrBadRequest ErrorCode = 400

	// ErrUnauthorized is used when the request is not authorized.
	ErrUnauthorized ErrorCode = 401

	// ErrForbidden is used when the request is forbidden.
	ErrForbidden ErrorCode = 403

	// ErrNotFound is used when an element was not found.
	ErrNotFound ErrorCode = 404

	// ErrNotAcceptable is used when the client does not have
	// application/json in its Accept header.
	ErrNotAcceptable ErrorCode = 406

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

	// ErrThrottled is used when there were too many requests.
	ErrThrottled ErrorCode = 429

	// ErrInternal depicts an arbitrary error. Its details
	// are never meant to reach the end user.
	ErrInternal ErrorCode = 500
)

type Error interface {
	error
	Code() ErrorCode
}

// BadRequestError stands for when the request does not include
// a body with application/json content-type and valid JSON
// contents in the body.
type BadRequestError struct{}

// Error implements the error interface in BadRequestError.
func (e BadRequestError) Error() string {
	return "bad request"
}

// Code implements the error code for the Error interface in BadRequestError.
func (e BadRequestError) Code() ErrorCode {
	return ErrBadRequest
}

// UnauthorizedError stands for when the request is not
// authorized, and requires authorization.
type UnauthorizedError struct{}

// Error implements the error interface in UnauthorizedError.
func (e UnauthorizedError) Error() string {
	return "unauthorized"
}

// Code implements the error code for the Error interface in UnauthorizedError.
func (e UnauthorizedError) Code() ErrorCode {
	return ErrUnauthorized
}

// ForbiddenError stands for when the request is not
// authorized, and requires authorization.
type ForbiddenError struct{}

// Error implements the error interface in ForbiddenError.
func (e ForbiddenError) Error() string {
	return "forbidden"
}

// Code implements the error code for the Error interface in ForbiddenError.
func (e ForbiddenError) Code() ErrorCode {
	return ErrForbidden
}

// InvalidIDError stands for when a parsed id is invalid.
type InvalidIDError struct {
	ElementName string `json:"element_name"`
	Key         string `json:"key"`
}

// Error implements the error interface in InvalidIDError.
func (e InvalidIDError) Error() string {
	return "element not found"
}

// Code implements the error code for the Error interface in NotFoundError,
// returning InvalidIDError.
func (e InvalidIDError) Code() ErrorCode {
	return ErrNotFound
}

// SingletonNotFoundError stands for when a parsed id is invalid.
type SingletonNotFoundError struct {
	ElementName string `json:"element_name"`
}

// Error implements the error interface in SingletonNotFoundError.
func (e SingletonNotFoundError) Error() string {
	return "element not found"
}

// Code implements the error code for the Error interface in NotFoundError,
// returning SingletonNotFoundError.
func (e SingletonNotFoundError) Code() ErrorCode {
	return ErrNotFound
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

// NotAcceptableError stands for when the request cannot be
// fulfilled because the client does not accept what the
// server can generate as output content.
type NotAcceptableError struct{}

// Error implements the error interface in NotAcceptableError.
func (e NotAcceptableError) Error() string {
	return "not acceptable"
}

// Code implements the error code for the Error interface in NotAcceptableError.
func (e NotAcceptableError) Code() ErrorCode {
	return ErrNotAcceptable
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

// SingletonAlreadyExistsError stands for when creating a singleton would
// collide with an existing active singleton in the same scope.
type SingletonAlreadyExistsError struct{}

// Error implements the error interface in SingletonAlreadyExistsError.
func (e SingletonAlreadyExistsError) Error() string {
	return "already exists"
}

// Code implements the error code for the Error interface in
// SingletonAlreadyExistsError, returning ErrConflict.
func (e SingletonAlreadyExistsError) Code() ErrorCode {
	return ErrConflict
}

// SingletonDeletedExistsError stands for when creating a singleton would
// collide with an existing deleted singleton in the same scope.
type SingletonDeletedExistsError struct{}

// Error implements the error interface in SingletonDeletedExistsError.
func (e SingletonDeletedExistsError) Error() string {
	return "restore it"
}

// Code implements the error code for the Error interface in
// SingletonDeletedExistsError, returning ErrConflict.
func (e SingletonDeletedExistsError) Code() ErrorCode {
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

// ThrottledError stands for when the request cannot be
// fulfilled since too many requests arrived from that host.
type ThrottledError struct{}

// Error implements the error interface in ThrottledError.
func (e ThrottledError) Error() string {
	return "too many requests"
}

// Code implements the error code for the Error interface in ThrottledError.
func (e ThrottledError) Code() ErrorCode {
	return ErrThrottled
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
	_ Error = BadRequestError{}
	_ Error = UnauthorizedError{}
	_ Error = ForbiddenError{}
	_ Error = NotFoundError[string]{}
	_ Error = InvalidIDError{}
	_ Error = NotDeletedError{}
	_ Error = NotAcceptableError{}
	_ Error = AlreadyUsedError{
		KeyFields: nil,
		Values:    nil,
	}
	_ Error = SingletonAlreadyExistsError{}
	_ Error = SingletonDeletedExistsError{}
	_ Error = ValidationError{
		Errors: nil,
	}
	_ Error = ThrottledError{}
	_ Error = InternalError{}
)

// RenderError renders an error to a map, to serve as a response.
func RenderError(e Error) (map[string]any, uint16) {
	rendered := map[string]any{}

	value := reflect.ValueOf(e)
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			rendered["code"] = e.Code()
			rendered["detail"] = e.Error()
			return rendered, 0
		}

		value = value.Elem()
	}

	if value.Kind() == reflect.Struct {
		valueType := value.Type()
		for i := range value.NumField() {
			fieldType := valueType.Field(i)
			if fieldType.PkgPath != "" {
				continue
			}

			name := fieldType.Name
			if tag := fieldType.Tag.Get("json"); tag != "" {
				tagName := strings.Split(tag, ",")[0]
				if tagName == "-" {
					continue
				}
				if tagName != "" {
					name = tagName
				}
			}

			rendered[name] = value.Field(i).Interface()
		}
	}

	rendered["detail"] = e.Error()
	return rendered, uint16(e.Code())
}
