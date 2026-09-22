package presets

import (
	"github.com/universe-10th/echo-resources/types"
	"github.com/universe-10th/echo-resources/utils"
)

// A ResourceValidator is just a wrapper component which performs a
// validation.
type ResourceValidator[IDT comparable, RT types.Resource[IDT]] struct {
	validator func(*RT) error
}

// NewResourceValidator creates a resource validator component using the
// default go-validator based validation behavior.
func NewResourceValidator[IDT comparable, RT types.Resource[IDT]]() ResourceValidator[IDT, RT] {
	return ResourceValidator[IDT, RT]{}
}

// UsingDefaultValidator updates the validator to use. If none is set,
// a default validation will be used instead, via go-validator.
func (resourceValidator *ResourceValidator[IDT, RT]) UsingDefaultValidator() {
	resourceValidator.validator = nil
}

// UsingCustomValidator updates the validator to use. If none is set,
// a default validation will be used instead, via go-validator.
func (resourceValidator *ResourceValidator[IDT, RT]) UsingCustomValidator(validator func(*RT) error) {
	resourceValidator.validator = validator
}

// Validate executes the underlying validator. If none is set, no validation will occur.
func (resourceValidator *ResourceValidator[IDT, RT]) Validate(element *RT) error {
	if resourceValidator.validator == nil {
		return utils.Validate(element)
	}
	return resourceValidator.validator(element)
}
