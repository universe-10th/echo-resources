package singleton

import (
	"errors"

	"github.com/labstack/echo/v4"
	"github.com/universe-10th/echo-resources/types"
)

// RestoreEndpointEngine provides methods to be used inside an endpoint stub.
type RestoreEndpointEngine[IDT comparable, RT types.SoftDeletedResource[IDT]] interface {
	MayScopeConstraints
	MayScopeDeletedElements
	RetrievesElement[IDT, RT]
	RestoresElement[IDT, RT]
	RendersElement[IDT, RT]
}

// RestoreEndpointStub implements a stub that provides an echo endpoint based on
// an implementation engine that restores the deleted singleton element.
type RestoreEndpointStub[IDT comparable, RT types.SoftDeletedResource[IDT]] struct {
	engine RestoreEndpointEngine[IDT, RT]
}

// NewRestoreEndpointStub creates a stub for restoring the deleted singleton element.
func NewRestoreEndpointStub[IDT comparable, RT types.SoftDeletedResource[IDT]](
	engine RestoreEndpointEngine[IDT, RT],
) RestoreEndpointStub[IDT, RT] {
	return RestoreEndpointStub[IDT, RT]{engine: engine}
}

func (restoreEndpointStub RestoreEndpointStub[IDT, RT]) Restore(context echo.Context) error {
	engine := restoreEndpointStub.engine

	element, err := retrieveScopedElement(engine, context, true)
	if err != nil {
		return err
	}

	restoredElement, err := engine.Restore(element.GetID())
	if err != nil {
		var err_ types.Error
		if errors.As(err, &err_) {
			return RenderError(context, err_)
		}
		return RenderError(context, types.InternalError{})
	}

	return engine.RenderElement(context, restoredElement)
}
