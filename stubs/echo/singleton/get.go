package singleton

import (
	"github.com/labstack/echo/v4"
	"github.com/universe-10th/echo-resources/types"
)

// GetEndpointEngine provides methods to be used inside an endpoint stub.
type GetEndpointEngine[IDT comparable, RT types.Resource[IDT]] interface {
	MayScopeConstraints
	MayScopeDeletedElements
	RetrievesElement[IDT, RT]
	RendersElement[IDT, RT]
}

func get[IDT comparable, RT types.Resource[IDT]](
	engine GetEndpointEngine[IDT, RT], context echo.Context, deleted bool,
) error {
	element, err := retrieveScopedElement(engine, context, deleted)
	if err != nil {
		return err
	}

	return engine.RenderElement(context, element)
}

// GetEndpointStub implements a stub that provides an echo endpoint based on an
// implementation engine.
type GetEndpointStub[IDT comparable, RT types.Resource[IDT]] struct {
	engine GetEndpointEngine[IDT, RT]
}

// NewGetEndpointStub creates a stub for retrieving the non-deleted singleton element.
func NewGetEndpointStub[IDT comparable, RT types.Resource[IDT]](
	engine GetEndpointEngine[IDT, RT],
) GetEndpointStub[IDT, RT] {
	return GetEndpointStub[IDT, RT]{engine: engine}
}

func (getEndpointStub GetEndpointStub[IDT, RT]) Get(context echo.Context) error {
	engine := getEndpointStub.engine
	return get(engine, context, false)
}

// SoftDeletedGetEndpointStub implements a stub that provides an echo endpoint
// based on an implementation engine that retrieves deleted elements.
type SoftDeletedGetEndpointStub[IDT comparable, RT types.SoftDeletedResource[IDT]] struct {
	engine GetEndpointEngine[IDT, RT]
}

// NewSoftDeletedGetEndpointStub creates a stub for retrieving the deleted singleton element.
func NewSoftDeletedGetEndpointStub[IDT comparable, RT types.SoftDeletedResource[IDT]](
	engine GetEndpointEngine[IDT, RT],
) SoftDeletedGetEndpointStub[IDT, RT] {
	return SoftDeletedGetEndpointStub[IDT, RT]{engine: engine}
}

func (softDeletedGetEndpointStub SoftDeletedGetEndpointStub[IDT, RT]) GetDeleted(context echo.Context) error {
	engine := softDeletedGetEndpointStub.engine
	return get(engine, context, true)
}
