package echo

import (
	"github.com/labstack/echo/v4"
	"github.com/universe-10th/echo-resources/types"
)

// RenderElement renders single elements.
func RenderElement[IDT comparable, RT types.Resource[IDT]](context echo.Context, code int, obj any) error {
	switch obj := obj.(type) {
	case RT:
		return context.JSON(code, obj)
	case *RT:
		return context.JSON(code, obj)
	default:
		response, code := types.RenderError(types.InternalError{})
		return context.JSON(int(code), response)
	}
}

// RenderList renders a collections of elements.
func RenderList[IDT comparable, RT types.Resource[IDT]](context echo.Context, code int, objs any, page int64, totalPages int64) error {
	switch objs := objs.(type) {
	case []RT:
		return context.JSON(code, map[string]any{
			"page":       page,
			"count":      len(objs),
			"totalPages": totalPages,
			"elements":   objs,
		})
	case []*RT:
		return context.JSON(code, map[string]any{
			"page":       page,
			"count":      len(objs),
			"totalPages": totalPages,
			"items":      objs,
		})
	default:
		response, code := types.RenderError(types.InternalError{})
		return context.JSON(int(code), response)
	}
}
