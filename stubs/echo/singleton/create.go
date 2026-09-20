package singleton

import (
	"errors"

	"github.com/labstack/echo/v4"
	"github.com/universe-10th/echo-resources/types"
)

// CreateEndpointEngine provides methods to be used inside an endpoint stub.
type CreateEndpointEngine[IDT comparable, RT types.Resource[IDT]] interface {
	MayScopeConstraints
	MayScopeDeletedElements
	RetrievesElement[IDT, RT]
	ReadsElementBody[IDT, RT]
	MayApplyElementConstraints[IDT, RT]
	ValidatesAndSavesElement[IDT, RT]
	RendersElement[IDT, RT]
}

// CreateEndpointStub implements a stub that provides an echo endpoint based on
// an implementation engine that creates the singleton element.
type CreateEndpointStub[IDT comparable, RT types.Resource[IDT]] struct {
	engine CreateEndpointEngine[IDT, RT]
}

func (createEndpointStub CreateEndpointStub[IDT, RT]) Create(context echo.Context) error {
	engine := createEndpointStub.engine
	var element RT

	filter, err := buildScopedFilter(engine, context, false, false)
	if err != nil {
		return err
	}

	existing, found, err := engine.RetrieveElement(filter)
	if err != nil {
		var err_ types.Error
		if errors.As(err, &err_) {
			return RenderError(context, err_)
		}
		return RenderError(context, types.InternalError{})
	}
	if found {
		if deleted, ok := any(existing).(types.SoftDeletedResource[IDT]); ok && deleted.IsDeleted() {
			return RenderError(context, types.SingletonDeletedExistsError{})
		}
		return RenderError(context, types.SingletonAlreadyExistsError{})
	}

	err = engine.ReadBody(context, &element)
	if err != nil {
		return RenderError(context, types.BadRequestError{})
	}

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
		return RenderError(context, types.BadRequestError{})
	}

	err = engine.Save(&element)
	if err != nil {
		var err_ types.Error
		if errors.As(err, &err_) {
			return RenderError(context, err_)
		}
		return RenderError(context, types.InternalError{})
	}

	return engine.RenderElement(context, element)
}
