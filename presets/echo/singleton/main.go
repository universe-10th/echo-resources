package singleton

import (
	"github.com/labstack/echo/v4"
	"github.com/universe-10th/echo-resources/types"
	echo2 "github.com/universe-10th/echo-resources/utils/echo"
)

// ReadOnlyResourceService stands for an element resource which is:
// - Read-Only
// - Not able to track deleted objects
// Implements: WithGet
type ReadOnlyResourceService[IDT comparable, RT types.Resource[IDT]] struct {
	elementRenderer        echo2.ElementRenderer
	deletedElementRenderer echo2.ElementRenderer
}

func (readOnlyResourceService *ReadOnlyResourceService[IDT, RT]) Get(context echo.Context) error {
	return nil
}

// RenderElement renders an element in the context of a request.
func (readOnlyResourceService *ReadOnlyResourceService[IDT, RT]) RenderElement(
	context echo.Context, code int, obj RT,
) error {
	return readOnlyResourceService.elementRenderer(context, code, obj)
}

// ResourceService stands for an element resource which is:
// - Read-Write
// - Not able to track deleted objects.
// Implements: WithGet, WithCreate, WithUpdate, WithDelete
type ResourceService[IDT comparable, RT types.Resource[IDT]] struct {
	ReadOnlyResourceService[IDT, RT]
}

func (resourceService *ResourceService[IDT, RT]) Create(context echo.Context) error {
	return nil
}

func (resourceService *ResourceService[IDT, RT]) Update(context echo.Context) error {
	return nil
}

func (resourceService *ResourceService[IDT, RT]) Delete(context echo.Context) error {
	return nil
}

// ReadOnlySoftDeletedResourceService stands for a collection resource which is:
// - Read-Write
// - Able to track deleted objects
// Implements: WithGet, WithSoftDeletedGet
type ReadOnlySoftDeletedResourceService[IDT comparable, RT types.Resource[IDT]] struct {
	ReadOnlyResourceService[IDT, RT]
}

func (readOnlySoftDeletedResourceService *ReadOnlySoftDeletedResourceService[IDT, RT]) GetDeleted(context echo.Context) error {
	return nil
}

// RenderDeletedElement renders an element in the context of a request.
func (readOnlySoftDeletedResourceService *ReadOnlySoftDeletedResourceService[IDT, RT]) RenderDeletedElement(
	context echo.Context, code int, obj RT,
) error {
	if readOnlySoftDeletedResourceService.deletedElementRenderer != nil {
		return readOnlySoftDeletedResourceService.deletedElementRenderer(context, code, obj)
	}
	return readOnlySoftDeletedResourceService.elementRenderer(context, code, obj)
}

// SoftDeletedResourceService stands for a collection resource which is:
// - Read-Write
// - Able to track deleted objects
// Implements: WithGet, WithCreate, WithUpdate, WithDelete,
//
//	WithSoftDeletedGet, WithSoftDeletedRestore,
//	WithSoftDeletedPrune
type SoftDeletedResourceService[IDT comparable, RT types.Resource[IDT]] struct {
	ReadOnlySoftDeletedResourceService[IDT, RT]
	ResourceService[IDT, RT]
}

func (SoftDeletedResourceService *SoftDeletedResourceService[IDT, RT]) Restore(context echo.Context) error {
	return nil
}

func (SoftDeletedResourceService *SoftDeletedResourceService[IDT, RT]) Prune(context echo.Context) error {
	return nil
}
