package collection

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/universe-10th/echo-resources/types"
	echo2 "github.com/universe-10th/echo-resources/utils/echo"
)

// ReadOnlyResourceService stands for a collection resource which is:
// - Read-Only
// - Not able to track deleted objects
// Implements: WithGet, WithList
type ReadOnlyResourceService[IDT comparable, RT types.Resource[IDT]] struct {
	// The underlying live-objects list engine.
	listEngine types.CollectionList[IDT, RT]

	// Retrievers for live objects.
	listRetriever    echo2.CollectionListRetrieverFunc[IDT, RT]
	elementRetriever echo2.CollectionElementRetrieverFunc[IDT, RT]

	// Renderers of retrieved live objects.
	listRenderer    echo2.ListRenderer
	elementRenderer echo2.ElementRenderer
}

// List retrieves and renders all the elements being queried.
func (readOnlyResourceService *ReadOnlyResourceService[IDT, RT]) List(context echo.Context) error {
	elements, skip, total, err := readOnlyResourceService.listRetriever(context, readOnlyResourceService.listEngine)
	if err != nil {
		return err
	}

	return readOnlyResourceService.RenderList(context, http.StatusOK, elements, skip, total)
}

// Get retrieves and renders the element being queried.
func (readOnlyResourceService *ReadOnlyResourceService[IDT, RT]) Get(context echo.Context) error {
	element, err := readOnlyResourceService.elementRetriever(context, readOnlyResourceService.listEngine)
	if err != nil {
		return err
	}

	return readOnlyResourceService.RenderElement(context, http.StatusOK, element)
}

// RenderList renders a page of elements in the context of a request.
func (readOnlyResourceService *ReadOnlyResourceService[IDT, RT]) RenderList(
	context echo.Context, code int, objs []RT, skip int64, total int64,
) error {
	return readOnlyResourceService.listRenderer(context, code, objs, skip, total)
}

// RenderElement renders a single element in the context of a request.
func (readOnlyResourceService *ReadOnlyResourceService[IDT, RT]) RenderElement(
	context echo.Context, code int, obj RT,
) error {
	return readOnlyResourceService.elementRenderer(context, code, obj)
}

// ResourceService stands for a collection resource which is:
// - Read-Write
// - Not able to track deleted objects.
// Implements: WithGet, WithList, WithCreate, WithUpdate, WithDelete
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
// Implements: WithGet, WithList, WithSoftDeletedGet, WithSoftDeletedList
type ReadOnlySoftDeletedResourceService[IDT comparable, RT types.SoftDeletedResource[IDT]] struct {
	ReadOnlyResourceService[IDT, RT]

	// The underlying dead-objects list engine.
	deletedListEngine types.CollectionSoftDeletedList[IDT, RT]

	// Retrievers for dead objects.
	deletedListRetriever    echo2.CollectionSoftDeletedListRetrieverFunc[IDT, RT]
	deletedElementRetriever echo2.CollectionSoftDeletedElementRetrieverFunc[IDT, RT]

	// Renderers of retrieved dead objects. If absent,
	// the live listRenderer / elementRenderer will be
	// used, respectively.
	deletedListRenderer    echo2.ListRenderer
	deletedElementRenderer echo2.ElementRenderer
}

// ListDeleted retrieves and renders all the deleted elements being queried.
func (readOnlySoftDeletedResourceService *ReadOnlySoftDeletedResourceService[IDT, RT]) ListDeleted(context echo.Context) error {
	elements, skip, total, err := readOnlySoftDeletedResourceService.deletedListRetriever(context, readOnlySoftDeletedResourceService.deletedListEngine)
	if err != nil {
		return err
	}

	return readOnlySoftDeletedResourceService.RenderDeletedList(context, http.StatusOK, elements, skip, total)

}

// GetDeleted retrieves and renders the element being queried.
func (readOnlySoftDeletedResourceService *ReadOnlySoftDeletedResourceService[IDT, RT]) GetDeleted(context echo.Context) error {
	element, err := readOnlySoftDeletedResourceService.deletedElementRetriever(context, readOnlySoftDeletedResourceService.deletedListEngine)
	if err != nil {
		return err
	}

	return readOnlySoftDeletedResourceService.RenderDeletedElement(context, http.StatusOK, element)
}

// RenderDeletedList renders a page of elements in the context of a request.
func (readOnlySoftDeletedResourceService *ReadOnlySoftDeletedResourceService[IDT, RT]) RenderDeletedList(
	context echo.Context, code int, objs []RT, skip int64, total int64,
) error {
	if readOnlySoftDeletedResourceService.deletedListRenderer != nil {
		return readOnlySoftDeletedResourceService.deletedListRenderer(context, code, objs, skip, total)
	}
	return readOnlySoftDeletedResourceService.listRenderer(context, code, objs, skip, total)
}

// RenderDeletedElement renders a single element in the context of a request.
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
// Implements: WithGet, WithList, WithCreate, WithUpdate, WithDelete,
//
//	WithSoftDeletedGet, WithSoftDeletedList, WithSoftDeletedRestore,
//	WithSoftDeletedPrune
type SoftDeletedResourceService[IDT comparable, RT types.SoftDeletedResource[IDT]] struct {
	ReadOnlySoftDeletedResourceService[IDT, RT]
	ResourceService[IDT, RT]
}

func (SoftDeletedResourceService *SoftDeletedResourceService[IDT, RT]) Restore(context echo.Context) error {
	return nil
}

func (SoftDeletedResourceService *SoftDeletedResourceService[IDT, RT]) Prune(context echo.Context) error {
	return nil
}
