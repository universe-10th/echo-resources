package presets

import (
	"github.com/labstack/echo/v4"
	"github.com/universe-10th/echo-resources/types"
)

// FilterApplier applies extra application-level constraints to a parsed filter
// expression.
type FilterApplier func(echo.Context, *types.FilterExpression) error

// ResourceFiltering is a wrapper component for singleton filter customization.
type ResourceFiltering[IDT comparable, RT types.Resource[IDT]] struct {
	applier FilterApplier
}

// NewResourceFiltering creates a singleton filtering component using default
// no-op custom-filter behavior.
func NewResourceFiltering[IDT comparable, RT types.Resource[IDT]]() ResourceFiltering[IDT, RT] {
	return ResourceFiltering[IDT, RT]{}
}

// UsingDefaultFiltering resets this component to its default no-op behavior.
func (resourceFiltering *ResourceFiltering[IDT, RT]) UsingDefaultFiltering() {
	resourceFiltering.applier = nil
}

// UsingCustomFilter updates the custom filter applier. Passing nil restores the
// default no-op applier.
func (resourceFiltering *ResourceFiltering[IDT, RT]) UsingCustomFilter(applier FilterApplier) {
	resourceFiltering.applier = applier
}

// ApplyCustomFilter applies configured application-specific filter constraints.
// The default behavior is a no-op.
func (resourceFiltering ResourceFiltering[IDT, RT]) ApplyCustomFilter(
	context echo.Context, filter *types.FilterExpression,
) error {
	if resourceFiltering.applier == nil {
		return nil
	}

	return resourceFiltering.applier(context, filter)
}
