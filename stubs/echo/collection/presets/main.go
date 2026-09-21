package presets

import (
	"github.com/universe-10th/echo-resources/stubs/echo/collection"
	"github.com/universe-10th/echo-resources/types"
)

// ReadOnlyResourceServiceEngine is the engine required by ReadOnlyResourceService.
type ReadOnlyResourceServiceEngine[IDT comparable, RT types.Resource[IDT]] interface {
	collection.GetEndpointEngine[IDT, RT]
	collection.ListEndpointEngine[IDT, RT]
}

// ReadOnlyResourceService is a collection preset for reading active resources.
type ReadOnlyResourceService[IDT comparable, RT types.Resource[IDT]] struct {
	collection.GetEndpointStub[IDT, RT]
	collection.ListEndpointStub[IDT, RT]
}

// NewReadOnlyResourceService creates a collection preset for reading active resources.
func NewReadOnlyResourceService[IDT comparable, RT types.Resource[IDT]](
	engine ReadOnlyResourceServiceEngine[IDT, RT],
) ReadOnlyResourceService[IDT, RT] {
	return ReadOnlyResourceService[IDT, RT]{
		GetEndpointStub:  collection.NewGetEndpointStub[IDT, RT](engine),
		ListEndpointStub: collection.NewListEndpointStub[IDT, RT](engine),
	}
}

// URLArg returns the value of the URLArg method of the internal engine this stub
// was instantiated with.
func (service ReadOnlyResourceService[IDT, RT]) URLArg() string {
	return service.GetEndpointStub.URLArg()
}

// ReadWriteResourceServiceEngine is the engine required by ReadWriteResourceService.
type ReadWriteResourceServiceEngine[IDT comparable, RT types.Resource[IDT]] interface {
	ReadOnlyResourceServiceEngine[IDT, RT]
	collection.CreateEndpointEngine[IDT, RT]
	collection.UpdateEndpointEngine[IDT, RT]
	collection.DeleteEndpointEngine[IDT, RT]
}

// ReadWriteResourceService is a collection preset for creating, reading,
// updating, and deleting active resources.
type ReadWriteResourceService[IDT comparable, RT types.Resource[IDT]] struct {
	ReadOnlyResourceService[IDT, RT]
	collection.CreateEndpointStub[IDT, RT]
	collection.UpdateEndpointStub[IDT, RT]
	collection.DeleteEndpointStub[IDT, RT]
}

// NewReadWriteResourceService creates a collection preset for creating,
// reading, updating, and deleting active resources.
func NewReadWriteResourceService[IDT comparable, RT types.Resource[IDT]](
	engine ReadWriteResourceServiceEngine[IDT, RT],
) ReadWriteResourceService[IDT, RT] {
	return ReadWriteResourceService[IDT, RT]{
		ReadOnlyResourceService: NewReadOnlyResourceService[IDT, RT](engine),
		CreateEndpointStub:      collection.NewCreateEndpointStub[IDT, RT](engine),
		UpdateEndpointStub:      collection.NewUpdateEndpointStub[IDT, RT](engine),
		DeleteEndpointStub:      collection.NewDeleteEndpointStub[IDT, RT](engine),
	}
}

// ReadOnlySoftDeletedResourceServiceEngine is the engine required by
// ReadOnlySoftDeletedResourceService.
type ReadOnlySoftDeletedResourceServiceEngine[IDT comparable, RT types.SoftDeletedResource[IDT]] interface {
	ReadOnlyResourceServiceEngine[IDT, RT]
}

// ReadOnlySoftDeletedResourceService is a collection preset for reading active
// resources and reading deleted resources.
type ReadOnlySoftDeletedResourceService[IDT comparable, RT types.SoftDeletedResource[IDT]] struct {
	ReadOnlyResourceService[IDT, RT]
	collection.SoftDeletedGetEndpointStub[IDT, RT]
	collection.SoftDeletedListEndpointStub[IDT, RT]
}

// NewReadOnlySoftDeletedResourceService creates a collection preset for reading
// active resources and reading deleted resources.
func NewReadOnlySoftDeletedResourceService[IDT comparable, RT types.SoftDeletedResource[IDT]](
	engine ReadOnlySoftDeletedResourceServiceEngine[IDT, RT],
) ReadOnlySoftDeletedResourceService[IDT, RT] {
	return ReadOnlySoftDeletedResourceService[IDT, RT]{
		ReadOnlyResourceService:     NewReadOnlyResourceService[IDT, RT](engine),
		SoftDeletedGetEndpointStub:  collection.NewSoftDeletedGetEndpointStub[IDT, RT](engine),
		SoftDeletedListEndpointStub: collection.NewSoftDeletedListEndpointStub[IDT, RT](engine),
	}
}

// ReadWriteSoftDeletedResourceServiceEngine is the engine required by
// ReadWriteSoftDeletedResourceService.
type ReadWriteSoftDeletedResourceServiceEngine[IDT comparable, RT types.SoftDeletedResource[IDT]] interface {
	ReadOnlySoftDeletedResourceServiceEngine[IDT, RT]
	collection.CreateEndpointEngine[IDT, RT]
	collection.UpdateEndpointEngine[IDT, RT]
	collection.DeleteEndpointEngine[IDT, RT]
	collection.RestoreEndpointEngine[IDT, RT]
	collection.PruneEndpointEngine[IDT, RT]
}

// ReadWriteSoftDeletedResourceService is a collection preset for creating,
// reading, updating, deleting, restoring, and pruning soft-deleted resources.
type ReadWriteSoftDeletedResourceService[IDT comparable, RT types.SoftDeletedResource[IDT]] struct {
	ReadOnlySoftDeletedResourceService[IDT, RT]
	collection.CreateEndpointStub[IDT, RT]
	collection.UpdateEndpointStub[IDT, RT]
	collection.DeleteEndpointStub[IDT, RT]
	collection.RestoreEndpointStub[IDT, RT]
	collection.PruneEndpointStub[IDT, RT]
}

// NewReadWriteSoftDeletedResourceService creates a collection preset for
// creating, reading, updating, deleting, restoring, and pruning soft-deleted
// resources.
func NewReadWriteSoftDeletedResourceService[IDT comparable, RT types.SoftDeletedResource[IDT]](
	engine ReadWriteSoftDeletedResourceServiceEngine[IDT, RT],
) ReadWriteSoftDeletedResourceService[IDT, RT] {
	return ReadWriteSoftDeletedResourceService[IDT, RT]{
		ReadOnlySoftDeletedResourceService: NewReadOnlySoftDeletedResourceService[IDT, RT](engine),
		CreateEndpointStub:                 collection.NewCreateEndpointStub[IDT, RT](engine),
		UpdateEndpointStub:                 collection.NewUpdateEndpointStub[IDT, RT](engine),
		DeleteEndpointStub:                 collection.NewDeleteEndpointStub[IDT, RT](engine),
		RestoreEndpointStub:                collection.NewRestoreEndpointStub[IDT, RT](engine),
		PruneEndpointStub:                  collection.NewPruneEndpointStub[IDT, RT](engine),
	}
}
