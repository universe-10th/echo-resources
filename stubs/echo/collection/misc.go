package collection

import (
	"errors"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/universe-10th/echo-resources/types"
)

type Allowance uint8

const (
	All = iota
	Only
	Except
)

// PicksAllowedFields allows an endpoint to tell which are
// the allowed fields.
type PicksAllowedFields interface {
	// GetAllowedFields returns a list of the involved fields
	// and tells whether the fields are the only ones allowed
	// or all the fields are allowed but the specified ones.
	// If ALL the fields are allowed, then the list of fields
	// is ignored.
	GetAllowedFields() ([]string, Allowance, error)
}

// MayScopeDeletedElements allows forcing whether the current
// endpoint involves deleted items or non-deleted items.
type MayScopeDeletedElements interface {
	// ApplyDeletedFilter applies a filter for the elements,
	// which can be to only include deleted items, or to only
	// include non-deleted items.
	ApplyDeletedFilter(filter *types.FilterExpression, deleted bool)
}

// MayScopeConstraints allows determining the constraints for
// the current request, from the URL and related to its definition.
type MayScopeConstraints interface {
	ApplyPathConstraints(context echo.Context, filter *types.FilterExpression) error
}

// MayApplyElementConstraints allows applying request path constraints directly
// into an element before it is saved.
type MayApplyElementConstraints[IDT comparable, RT types.Resource[IDT]] interface {
	ApplyPathConstraintsToElement(context echo.Context, element *RT) error
}

// ParsesID allows parsing the current element id from the URL.
type ParsesID[IDT comparable] interface {
	// URLArg tells which one is the name to use for the in-URL / in-PATH
	// argument. For example, if URLArg() returns "p_id", the echo router
	// will use .../{prefix}/:p_id as URL. This value MUST be constant and
	// MUST satisfy the regex: ^[a-zA-Z0-9]+([-_][a-zA-Z0-9]+)*$.
	URLArg() string

	// ParseID parses the ID from the URL. It MUST be implemented to use
	// the argument returned by the URLArg method.
	ParseID(context echo.Context) (IDT, error)
}

// RetrievesList allows retrieving lists of elements.
type RetrievesList[IDT comparable, RT types.Resource[IDT]] interface {
	// ParseFilterAndSort parses the filter= and sort= arguments from the
	// query string, always according to the model.
	ParseFilterAndSort(context echo.Context, fields []string, allowance Allowance) (*types.FilterExpression, *types.SortExpression, error)

	// ApplyCustomFilterAndSort enhances the parsed filter and sort, giving
	// the possibility to enhance it into more restrictions.
	ApplyCustomFilterAndSort(context echo.Context, filter *types.FilterExpression, sort *types.SortExpression) error

	// PageSize tells the size of element pages. This value will
	// be positive or a default value will be regarded.
	PageSize() int64

	// RetrieveList retrieves a list of elements. If page < 0,
	// it will be treated as 0. The second return value is the
	// total amount of available pages.
	RetrieveList(filter *types.FilterExpression, sort *types.SortExpression, page int64) ([]RT, int64, error)
}

// RendersList allows rendering lists of elements.
type RendersList[IDT comparable, RT types.Resource[IDT]] interface {
	// RenderList renders to JSON the contents of the list, along with
	// the metadata (e.g. page index, total elements).
	RenderList(context echo.Context, elements []RT, page int64, totalPages int64) error
}

// RetrievesElementById allows retrieving a single element by its id.
type RetrievesElementById[IDT comparable, RT types.Resource[IDT]] interface {
	// ApplyCustomFilter enhances an initial (empty) filter, giving the
	// possibility to enhance it into more restrictions.
	ApplyCustomFilter(context echo.Context, filter *types.FilterExpression) error

	// RetrieveElement retrieves a single element by the id and the
	// current (enhanced / applied) filter.
	RetrieveElement(id IDT, filter *types.FilterExpression) (RT, bool, error)
}

// RetrievesOnlyElement allows retrieving a single element.
type RetrievesOnlyElement[IDT comparable, RT types.Resource[IDT]] interface {
	// RetrieveElement retrieves the only single element.
	RetrieveElement(filter *types.FilterExpression) (RT, error)
}

// RendersElement allows rendering an element.
type RendersElement[IDT comparable, RT types.Resource[IDT]] interface {
	// RenderElement renders to JSON the contents of the element.
	RenderElement(context echo.Context, elements RT) error
}

// DeletesElement allows deleting an element by id.
type DeletesElement[IDT comparable] interface {
	Delete(id IDT) error
}

// RestoresElement allows restoring a soft-deleted element by id.
type RestoresElement[IDT comparable, RT types.SoftDeletedResource[IDT]] interface {
	Restore(id IDT) (RT, error)
}

// PrunesElement allows permanently removing a soft-deleted element by id.
type PrunesElement[IDT comparable] interface {
	Prune(id IDT) error
}

// RendersEmpty allows rendering an empty successful response.
type RendersEmpty interface {
	RenderEmpty(context echo.Context) error
}

// PreservesStampsAndConstraints allows preserving fields that must survive
// the request body patch and then restoring them before saving.
type PreservesStampsAndConstraints[IDT comparable, RT types.Resource[IDT]] interface {
	PreserveStampsAndConstraints(context echo.Context, element *RT) (
		createdAt time.Time, constraints any,
	)
	RestoreIDStampsAndConstraints(element *RT, id IDT, createdAt time.Time, constraints any)
}

// ReadsElementBody allows reading a request body into an element.
type ReadsElementBody[IDT comparable, RT types.Resource[IDT]] interface {
	ReadBody(context echo.Context, element *RT) error
}

// ValidatesAndSavesElement allows validating and persisting an element.
type ValidatesAndSavesElement[IDT comparable, RT types.Resource[IDT]] interface {
	Validate(element *RT) error
	Save(element *RT) error
}

// RenderError renders an error in a context, using the underlying
// types.RenderError.
func RenderError(context echo.Context, err types.Error) error {
	content, code := types.RenderError(err)
	_ = context.JSON(int(code), content)
	return err
}

type retrievesScopedElement[IDT comparable, RT types.Resource[IDT]] interface {
	ParsesID[IDT]
	MayScopeConstraints
	MayScopeDeletedElements
	RetrievesElementById[IDT, RT]
}

func retrieveScopedElement[IDT comparable, RT types.Resource[IDT]](
	engine retrievesScopedElement[IDT, RT], context echo.Context, deleted bool,
) (IDT, RT, error) {
	var zeroID IDT
	var zeroElement RT

	id, err := engine.ParseID(context)
	if err != nil {
		return zeroID, zeroElement, RenderError(context, types.BadRequestError{})
	}

	filter := &types.FilterExpression{}
	err = engine.ApplyCustomFilter(context, filter)
	if err != nil {
		return zeroID, zeroElement, RenderError(context, types.BadRequestError{})
	}

	err = engine.ApplyPathConstraints(context, filter)
	if err != nil {
		return zeroID, zeroElement, RenderError(context, types.BadRequestError{})
	}

	var r RT
	if _, ok := any(r).(types.SoftDeletedResource[IDT]); ok {
		engine.ApplyDeletedFilter(filter, deleted)
	}

	element, found, err := engine.RetrieveElement(id, filter)
	if err != nil {
		var err_ types.Error
		if errors.As(err, &err_) {
			return zeroID, zeroElement, RenderError(context, err_)
		}
		return zeroID, zeroElement, RenderError(context, types.InternalError{})
	}
	if !found {
		return zeroID, zeroElement, RenderError(context, types.NotFoundError[IDT]{Key: id})
	}

	return id, element, nil
}
