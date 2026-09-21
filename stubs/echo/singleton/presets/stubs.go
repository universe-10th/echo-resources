package presets

import (
	"github.com/universe-10th/echo-resources/stubs/echo/singleton"
	"github.com/universe-10th/echo-resources/types"
)

// ReadOnlyResourceServiceEngine is the engine required by ReadOnlyResourceService.
type ReadOnlyResourceServiceEngine[IDT comparable, RT types.Resource[IDT]] interface {
	singleton.GetEndpointEngine[IDT, RT]
	PrefixName() string
}

// ReadOnlyResourceService is a singleton preset for reading the active resource.
type ReadOnlyResourceService[IDT comparable, RT types.Resource[IDT]] struct {
	singleton.GetEndpointStub[IDT, RT]
	engine ReadOnlyResourceServiceEngine[IDT, RT]
}

// PrefixName returns the route prefix name for this preset.
func (service ReadOnlyResourceService[IDT, RT]) PrefixName() string {
	return service.engine.PrefixName()
}

// NewReadOnlyResourceService creates a singleton preset for reading the active resource.
func NewReadOnlyResourceService[IDT comparable, RT types.Resource[IDT]](
	engine ReadOnlyResourceServiceEngine[IDT, RT],
) ReadOnlyResourceService[IDT, RT] {
	return ReadOnlyResourceService[IDT, RT]{
		engine:          engine,
		GetEndpointStub: singleton.NewGetEndpointStub[IDT, RT](engine),
	}
}

// ReadWriteResourceServiceEngine is the engine required by ReadWriteResourceService.
type ReadWriteResourceServiceEngine[IDT comparable, RT types.Resource[IDT]] interface {
	ReadOnlyResourceServiceEngine[IDT, RT]
	singleton.CreateEndpointEngine[IDT, RT]
	singleton.UpdateEndpointEngine[IDT, RT]
	singleton.DeleteEndpointEngine[IDT, RT]
}

// ReadWriteResourceService is a singleton preset for creating, reading,
// updating, and deleting the active resource.
type ReadWriteResourceService[IDT comparable, RT types.Resource[IDT]] struct {
	ReadOnlyResourceService[IDT, RT]
	singleton.CreateEndpointStub[IDT, RT]
	singleton.UpdateEndpointStub[IDT, RT]
	singleton.DeleteEndpointStub[IDT, RT]
}

// NewReadWriteResourceService creates a singleton preset for creating, reading,
// updating, and deleting the active resource.
func NewReadWriteResourceService[IDT comparable, RT types.Resource[IDT]](
	engine ReadWriteResourceServiceEngine[IDT, RT],
) ReadWriteResourceService[IDT, RT] {
	return ReadWriteResourceService[IDT, RT]{
		ReadOnlyResourceService: NewReadOnlyResourceService[IDT, RT](engine),
		CreateEndpointStub:      singleton.NewCreateEndpointStub[IDT, RT](engine),
		UpdateEndpointStub:      singleton.NewUpdateEndpointStub[IDT, RT](engine),
		DeleteEndpointStub:      singleton.NewDeleteEndpointStub[IDT, RT](engine),
	}
}

// ReadOnlySoftDeletedResourceServiceEngine is the engine required by
// ReadOnlySoftDeletedResourceService.
type ReadOnlySoftDeletedResourceServiceEngine[IDT comparable, RT types.SoftDeletedResource[IDT]] interface {
	ReadOnlyResourceServiceEngine[IDT, RT]
}

// ReadOnlySoftDeletedResourceService is a singleton preset for reading the
// active resource and reading the deleted resource.
type ReadOnlySoftDeletedResourceService[IDT comparable, RT types.SoftDeletedResource[IDT]] struct {
	ReadOnlyResourceService[IDT, RT]
	singleton.SoftDeletedGetEndpointStub[IDT, RT]
}

// NewReadOnlySoftDeletedResourceService creates a singleton preset for reading
// the active resource and reading the deleted resource.
func NewReadOnlySoftDeletedResourceService[IDT comparable, RT types.SoftDeletedResource[IDT]](
	engine ReadOnlySoftDeletedResourceServiceEngine[IDT, RT],
) ReadOnlySoftDeletedResourceService[IDT, RT] {
	return ReadOnlySoftDeletedResourceService[IDT, RT]{
		ReadOnlyResourceService:    NewReadOnlyResourceService[IDT, RT](engine),
		SoftDeletedGetEndpointStub: singleton.NewSoftDeletedGetEndpointStub[IDT, RT](engine),
	}
}

// ReadWriteSoftDeletedResourceServiceEngine is the engine required by
// ReadWriteSoftDeletedResourceService.
type ReadWriteSoftDeletedResourceServiceEngine[IDT comparable, RT types.SoftDeletedResource[IDT]] interface {
	ReadOnlySoftDeletedResourceServiceEngine[IDT, RT]
	singleton.CreateEndpointEngine[IDT, RT]
	singleton.UpdateEndpointEngine[IDT, RT]
	singleton.DeleteEndpointEngine[IDT, RT]
	singleton.RestoreEndpointEngine[IDT, RT]
	singleton.PruneEndpointEngine[IDT, RT]
}

// ReadWriteSoftDeletedResourceService is a singleton preset for creating,
// reading, updating, deleting, restoring, and pruning a soft-deleted resource.
type ReadWriteSoftDeletedResourceService[IDT comparable, RT types.SoftDeletedResource[IDT]] struct {
	ReadOnlySoftDeletedResourceService[IDT, RT]
	singleton.CreateEndpointStub[IDT, RT]
	singleton.UpdateEndpointStub[IDT, RT]
	singleton.DeleteEndpointStub[IDT, RT]
	singleton.RestoreEndpointStub[IDT, RT]
	singleton.PruneEndpointStub[IDT, RT]
}

// NewReadWriteSoftDeletedResourceService creates a singleton preset for
// creating, reading, updating, deleting, restoring, and pruning a soft-deleted
// resource.
func NewReadWriteSoftDeletedResourceService[IDT comparable, RT types.SoftDeletedResource[IDT]](
	engine ReadWriteSoftDeletedResourceServiceEngine[IDT, RT],
) ReadWriteSoftDeletedResourceService[IDT, RT] {
	return ReadWriteSoftDeletedResourceService[IDT, RT]{
		ReadOnlySoftDeletedResourceService: NewReadOnlySoftDeletedResourceService[IDT, RT](engine),
		CreateEndpointStub:                 singleton.NewCreateEndpointStub[IDT, RT](engine),
		UpdateEndpointStub:                 singleton.NewUpdateEndpointStub[IDT, RT](engine),
		DeleteEndpointStub:                 singleton.NewDeleteEndpointStub[IDT, RT](engine),
		RestoreEndpointStub:                singleton.NewRestoreEndpointStub[IDT, RT](engine),
		PruneEndpointStub:                  singleton.NewPruneEndpointStub[IDT, RT](engine),
	}
}
