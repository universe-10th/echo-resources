package services

import (
	"errors"

	"github.com/universe-10th/echo-resources/types"
	"github.com/universe-10th/echo-resources/utils"
)

var (
	ErrInvalidStorage              = errors.New("invalid storage")
	defaultCollectionResourceVerbs = NewResourceVerbs(
		ResourceGet, ResourceList,
		ResourceCreate, ResourceUpdate, ResourceDelete,
	)
	defaultSingletonResourceVerbs = NewResourceVerbs(
		ResourceGet,
		ResourceCreate, ResourceUpdate, ResourceDelete,
	)
	defaultSoftDeletedCollectionResourceVerbs = NewResourceVerbs(
		ResourceGet, ResourceList, ResourceGetDeleted, ResourceListDeleted,
		ResourceCreate, ResourceUpdate, ResourceDelete, ResourceRestore, ResourcePrune,
	)
	defaultSoftDeletedSingletonResourceVerbs = NewResourceVerbs(
		ResourceGet, ResourceGetDeleted,
		ResourceCreate, ResourceUpdate, ResourceDelete, ResourceRestore, ResourcePrune,
	)
)

// ResourceService describes a service that relates to elements
// being served through a set of known endpoints.
type ResourceService[IDT comparable, RT types.Resource[IDT]] struct {
	// The prefix is the name to use for the URL chunk for the
	// resource service in particular. Something like /foo/:id
	// or /bar will have a prefix like "foo" or "bar".
	prefix string

	// The storage field keeps the internal engine used to store
	// and retrieve elements.
	storage types.Storage[IDT, RT]

	// The singleton field tells whether this resource is singleton
	// or is it a collection.
	singleton bool

	// The verbs field tells which verbs will be considered for
	// the resource.
	verbs ResourceVerbs

	// The filter field keeps a custom filter applier. By default,
	// no extra filter is applied.
	filter FilterFunc
}

// Prefix returns the prefix used to register this service.
func (service ResourceService[IDT, RT]) Prefix() string {
	return service.prefix
}

// Storage returns the underlying storage for this resource.
func (service ResourceService[IDT, RT]) Storage() types.Storage[IDT, RT] {
	return service.storage
}

// Here is where the configuration starts.

// UsingVerbs sets the verbs to enable for this resource.
func (service *ResourceService[IDT, RT]) UsingVerbs(verbs ...ResourceVerb) {
	service.verbs = ResourceVerbs(utils.NewFlags[ResourceVerb](verbs...))
}

// Verbs returns the flag of verbs to use. Children classes
// MUST override this behavior if the verbs set here are none,
// so they use the default (full) set for the resource.
func (service ResourceService[IDT, RT]) Verbs() ResourceVerbs {
	if service.verbs == ResourceVerbs(0) {
		var r RT
		if _, ok := any(r).(types.SoftDeletedResource[IDT]); ok {
			if service.singleton {
				return defaultSoftDeletedSingletonResourceVerbs
			}
			return defaultSoftDeletedCollectionResourceVerbs
		}
		if service.singleton {
			return defaultSingletonResourceVerbs
		}
		return defaultCollectionResourceVerbs
	}
	return service.verbs
}
