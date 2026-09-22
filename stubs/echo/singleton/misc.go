package singleton

import (
	"errors"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/universe-10th/echo-resources/types"
)

// MayScopeDeletedElements allows forcing whether the current endpoint involves
// deleted items or non-deleted items.
type MayScopeDeletedElements interface {
	ApplyDeletedFilter(filter *types.FilterExpression, deleted bool)
}

// MayScopeConstraints allows determining the constraints for the current
// request from the URL and related to its definition.
type MayScopeConstraints interface {
	ApplyPathConstraintsToFilter(context echo.Context, filter *types.FilterExpression) error
}

// MayApplyElementConstraints allows applying request path constraints directly
// into an element before it is saved.
type MayApplyElementConstraints[IDT comparable, RT types.Resource[IDT]] interface {
	ApplyPathConstraintsToElement(context echo.Context, element *RT) error
}

// RetrievesElement allows retrieving the singleton element for a filter.
type RetrievesElement[IDT comparable, RT types.Resource[IDT]] interface {
	ApplyCustomFilter(context echo.Context, filter *types.FilterExpression) error
	RetrieveElement(filter *types.FilterExpression) (RT, bool, error)
}

// RendersElement allows rendering an element.
type RendersElement[IDT comparable, RT types.Resource[IDT]] interface {
	RenderElement(context echo.Context, element RT, created bool) error
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

// PreservesStampsAndConstraints allows preserving fields that must survive the
// request body patch and then restoring them before saving.
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
	MayScopeConstraints
	MayScopeDeletedElements
	RetrievesElement[IDT, RT]
}

func buildScopedFilter[IDT comparable, RT types.Resource[IDT]](
	engine retrievesScopedElement[IDT, RT], context echo.Context, applyDeletedFilter bool, deleted bool,
) (*types.FilterExpression, error) {
	filter := &types.FilterExpression{}

	if err := engine.ApplyCustomFilter(context, filter); err != nil {
		return nil, RenderError(context, types.BadRequestError{})
	}

	if err := engine.ApplyPathConstraintsToFilter(context, filter); err != nil {
		return nil, RenderError(context, types.BadRequestError{})
	}

	if applyDeletedFilter {
		engine.ApplyDeletedFilter(filter, deleted)
	}

	return filter, nil
}

func retrieveScopedElement[IDT comparable, RT types.Resource[IDT]](
	engine retrievesScopedElement[IDT, RT], context echo.Context, deleted bool,
) (RT, error) {
	var zeroElement RT

	filter, err := buildScopedFilter(engine, context, true, deleted)
	if err != nil {
		return zeroElement, err
	}

	element, found, err := engine.RetrieveElement(filter)
	if err != nil {
		var err_ types.Error
		if errors.As(err, &err_) {
			return zeroElement, RenderError(context, err_)
		}
		return zeroElement, RenderError(context, types.InternalError{})
	}
	if !found {
		return zeroElement, RenderError(context, types.NotFoundError[IDT]{})
	}

	return element, nil
}
