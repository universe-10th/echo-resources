package resources

import (
	"reflect"
	"strconv"
)

// Context is an abstract context over any supported library.
// It is, typically, a wrapper over a request context.
type Context interface {
	// Native gets the underlying native context, according
	// to what the underlying library supports.
	Native() any

	// GetPathParam gets a param from the URL path.
	GetPathParam(string) (string, error)
}

// PathParamType defines the available types for the params
// that can occur in the path.
type PathParamType interface {
	~string |
		~int8 | ~int16 | ~int32 | ~int64 | ~int |
		~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uint
}

// ParsePathParam parses a path parameter string into one of the supported path
// parameter scalar types.
func ParsePathParam[T PathParamType](v string) (T, error) {
	var zero T
	valueType := reflect.TypeOf(zero)

	switch valueType.Kind() {
	case reflect.String:
		return reflect.ValueOf(v).Convert(valueType).Interface().(T), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsed, err := strconv.ParseInt(v, 10, valueType.Bits())
		if err != nil {
			return zero, err
		}

		return reflect.ValueOf(parsed).Convert(valueType).Interface().(T), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		parsed, err := strconv.ParseUint(v, 10, valueType.Bits())
		if err != nil {
			return zero, err
		}

		return reflect.ValueOf(parsed).Convert(valueType).Interface().(T), nil
	default:
		return zero, strconv.ErrSyntax
	}
}
