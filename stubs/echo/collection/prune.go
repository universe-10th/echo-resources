package collection

import (
	"errors"

	"github.com/labstack/echo/v4"
	"github.com/universe-10th/echo-resources/types"
)

// PruneEndpointEngine provides methods to be used inside an endpoint stub.
type PruneEndpointEngine[IDT comparable, RT types.SoftDeletedResource[IDT]] interface {
	ParsesID[IDT]
	MayScopeConstraints
	MayScopeDeletedElements
	RetrievesElementById[IDT, RT]
	PrunesElement[IDT]
	RendersEmpty
}

// PruneEndpointStub implements a stub that provides an echo endpoint based
// on an implementation engine that permanently removes deleted elements.
type PruneEndpointStub[IDT comparable, RT types.SoftDeletedResource[IDT]] struct {
	engine PruneEndpointEngine[IDT, RT]
}

// NewPruneEndpointStub creates a stub for permanently removing deleted collection elements.
func NewPruneEndpointStub[IDT comparable, RT types.SoftDeletedResource[IDT]](
	engine PruneEndpointEngine[IDT, RT],
) PruneEndpointStub[IDT, RT] {
	return PruneEndpointStub[IDT, RT]{engine: engine}
}

func (pruneEndpointStub PruneEndpointStub[IDT, RT]) Prune(context echo.Context) error {
	engine := pruneEndpointStub.engine

	// 1-5. Retrieve the deleted element by id and scoped constraints.
	_, element, err := retrieveScopedElement(engine, context, true)
	if err != nil {
		return err
	}

	// 6. Prune the element.
	err = engine.Prune(element.GetID())
	if err != nil {
		var err_ types.Error
		if errors.As(err, &err_) {
			return RenderError(context, err_)
		}
		return RenderError(context, types.InternalError{})
	}

	// 7. Render an empty response.
	return engine.RenderEmpty(context)
}
