package presets

import (
	"mime"

	"github.com/labstack/echo/v4"
	"github.com/universe-10th/echo-resources/types"
	echo2 "github.com/universe-10th/echo-resources/utils/echo"
)

// A ResourceBody is just a wrapper component which performs
// a body capture. This can be of the same type or of a new,
// intermediate, type.
type ResourceBody[IDT comparable, RT types.Resource[IDT]] struct {
	reader func(echo.Context, *RT) error
}

// UsingDefaultBody ensures the reader uses a target object
// of the exact same type, with no modification at all.
func (resourceBody *ResourceBody[IDT, RT]) UsingDefaultBody() {
	resourceBody.reader = nil
}

// UsingCustomBody ensures the reader uses a custom function
// to read the body. This function takes:
//   - The context, for other checks (e.g. data from middleware).
//   - The final element itself. This element must be populated from
//     any object whose address (&object) is passed to a binder(.) call.
//
// Ensure you use utils/echo.Capture[T](context) and then map the result
// to the appropriate RT type.
func (resourceBody *ResourceBody[IDT, RT]) UsingCustomBody(
	reader func(echo.Context, *RT) error,
) {
	resourceBody.reader = reader
}

// ReadBody attempts a read of the full body object, against
// the in-use settings.
func (resourceBody *ResourceBody[IDT, RT]) ReadBody(
	context echo.Context, element *RT,
) error {
	// 1. First, understand the request is JSON.
	contentType, _, err := mime.ParseMediaType(context.Request().Header.Get(echo.HeaderContentType))
	if err != nil || contentType != echo.MIMEApplicationJSON {
		return renderBadRequest(context)
	}

	// 2. Then, if no reader is used, bind by default.
	if resourceBody.reader == nil {
		*element, err = echo2.Capture[RT](context)
		return err
	}

	// 3. Otherwise, pass the binder to the reader so users do what they please.
	return resourceBody.reader(context, element)
}

func renderBadRequest(context echo.Context) error {
	content, code := types.RenderError(types.BadRequestError{})
	if err := context.JSON(int(code), content); err != nil {
		return err
	}

	return types.BadRequestError{}
}
