package echo

import (
	"mime"

	"github.com/labstack/echo/v4"
	"github.com/universe-10th/echo-resources/types"
)

// An ElementCapturer is a function used to capture an input request body.
// Typically, this capturer always expects JSON, and only for one element.
type ElementCapturer[IDT comparable, RT types.Resource[IDT]] func(context echo.Context) (RT, error)

// MakeBodyCapturer creates a capturer that requires an application/json request
// body and binds it directly into one resource element.
func MakeBodyCapturer[IDT comparable, RT types.Resource[IDT]]() ElementCapturer[IDT, RT] {
	return func(context echo.Context) (RT, error) {
		var element RT

		contentType, _, err := mime.ParseMediaType(context.Request().Header.Get(echo.HeaderContentType))
		if err != nil || contentType != echo.MIMEApplicationJSON {
			return element, renderBadRequest(context)
		}

		if err := context.Bind(&element); err != nil {
			return element, renderBadRequest(context)
		}

		return element, nil
	}
}

// MakeMappedBodyCapturer creates a capturer that requires an application/json
// request body, binds it into an input type, and maps it to one resource element.
func MakeMappedBodyCapturer[IDT comparable, RT types.Resource[IDT], IT any](mapper func(*IT) RT) ElementCapturer[IDT, RT] {
	return func(context echo.Context) (RT, error) {
		var empty RT
		var input IT

		contentType, _, err := mime.ParseMediaType(context.Request().Header.Get(echo.HeaderContentType))
		if err != nil || contentType != echo.MIMEApplicationJSON {
			return empty, renderBadRequest(context)
		}

		if err := context.Bind(&input); err != nil {
			return empty, renderBadRequest(context)
		}

		return mapper(&input), nil
	}
}

func renderBadRequest(context echo.Context) error {
	content, code := types.RenderError(types.BadRequestError{})
	if err := context.JSON(int(code), content); err != nil {
		return err
	}

	return types.BadRequestError{}
}
