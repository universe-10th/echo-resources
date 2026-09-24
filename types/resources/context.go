package resources

import (
	"reflect"
	"strconv"
)

// CookieSameSite describes the SameSite policy for a response cookie.
type CookieSameSite uint8

const (
	// CookieSameSiteDefault leaves SameSite unspecified.
	CookieSameSiteDefault CookieSameSite = iota

	// CookieSameSiteLax maps to the common Lax SameSite policy.
	CookieSameSiteLax

	// CookieSameSiteStrict maps to the common Strict SameSite policy.
	CookieSameSiteStrict

	// CookieSameSiteNone maps to the common None SameSite policy.
	CookieSameSiteNone
)

// Cookie is a portable cookie representation that can be mapped to common Go
// web frameworks without exposing one framework's native cookie type.
type Cookie struct {
	Name     string
	Value    string
	Path     string
	Domain   string
	MaxAge   int
	Secure   bool
	HTTPOnly bool
	SameSite CookieSameSite
}

// Context is an abstract context over any supported library.
// It is, typically, a wrapper over a request context.
type Context interface {
	// Native gets the underlying native context, according
	// to what the underlying library supports.
	Native() any

	// GetPathParam gets a param from the URL path.
	GetPathParam(string) (string, error)

	// GetQueryParam gets the first value of a query-string parameter.
	GetQueryParam(string) (string, error)

	// GetQueryParams gets all values of a query-string parameter.
	GetQueryParams(string) ([]string, error)

	// GetHeader gets the first value of a request header.
	GetHeader(string) (string, error)

	// GetHeaders gets all values of a request header.
	GetHeaders(string) ([]string, error)

	// GetCookie gets a request cookie by name.
	GetCookie(string) (Cookie, error)

	// BindJSON binds the JSON request body into target.
	BindJSON(target any) error

	// SetHeader sets a response header value.
	SetHeader(name string, value string)

	// SetCookie adds a response cookie.
	SetCookie(cookie Cookie)

	// RenderJSON renders body as a JSON response with status.
	RenderJSON(status int, body any) error

	// RenderNoContent renders a response status without a body.
	RenderNoContent(status int) error
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
