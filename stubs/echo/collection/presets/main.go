package presets

import (
	"github.com/universe-10th/echo-resources/stubs/echo/collection"
	"github.com/universe-10th/echo-resources/types"
)

// ReadOnlyResourceService is a collection preset for reading active resources.
type ReadOnlyResourceService[IDT comparable, RT types.Resource[IDT]] struct {
	collection.GetEndpointStub[IDT, RT]
	collection.ListEndpointStub[IDT, RT]
}

// ReadWriteResourceService is a collection preset for creating, reading,
// updating, and deleting active resources.
type ReadWriteResourceService[IDT comparable, RT types.Resource[IDT]] struct {
	ReadOnlyResourceService[IDT, RT]
	collection.CreateEndpointStub[IDT, RT]
	collection.UpdateEndpointStub[IDT, RT]
	collection.DeleteEndpointStub[IDT, RT]
}

// ReadOnlySoftDeletedResourceService is a collection preset for reading active
// resources and reading deleted resources.
type ReadOnlySoftDeletedResourceService[IDT comparable, RT types.SoftDeletedResource[IDT]] struct {
	ReadOnlyResourceService[IDT, RT]
	collection.SoftDeletedGetEndpointStub[IDT, RT]
	collection.SoftDeletedListEndpointStub[IDT, RT]
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
