package singleton

import (
	"errors"

	"github.com/labstack/echo/v4"
	"github.com/universe-10th/echo-resources/types"
)

// UpdateEndpointEngine provides methods to be used inside an endpoint stub.
type UpdateEndpointEngine[IDT comparable, RT types.Resource[IDT]] interface {
	MayScopeConstraints
	MayScopeDeletedElements
	RetrievesElement[IDT, RT]
	ReadsElementBody[IDT, RT]
	MayApplyElementConstraints[IDT, RT]
	ValidatesAndSavesElement[IDT, RT]
	RendersElement[IDT, RT]
}

// UpdateEndpointStub implements a stub that provides an echo endpoint based on
// an implementation engine that updates the non-deleted singleton element.
type UpdateEndpointStub[IDT comparable, RT types.Resource[IDT]] struct {
	engine UpdateEndpointEngine[IDT, RT]
}

// NewUpdateEndpointStub creates a stub for updating the non-deleted singleton element.
func NewUpdateEndpointStub[IDT comparable, RT types.Resource[IDT]](
	engine UpdateEndpointEngine[IDT, RT],
) UpdateEndpointStub[IDT, RT] {
	return UpdateEndpointStub[IDT, RT]{engine: engine}
}

func (updateEndpointStub UpdateEndpointStub[IDT, RT]) Update(context echo.Context) error {
	engine := updateEndpointStub.engine

	element, err := retrieveScopedElement(engine, context, false)
	if err != nil {
		return err
	}

	id := element.GetID()
	createdAt := element.GetCreationTime()

	err = engine.ReadBody(context, &element)
	if err != nil {
		return RenderError(context, types.BadRequestError{})
	}

	element.SetID(id)
	element.RestoreCreationTime(createdAt)

	err = engine.ApplyPathConstraintsToElement(context, &element)
	if err != nil {
		return RenderError(context, types.BadRequestError{})
	}

	err = engine.Validate(&element)
	if err != nil {
		var validationError types.ValidationError
		if errors.As(err, &validationError) {
			return RenderError(context, validationError)
		}
		return RenderError(context, types.ValidationError{})
	}

	err = engine.Save(&element)
	if err != nil {
		var err_ types.Error
		if errors.As(err, &err_) {
			return RenderError(context, err_)
		}
		return RenderError(context, types.InternalError{})
	}

	return engine.RenderElement(context, element, false)
}
