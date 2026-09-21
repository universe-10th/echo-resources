package presets

import "github.com/universe-10th/echo-resources/types"

// A ResourceValidator is just a wrapper component which performs a
// validation
type ResourceValidator[IDT comparable, RT types.Resource[IDT]] struct {
	validator func(*RT) error
}

// UsingValidator updates the validator to use.
func (resourceValidator *ResourceValidator[IDT, RT]) UsingValidator(validator func(*RT) error) {
	resourceValidator.validator = validator
}

// Validate executes the underlying validator. If none is set, no validation will occur.
func (resourceValidator *ResourceValidator[IDT, RT]) Validate(element *RT) error {
	if resourceValidator.validator == nil {
		return nil
	}
	return resourceValidator.validator(element)
}
