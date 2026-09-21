package presets

import (
	// "time"

	"github.com/labstack/echo/v4"
	// "github.com/universe-10th/echo-resources/stubs/echo/collection"
	// "github.com/universe-10th/echo-resources/types"
)

// ResourceURL defines the prefix and the URL argument, and provides
// a way to retrieve the underlying ID.
type ResourceURL[IDT comparable] struct {
	prefixName string
	urlArg     string
}

// PrefixName returns the name of the prefix for the current resource's URL.
func (resourceURL ResourceURL[IDT]) PrefixName() string {
	return resourceURL.prefixName
}

// URLArg returns the name of the url/path argument for this resource.
func (resourceURL ResourceURL[IDT]) URLArg() string {
	return resourceURL.urlArg
}

// ParseID parses the ID of the resource from the url arg at URLArg().
func (resourceURL ResourceURL[IDT]) ParseID(context echo.Context) (IDT, error) {
	return echo.PathParam[IDT](context, resourceURL.urlArg)
}

// NewURL creates a new ResourceURL[IDT] instance.
func NewURL[IDT comparable](prefixName string, urlArg string) ResourceURL[IDT] {
	return ResourceURL[IDT]{prefixName, urlArg}
}
