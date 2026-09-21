package collection

import (
	"errors"

	"github.com/labstack/echo/v4"
	"github.com/universe-10th/echo-resources/types"
)

// ListEndpointEngine provides methods to be used inside an endpoint stub.
type ListEndpointEngine[IDT comparable, RT types.Resource[IDT]] interface {
	PicksAllowedFields
	MayScopeConstraints
	MayScopeDeletedElements
	RetrievesList[IDT, RT]
	RendersList[IDT, RT]
}

type pagingSettings struct {
	Page int64 `query:"page"`
}

func list[IDT comparable, RT types.Resource[IDT]](engine ListEndpointEngine[IDT, RT], context echo.Context, deleted bool) error {
	// 1. Determine which fields are valid for the user to explore
	//    or filter by.
	fieldsList, allowanceType, err := engine.GetAllowedFields()
	if err != nil {
		return RenderError(context, types.BadRequestError{})
	}

	// 2. Parse the paging settings.
	var paging pagingSettings
	if err := context.Bind(&paging); err != nil {
		return RenderError(context, types.BadRequestError{})
	}

	// 3. Parse the filter and sort.
	filter, sort, err := engine.ParseFilterAndSort(context, fieldsList, allowanceType)
	if err != nil {
		var err_ types.Error
		if errors.As(err, &err_) {
			return RenderError(context, err_)
		}
		return RenderError(context, types.BadRequestError{})
	}

	// 4. Enhance the filter.
	err = engine.ApplyCustomFilterAndSort(context, filter, sort)
	if err != nil {
		return RenderError(context, types.BadRequestError{})
	}

	// 5. Apply the constraints.
	err = engine.ApplyPathConstraints(context, filter)
	if err != nil {
		return RenderError(context, types.BadRequestError{})
	}

	// 6. Ensuring the filter includes the deleted criterion on it.
	var r RT
	if _, ok := any(r).(types.SoftDeletedResource[IDT]); ok {
		engine.ApplyDeletedFilter(filter, deleted)
	}

	// 7. Retrieving the list of elements.
	elements, totalPages, err := engine.RetrieveList(filter, sort, paging.Page)
	if err != nil {
		return RenderError(context, types.InternalError{})
	}

	// 8. Render everything.
	return engine.RenderList(context, elements, paging.Page, totalPages)
}

// ListEndpointStub implements a stub that provides an echo endpoint based
// on an implementation engine that lists non-deleted elements.
type ListEndpointStub[IDT comparable, RT types.Resource[IDT]] struct {
	engine ListEndpointEngine[IDT, RT]
}

// NewListEndpointStub creates a stub for listing non-deleted collection elements.
func NewListEndpointStub[IDT comparable, RT types.Resource[IDT]](
	engine ListEndpointEngine[IDT, RT],
) ListEndpointStub[IDT, RT] {
	return ListEndpointStub[IDT, RT]{engine: engine}
}

func (listEndpointStub ListEndpointStub[IDT, RT]) List(context echo.Context) error {
	engine := listEndpointStub.engine
	return list(engine, context, false)
}

// SoftDeletedListEndpointStub implements a stub that provides an echo endpoint
// based on an implementation engine that lists deleted elements.
type SoftDeletedListEndpointStub[IDT comparable, RT types.SoftDeletedResource[IDT]] struct {
	engine ListEndpointEngine[IDT, RT]
}

// NewSoftDeletedListEndpointStub creates a stub for listing deleted collection elements.
func NewSoftDeletedListEndpointStub[IDT comparable, RT types.SoftDeletedResource[IDT]](
	engine ListEndpointEngine[IDT, RT],
) SoftDeletedListEndpointStub[IDT, RT] {
	return SoftDeletedListEndpointStub[IDT, RT]{engine: engine}
}

func (softDeletedListEndpointStub SoftDeletedListEndpointStub[IDT, RT]) ListDeleted(context echo.Context) error {
	engine := softDeletedListEndpointStub.engine
	return list(engine, context, true)
}
