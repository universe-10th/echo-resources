package presets

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/universe-10th/echo-resources/types"
	echo2 "github.com/universe-10th/echo-resources/utils/echo"
)

// ElementRendererFunc is a callback used to render an element.
type ElementRendererFunc func(ctx echo.Context, code int, element any) error

// ListRendererFunc is a callback used to render a list of elements
type ListRendererFunc func(ctx echo.Context, code int, elements any, page int64, totalPages int64) error

// ResourceRendering is a wrapper component that renders resource endpoint
// responses.
type ResourceRendering[IDT comparable, RT types.Resource[IDT]] struct {
	elementRenderer ElementRendererFunc
	listRenderer    ListRendererFunc
}

// NewResourceRendering creates a rendering component using the default JSON
// renderers.
func NewResourceRendering[IDT comparable, RT types.Resource[IDT]]() ResourceRendering[IDT, RT] {
	return ResourceRendering[IDT, RT]{}
}

// UsingDefaultRendering resets this component to the default JSON renderers.
func (resourceRendering *ResourceRendering[IDT, RT]) UsingDefaultRendering() {
	resourceRendering.elementRenderer = nil
	resourceRendering.listRenderer = nil
}

// UsingCustomElementRenderer updates the renderer for single-resource
// responses. Passing nil restores the default element renderer.
func (resourceRendering *ResourceRendering[IDT, RT]) UsingCustomElementRenderer(
	renderer ElementRendererFunc,
) {
	resourceRendering.elementRenderer = renderer
}

// UsingCustomListRenderer updates the renderer for resource-list responses.
// Passing nil restores the default list renderer.
func (resourceRendering *ResourceRendering[IDT, RT]) UsingCustomListRenderer(
	renderer ListRendererFunc,
) {
	resourceRendering.listRenderer = renderer
}

// RenderElement renders a single resource. Newly-created resources are rendered
// with HTTP 201 Created; all other resources are rendered with HTTP 200 OK.
func (resourceRendering *ResourceRendering[IDT, RT]) RenderElement(
	context echo.Context, element RT, created bool,
) error {
	code := http.StatusOK
	if created {
		code = http.StatusCreated
	}

	renderer := resourceRendering.elementRenderer
	if renderer == nil {
		return echo2.RenderElement[IDT, RT](context, code, element)
	}

	return renderer(context, code, element)
}

// RenderList renders a resource list with HTTP 200 OK and the supplied paging
// metadata.
func (resourceRendering *ResourceRendering[IDT, RT]) RenderList(
	context echo.Context, elements []RT, page int64, totalPages int64,
) error {
	renderer := resourceRendering.listRenderer
	if renderer == nil {
		return echo2.RenderList[IDT, RT](context, http.StatusOK, elements, page, totalPages)
	}

	return renderer(context, http.StatusOK, elements, page, totalPages)
}

// RenderEmpty renders a successful response without body using HTTP 204 No
// Content.
func (resourceRendering *ResourceRendering[IDT, RT]) RenderEmpty(context echo.Context) error {
	return context.JSON(http.StatusNoContent, nil)
}
