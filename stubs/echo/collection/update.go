package collection

import (
	"errors"

	"github.com/labstack/echo/v4"
	"github.com/universe-10th/echo-resources/types"
)

// UpdateEndpointEngine provides methods to be used inside an endpoint stub.
type UpdateEndpointEngine[IDT comparable, RT types.Resource[IDT]] interface {
	ParsesID[IDT]
	MayScopeConstraints
	MayScopeDeletedElements
	RetrievesElementById[IDT, RT]
	PreservesStampsAndConstraints[IDT, RT]
	ReadsElementBody[IDT, RT]
	MayApplyElementConstraints[IDT, RT]
	ValidatesAndSavesElement[IDT, RT]
	RendersElement[IDT, RT]
}

// UpdateEndpointStub implements a stub that provides an echo endpoint based
// on an implementation engine that updates non-deleted elements.
type UpdateEndpointStub[IDT comparable, RT types.Resource[IDT]] struct {
	engine UpdateEndpointEngine[IDT, RT]
}

// NewUpdateEndpointStub creates a stub for updating non-deleted collection elements.
func NewUpdateEndpointStub[IDT comparable, RT types.Resource[IDT]](
	engine UpdateEndpointEngine[IDT, RT],
) UpdateEndpointStub[IDT, RT] {
	return UpdateEndpointStub[IDT, RT]{engine: engine}
}

func (updateEndpointStub UpdateEndpointStub[IDT, RT]) Update(context echo.Context) error {
	engine := updateEndpointStub.engine

	// 1-5. Retrieve the active element by id and scoped constraints.
	id, element, err := retrieveScopedElement(engine, context, false)
	if err != nil {
		return err
	}

	// 6. Preserve the stamps and constraints that must not be overwritten.
	createdAt, constraints := engine.PreserveStampsAndConstraints(context, &element)

	// 7. Read the request body.
	err = engine.ReadBody(context, &element)
	if err != nil {
		return RenderError(context, types.BadRequestError{})
	}

	// 8. Restore the preserved id, stamps, and constraints.
	engine.RestoreIDStampsAndConstraints(&element, id, createdAt, constraints)

	// 9. Apply the constraints from the path into the object.
	err = engine.ApplyPathConstraintsToElement(context, &element)
	if err != nil {
		return RenderError(context, types.BadRequestError{})
	}

	// 10. Validate the new element state.
	err = engine.Validate(&element)
	if err != nil {
		var validationError types.ValidationError
		if errors.As(err, &validationError) {
			return RenderError(context, validationError)
		}
		return RenderError(context, types.BadRequestError{})
	}

	// 11. Save the new element state.
	err = engine.Save(&element)
	if err != nil {
		var err_ types.Error
		if errors.As(err, &err_) {
			return RenderError(context, err_)
		}
		return RenderError(context, types.InternalError{})
	}

	// 12. Render the element.
	return engine.RenderElement(context, element)
}
