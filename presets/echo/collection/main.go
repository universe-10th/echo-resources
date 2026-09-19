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
	listEngine         types.CollectionList[IDT, RT]
	collectionRenderer echo2.ListRenderer
	elementRenderer    echo2.ElementRenderer
	listRetriever      echo2.CollectionListRetrieverFunc[IDT, RT]
	elementRetriever   echo2.CollectionElementRetrieverFunc[IDT, RT]
}

// List renders all the elements being queried.
func (readOnlyResourceService *ReadOnlyResourceService[IDT, RT]) List(context echo.Context) error {
	elements, skip, total, err := readOnlyResourceService.listRetriever(context, readOnlyResourceService.listEngine)
	if err != nil {
		return err
	}

	return readOnlyResourceService.RenderCollection(context, http.StatusOK, elements, skip, total)
}

func (readOnlyResourceService *ReadOnlyResourceService[IDT, RT]) Get(context echo.Context) error {
	return nil
}

// RenderCollection renders a page of elements in the context of a request.
func (readOnlyResourceService *ReadOnlyResourceService[IDT, RT]) RenderCollection(
	context echo.Context, code int, objs []RT, skip int64, total int64,
) error {
	return readOnlyResourceService.collectionRenderer(context, code, objs, skip, total)
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
type ReadOnlySoftDeletedResourceService[IDT comparable, RT types.Resource[IDT]] struct {
	ReadOnlyResourceService[IDT, RT]
	deletedCollectionRenderer echo2.ListRenderer
	deletedElementRenderer    echo2.ElementRenderer
}

func (readOnlySoftDeletedResourceService *ReadOnlySoftDeletedResourceService[IDT, RT]) ListDeleted(context echo.Context) error {
	return nil
}

func (readOnlySoftDeletedResourceService *ReadOnlySoftDeletedResourceService[IDT, RT]) GetDeleted(context echo.Context) error {
	return nil
}

// RenderDeletedCollection renders a page of elements in the context of a request.
func (readOnlySoftDeletedResourceService *ReadOnlySoftDeletedResourceService[IDT, RT]) RenderDeletedCollection(
	context echo.Context, code int, objs []RT, skip int64, total int64,
) error {
	if readOnlySoftDeletedResourceService.deletedCollectionRenderer != nil {
		return readOnlySoftDeletedResourceService.deletedCollectionRenderer(context, code, objs, skip, total)
	}
	return readOnlySoftDeletedResourceService.collectionRenderer(context, code, objs, skip, total)
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
