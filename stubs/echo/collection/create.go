package collection

import (
	"errors"

	"github.com/labstack/echo/v4"
	"github.com/universe-10th/echo-resources/types"
)

// CreateEndpointEngine provides methods to be used inside an endpoint stub.
type CreateEndpointEngine[IDT comparable, RT types.Resource[IDT]] interface {
	ReadsElementBody[IDT, RT]
	MayApplyElementConstraints[IDT, RT]
	ValidatesAndSavesElement[IDT, RT]
	RendersElement[IDT, RT]
}

// CreateEndpointStub implements a stub that provides an echo endpoint based
// on an implementation engine that creates elements.
type CreateEndpointStub[IDT comparable, RT types.Resource[IDT]] struct {
	engine CreateEndpointEngine[IDT, RT]
}

// NewCreateEndpointStub creates a stub for creating collection elements.
func NewCreateEndpointStub[IDT comparable, RT types.Resource[IDT]](
	engine CreateEndpointEngine[IDT, RT],
) CreateEndpointStub[IDT, RT] {
	return CreateEndpointStub[IDT, RT]{engine: engine}
}

func (createEndpointStub CreateEndpointStub[IDT, RT]) Create(context echo.Context) error {
	engine := createEndpointStub.engine
	var element RT

	// 1. Read the request body.
	err := engine.ReadBody(context, &element)
	if err != nil {
		return RenderError(context, types.BadRequestError{})
	}

	// 2. Apply the constraints from the path into the object.
	err = engine.ApplyPathConstraintsToElement(context, &element)
	if err != nil {
		return RenderError(context, types.BadRequestError{})
	}

	// 3. Set the creation time to the current timestamp.
	element.SetCreationTime()

	// 4. Validate the constrained request body.
	err = engine.Validate(&element)
	if err != nil {
		var validationError types.ValidationError
		if errors.As(err, &validationError) {
			return RenderError(context, validationError)
		}
		return RenderError(context, types.ValidationError{})
	}

	// 5. Save the element.
	err = engine.Save(&element)
	if err != nil {
		var err_ types.Error
		if errors.As(err, &err_) {
			return RenderError(context, err_)
		}
		return RenderError(context, types.InternalError{})
	}

	// 6. Render the final element.
	return engine.RenderElement(context, element, true)
}
