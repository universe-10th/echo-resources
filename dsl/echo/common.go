package echo

import (
	"errors"

	"github.com/labstack/echo/v4"
	"github.com/universe-10th/echo-resources/types"
	"github.com/universe-10th/echo-resources/utils"
)

var (
	ErrInvalidStorage = errors.New("invalid storage")
)

// HierarchyLevel is an interface capable of defining children
// groups. Stands for a group or the Echo main object.
type HierarchyLevel interface {
	Group(prefix string, middleware ...echo.MiddlewareFunc) *echo.Group
}

// ResourceService describes a service that relates to elements
// being served through a set of known endpoints.
type ResourceService[IDT comparable, RT types.Resource[IDT]] struct {
	// The storage field keeps the internal engine used to store
	// and retrieve elements.
	storage types.Storage[IDT, RT]

	// The prefix is the name to use for the URL chunk for the
	// resource service in particular. Something like /foo/:id
	// or /bar will have a prefix like "foo" or "bar".
	prefix string
}

// Prefix returns the prefix used to register this service.
func (service ResourceService[IDT, RT]) Prefix() string {
	return service.prefix
}

// Storage returns the underlying storage for this resource.
func (service ResourceService[IDT, RT]) Storage() types.Storage[IDT, RT] {
	return service.storage
}

// CreateResourceService tries creating an instance of base service
// struct, perhaps raising an error if the storage is nil or the prefix
// is not valid.
func CreateResourceService[IDT comparable, RT types.Resource[IDT]](
	storage types.Storage[IDT, RT], prefix string,
) (ResourceService[IDT, RT], error) {
	var resource ResourceService[IDT, RT]
	if storage == nil {
		return resource, ErrInvalidStorage
	}
	if err := utils.CheckPrefix(prefix); err != nil {
		return resource, err
	}
	resource.storage = storage
	resource.prefix = prefix
	return resource, nil
}

// MustCreateResourceService tries creating an instance of base service
// struct, panicking in the same conditions CreateResourceService would
// raise an error.
func MustCreateResourceService[IDT comparable, RT types.Resource[IDT]](
	storage types.Storage[IDT, RT], prefix string,
) ResourceService[IDT, RT] {
	if resource, err := CreateResourceService[IDT, RT](storage, prefix); err != nil {
		panic(err)
	} else {
		return resource
	}
}
