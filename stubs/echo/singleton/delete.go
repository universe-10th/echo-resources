package singleton

import (
	"errors"

	"github.com/labstack/echo/v4"
	"github.com/universe-10th/echo-resources/types"
)

// DeleteEndpointEngine provides methods to be used inside an endpoint stub.
type DeleteEndpointEngine[IDT comparable, RT types.Resource[IDT]] interface {
	MayScopeConstraints
	MayScopeDeletedElements
	RetrievesElement[IDT, RT]
	DeletesElement[IDT]
	RendersEmpty
}

// DeleteEndpointStub implements a stub that provides an echo endpoint based on
// an implementation engine that deletes the non-deleted singleton element.
type DeleteEndpointStub[IDT comparable, RT types.Resource[IDT]] struct {
	engine DeleteEndpointEngine[IDT, RT]
}

// NewDeleteEndpointStub creates a stub for deleting the non-deleted singleton element.
func NewDeleteEndpointStub[IDT comparable, RT types.Resource[IDT]](
	engine DeleteEndpointEngine[IDT, RT],
) DeleteEndpointStub[IDT, RT] {
	return DeleteEndpointStub[IDT, RT]{engine: engine}
}

func (deleteEndpointStub DeleteEndpointStub[IDT, RT]) Delete(context echo.Context) error {
	engine := deleteEndpointStub.engine

	element, err := retrieveScopedElement(engine, context, false)
	if err != nil {
		return err
	}

	err = engine.Delete(element.GetID())
	if err != nil {
		var err_ types.Error
		if errors.As(err, &err_) {
			return RenderError(context, err_)
		}
		return RenderError(context, types.InternalError{})
	}

	return engine.RenderEmpty(context)
}
