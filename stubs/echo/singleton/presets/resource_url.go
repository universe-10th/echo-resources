package presets

// ResourceURL defines the prefix.
type ResourceURL[IDT comparable] struct {
	prefixName string
}

// PrefixName returns the name of the prefix for the current resource's URL.
func (resourceURL ResourceURL[IDT]) PrefixName() string {
	return resourceURL.prefixName
}

// NewURL creates a new ResourceURL[IDT] instance.
func NewURL[IDT comparable](prefixName string, urlArg string) ResourceURL[IDT] {
	return ResourceURL[IDT]{prefixName}
}
