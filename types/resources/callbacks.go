package resources

import "github.com/universe-10th/echo-resources/types"

// Allowance is the type of allowance when dealing with
// which fields are allowed for a user to filter or sort
// by. Please note that this relates to a list of fields
// which are JSON names.
type Allowance uint8

const (
	All = iota
	Only
	Except
)

// FilterFunc is a callback that tells the criterion to
// modify the current filter any way they please, based
// on the current context.
//
// Any per-user filter (when listing; if provided) will
// be extended by whatever a function of this type does.
type FilterFunc func(Context, *types.FilterExpression)

// AllowingFieldsFunc is a callback telling which fields
// the current context (user) is allowed to field by or
// to sort by.
//
// The returned list contains JSON names of allowed fields.
type AllowingFieldsFunc func(Context) ([]string, Allowance)

// DefaultSortFunc is a callback telling which fields will
// be used by default when the user does not specify any
// sort at all. The returned list contains JSON names for
// the fields that will be used for sorting, perhaps with
// a prefix of "-" (a single minus sign) if the criterion
// is DESC instead of ASC.
type DefaultSortFunc func(Context) []string

// ElementRendererFunc is a function that tells how an element
// is rendered. This function is type-aware, taking the element
// to render. The user might want to project that element on
// a new type (e.g. with less data)
type ElementRendererFunc[IDT comparable, RT types.Resource[IDT]] func(Context, RT) error

// PageRendererFunc is a function that tells how an elements'
// page is renderer. The elements are rendered but also the
// numbers of the current page and amount of pages are rendered.
type PageRendererFunc[IDT comparable, RT types.Resource[IDT]] func(Context, RT, int64, int64) error

// ValidatorFunc is a function that validates an item. When
// a validation error occurs, the error result must be of
// type types.ValidationError. If there is no error, then the
// result must be nil.
type ValidatorFunc[IDT comparable, RT types.Resource[IDT]] func(Context, RT) error
