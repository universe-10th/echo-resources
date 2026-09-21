package echo

import (
	"github.com/labstack/echo/v4"
	"github.com/universe-10th/echo-resources/types"
)

// Capture reads and returns a specific struct from the input body.
func Capture[IT any](context echo.Context) (IT, error) {
	var element IT
	binder := func(v any) error { return (&echo.DefaultBinder{}).BindBody(context, v) }
	if err := binder(element); err != nil {
		return element, renderBadRequest(context)
	}
	return element, nil
}

func renderBadRequest(context echo.Context) error {
	content, code := types.RenderError(types.BadRequestError{})
	if err := context.JSON(int(code), content); err != nil {
		return err
	}

	return types.BadRequestError{}
}
