package echo

import (
	"github.com/labstack/echo/v4"
	"github.com/universe-10th/echo-resources/types"
)

// An ElementRenderer is an arbitrary function used to render a response.
// The renderer does not, by default, set dynamic headers. This response
// is tuned for single elements.
type ElementRenderer func(context echo.Context, code int, obj any) error

// A CollectionRenderer is an arbitrary function used to render a response.
// The renderer does not, by default, set dynamic headers. This response
// is toned for lists of elements.
type CollectionRenderer func(context echo.Context, code int, objs any, skip int64, total int64) error

// MakeElementRenderer creates a renderer for single elements.
func MakeElementRenderer[IDT comparable, RT types.Resource[IDT]]() ElementRenderer {
	return func(context echo.Context, code int, obj any) error {
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
}

// MakeMappedElementRenderer creates a renderer for single elements, using
// a proper mapping function (such function accepts a pointer) converting
// the resource element to a specific output type, serving as projection.
func MakeMappedElementRenderer[IDT comparable, RT types.Resource[IDT], OT any](mapper func(*RT) OT) ElementRenderer {
	return func(context echo.Context, code int, obj any) error {
		switch obj := obj.(type) {
		case RT:
			return context.JSON(code, mapper(&obj))
		case *RT:
			return context.JSON(code, mapper(obj))
		default:
			response, code := types.RenderError(types.InternalError{})
			return context.JSON(int(code), response)
		}
	}
}

// MakeCollectionRenderer creates a renderer for collections of elements.
func MakeCollectionRenderer[IDT comparable, RT types.Resource[IDT]]() CollectionRenderer {
	return func(context echo.Context, code int, objs any, skip int64, total int64) error {
		switch objs := objs.(type) {
		case []RT:
			return context.JSON(code, map[string]any{
				"skip":  skip,
				"count": len(objs),
				"total": total,
			})
		case []*RT:
			return context.JSON(code, map[string]any{
				"skip":  skip,
				"count": len(objs),
				"total": total,
			})
		default:
			response, code := types.RenderError(types.InternalError{})
			return context.JSON(int(code), response)
		}
	}
}

// MakeMappedCollectionRenderer creates a renderer for collections of elements,
// using a proper mapping function (such function accepts a pointer) converting
// each resource element to a specific output type, serving as projector.
func MakeMappedCollectionRenderer[IDT comparable, RT types.Resource[IDT], OT any](mapper func(*RT) OT) CollectionRenderer {
	return func(context echo.Context, code int, objs any, skip int64, total int64) error {
		switch objs := objs.(type) {
		case []RT:
			len_ := len(objs)
			projected := make([]OT, len_)
			for index, value := range objs {
				projected[index] = mapper(&value)
			}
			return context.JSON(code, map[string]any{
				"skip":  skip,
				"count": len_,
				"total": total,
			})
		case []*RT:
			len_ := len(objs)
			projected := make([]OT, len_)
			for index, value := range objs {
				projected[index] = mapper(value)
			}
			return context.JSON(code, map[string]any{
				"skip":  skip,
				"count": len_,
				"total": total,
			})
		default:
			response, code := types.RenderError(types.InternalError{})
			return context.JSON(int(code), response)
		}
	}
}
