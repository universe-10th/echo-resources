package singleton

import (
	"errors"

	"github.com/labstack/echo/v4"
	"github.com/universe-10th/echo-resources/types"
)

// PruneEndpointEngine provides methods to be used inside an endpoint stub.
type PruneEndpointEngine[IDT comparable, RT types.SoftDeletedResource[IDT]] interface {
	MayScopeConstraints
	MayScopeDeletedElements
	RetrievesElement[IDT, RT]
	PrunesElement[IDT]
	RendersEmpty
}

// PruneEndpointStub implements a stub that provides an echo endpoint based on
// an implementation engine that permanently removes the deleted singleton.
type PruneEndpointStub[IDT comparable, RT types.SoftDeletedResource[IDT]] struct {
	engine PruneEndpointEngine[IDT, RT]
}

// NewPruneEndpointStub creates a stub for permanently removing the deleted singleton element.
func NewPruneEndpointStub[IDT comparable, RT types.SoftDeletedResource[IDT]](
	engine PruneEndpointEngine[IDT, RT],
) PruneEndpointStub[IDT, RT] {
	return PruneEndpointStub[IDT, RT]{engine: engine}
}

func (pruneEndpointStub PruneEndpointStub[IDT, RT]) Prune(context echo.Context) error {
	engine := pruneEndpointStub.engine

	element, err := retrieveScopedElement(engine, context, true)
	if err != nil {
		return err
	}

	err = engine.Prune(element.GetID())
	if err != nil {
		var err_ types.Error
		if errors.As(err, &err_) {
			return RenderError(context, err_)
		}
		return RenderError(context, types.InternalError{})
	}

	return engine.RenderEmpty(context)
}
