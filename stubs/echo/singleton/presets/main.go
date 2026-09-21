package presets

import (
	"github.com/universe-10th/echo-resources/stubs/echo/singleton"
	"github.com/universe-10th/echo-resources/types"
)

// ReadOnlyResourceService is a singleton preset for reading the active resource.
type ReadOnlyResourceService[IDT comparable, RT types.Resource[IDT]] struct {
	singleton.GetEndpointStub[IDT, RT]
}

// ReadWriteResourceService is a singleton preset for creating, reading,
// updating, and deleting the active resource.
type ReadWriteResourceService[IDT comparable, RT types.Resource[IDT]] struct {
	ReadOnlyResourceService[IDT, RT]
	singleton.CreateEndpointStub[IDT, RT]
	singleton.UpdateEndpointStub[IDT, RT]
	singleton.DeleteEndpointStub[IDT, RT]
}

// ReadOnlySoftDeletedResourceService is a singleton preset for reading the
// active resource and reading the deleted resource.
type ReadOnlySoftDeletedResourceService[IDT comparable, RT types.SoftDeletedResource[IDT]] struct {
	ReadOnlyResourceService[IDT, RT]
	singleton.SoftDeletedGetEndpointStub[IDT, RT]
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
