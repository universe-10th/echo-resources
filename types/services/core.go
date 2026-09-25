package services

import (
	"errors"
	"fmt"
	"log/slog"
	"mime"
	"reflect"

	"github.com/universe-10th/echo-resources/types"
	"github.com/universe-10th/echo-resources/utils"
)

var (
	ErrInvalidStorage              = errors.New("invalid storage")
	logger                         = slog.Default()
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
	defaultPageSize int64 = 10
)

const contentTypeApplicationJSON = "application/json"

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

	// The urlArg is only used in collection resources to tell
	// the name of the capture parameter in the path for the
	// current resource.
	urlArg string

	// The constraintJSONField is used when the resource is child of
	// another resource: it's the JSON name of a field to look up,
	// as part of the current filter lookup.
	constraintJSONField string

	// The verbs field tells which verbs will be considered for
	// the resource.
	verbs ResourceVerbs

	// The filter field keeps a custom filter applier. By default,
	// no extra filter is applied.
	filter FilterFunc

	// The reader field keeps a custom reader for the body. By
	// default, a standard JSON-read is applied. A custom func is
	// used for when the type to read is different.
	reader ReaderFunc[IDT, RT]

	// The pageSize field tells how many elements will be rendered
	// when listing elements. By default, 10 element will be used.
	pageSize int64

	// The defaultSort function tells which one is the default sort
	// criterion for the data.
	defaultSort DefaultSortFunc

	// The allowedFields function tells the function that determines
	// the per-user allowed fields. If not set, all the fields will
	// be allowed.
	allowedFields AllowedFieldsFunc

	// The elementRenderer function tells how to render an element.
	elementRenderer ElementRendererFunc[IDT, RT]

	// The pageRenderer function tells how to render a page of elements.
	pageRenderer PageRendererFunc[IDT, RT]

	// The validator function tells what's the criterion to perform
	// the validation of a body. By default, it uses go-validate.
	// Typically, this is not used unless RT has very complex
	// validation requirements.
	validator ValidatorFunc[IDT, RT]

	// The middlewares field tells the middlewares to use for all
	// the endpoints.
	middlewares []MiddlewareFunc
}

// Prefix returns the prefix used to register this service.
func (service ResourceService[IDT, RT]) Prefix() string {
	return service.prefix
}

// Storage returns the underlying storage for this resource.
func (service ResourceService[IDT, RT]) Storage() types.Storage[IDT, RT] {
	return service.storage
}

// IsSingleton tells whether the current resource is singleton
// or not (i.e. is a resource).
func (service ResourceService[IDT, RT]) IsSingleton() bool {
	return service.singleton
}

// URLArg tells the name of the argument used to capture the
// id of the current element.
func (service ResourceService[IDT, RT]) URLArg() string {
	return service.urlArg
}

// Here is where the configuration starts.

// UsingVerbs sets the verbs to enable for this resource.
func (service *ResourceService[IDT, RT]) UsingVerbs(verbs ...ResourceVerb) *ResourceService[IDT, RT] {
	service.verbs = ResourceVerbs(utils.NewFlags[ResourceVerb](verbs...))
	return service
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

// UsingFilter sets the filter to use for data retrieval.
func (service *ResourceService[IDT, RT]) UsingFilter(filter FilterFunc) *ResourceService[IDT, RT] {
	service.filter = filter
	return service
}

// Filter returns the filter to use for the data retrieval.
func (service ResourceService[IDT, RT]) Filter() FilterFunc {
	return service.filter
}

// UsingReader sets the reader to use for body retrieval.
func (service *ResourceService[IDT, RT]) UsingReader(reader ReaderFunc[IDT, RT]) *ResourceService[IDT, RT] {
	service.reader = reader
	return service
}

// Reader returns the underlying reader for body data.
func (service ResourceService[IDT, RT]) Reader() ReaderFunc[IDT, RT] {
	return service.reader
}

// The read method is private and reads an element from the body, according
// to the set reader function.
func (service ResourceService[IDT, RT]) read(context Context, obj *RT) error {
	contentTypeHeader, err := context.GetHeader("Content-Type")
	if err != nil {
		return types.BadRequestError{}
	}

	contentType, _, err := mime.ParseMediaType(contentTypeHeader)
	if err != nil || contentType != contentTypeApplicationJSON {
		return types.BadRequestError{}
	}

	if service.reader != nil {
		return service.reader(context, obj)
	}
	return context.BindJSON(obj)
}

// UsingDefaultSort sets what's the sort criterion when no
// sort is specified.
func (service *ResourceService[IDT, RT]) UsingDefaultSort(defaultSort DefaultSortFunc) *ResourceService[IDT, RT] {
	if service.singleton && defaultSort != nil {
		logger.Warn(
			"default sort is not used in singleton resources",
			"prefix", service.prefix,
		)
	} else {
		service.defaultSort = defaultSort
	}

	return service
}

// DefaultSort returns the default sort function (the function
// that tells which sort to apply when it's not specified).
func (service ResourceService[IDT, RT]) DefaultSort() DefaultSortFunc {
	return service.defaultSort
}

// UsingAllowedFields sets what's the criterion to select the
// allowed fields for a query. By default, all the valid fields
// are allowed (for filter and sort).
func (service *ResourceService[IDT, RT]) UsingAllowedFields(allowedFields AllowedFieldsFunc) *ResourceService[IDT, RT] {
	service.allowedFields = allowedFields
	return service
}

// AllowedFields returns the function that tells what are the
// per-user allowed fields.
func (service ResourceService[IDT, RT]) AllowedFields() AllowedFieldsFunc {
	return service.allowedFields
}

// UsingValidator sets what's the validator to use for RT.
func (service *ResourceService[IDT, RT]) UsingValidator(validator ValidatorFunc[IDT, RT]) *ResourceService[IDT, RT] {
	service.validator = validator
	return service
}

// Validator returns the validator being used.
func (service ResourceService[IDT, RT]) Validator() ValidatorFunc[IDT, RT] {
	return service.validator
}

// UsingElementRenderer sets what's the element renderer.
func (service *ResourceService[IDT, RT]) UsingElementRenderer(elementRenderer ElementRendererFunc[IDT, RT]) *ResourceService[IDT, RT] {
	service.elementRenderer = elementRenderer
	return service
}

// ElementRenderer returns the element renderer being used.
func (service ResourceService[IDT, RT]) ElementRenderer() ElementRendererFunc[IDT, RT] {
	return service.elementRenderer
}

// UsingPageRenderer sets what's the page renderer.
func (service *ResourceService[IDT, RT]) UsingPageRenderer(pageRenderer PageRendererFunc[IDT, RT]) *ResourceService[IDT, RT] {
	service.pageRenderer = pageRenderer
	return service
}

// PageRenderer returns the element renderer being used.
func (service ResourceService[IDT, RT]) PageRenderer() PageRendererFunc[IDT, RT] {
	return service.pageRenderer
}

// RenderElement renders a single element, perhaps using the renderer.
func (service ResourceService[IDT, RT]) RenderElement(context Context, status int, element RT) error {
	if service.elementRenderer != nil {
		return service.elementRenderer(context, element)
	}
	return context.RenderJSON(status, element)
}

// RenderPage renders a page of elements, perhaps using the renderer.
func (service ResourceService[IDT, RT]) RenderPage(context Context, status int, elements []RT, page int64, totalPages int64) error {
	if service.pageRenderer != nil {
		return service.pageRenderer(context, elements, page, totalPages)
	}
	return context.RenderJSON(status, map[string]any{
		"elements":   elements,
		"page":       page,
		"totalPages": totalPages,
	})
}

// UsingPageSize sets the amount of elements being listed per page.
func (service *ResourceService[IDT, RT]) UsingPageSize(pageSize int64) *ResourceService[IDT, RT] {
	if pageSize <= 0 {
		logger.Warn(fmt.Sprintf("invalid page size - changing to %d", defaultPageSize))
		pageSize = defaultPageSize
	}

	if service.singleton {
		logger.Warn(
			"page size is not used in singleton resources",
			"prefix", service.prefix,
		)
	}

	service.pageSize = pageSize
	return service
}

// PageSize returns the amount of elements being listed per page.
func (service ResourceService[IDT, RT]) PageSize() int64 {
	if service.pageSize == 0 {
		return defaultPageSize
	}
	return service.pageSize
}

// Here is where the utility functions for the middleware start.

// getStackedElement gets the element at the last constraint level.
func (service ResourceService[IDT, RT]) getStackedElement(context Context, index int) (RT, error) {
	var zero RT
	lastRaw, exists := context.PeekElement(index)
	if !exists {
		logger.Error("no element in context stack - provably called outside element middleware")
		return zero, types.InternalError{}
	}

	last, ok := lastRaw.(RT)
	if !ok {
		logger.Error("invalid element in context stack - provably called outside element middleware")
		return zero, types.InternalError{}
	}

	return last, nil
}

// makeElementFilter assembles a filter from the current request.
func (service ResourceService[IDT, RT]) makeElementFilter(context Context, deleted bool) (
	*types.FilterExpression, IDT, error,
) {
	// First, declare the filter.
	var filter types.FilterExpression
	var id IDT

	if !service.singleton {
		// 1. Get the ID from the path.
		urlArg := service.urlArg
		rawId, err := context.GetPathParam(urlArg)
		if err != nil {
			return nil, id, types.NotFoundError[string]{}
		}

		// 2. Second, parse it to a valid value.
		id, err := ParsePathParam[IDT](rawId)
		if err != nil {
			return nil, id, types.InvalidIDError{
				ElementName: service.prefix,
				Key:         rawId,
			}
		}

		// 3. Add the ID filter.
		service.storage.AddIDFilter(&filter, id)
	}

	// Then, add the per-context filter.
	if service.filter != nil {
		service.filter(context, &filter)
	}

	// Then, add a constraint, if any.
	if service.constraintJSONField != "" {
		// We use index 0 since the idea is to get the constraint
		// based on the current (last) element.
		last, err := service.getStackedElement(context, 0)
		if err != nil {
			return nil, id, err
		}

		filter.Restrict(&types.FilterExpression{
			Operator:    types.FilterEQ,
			Field:       service.constraintJSONField,
			Value:       last.GetID(),
			Expressions: nil,
		})
	}

	// Then, add the per-deleted filter.
	service.storage.AddDeletedFilter(&filter, deleted)

	// Finally, validate the filter.
	if err := service.storage.ValidateFilter(&filter); err != nil {
		logger.Error("invalid filter (should be fixed, since the user is not choosing this one)")
		return nil, id, types.InternalError{}
	}

	// And return.
	return &filter, id, nil
}

// applyPreviousConstraint applies the current constraint to the element, so it's
// always consistent.
func (service ResourceService[IDT, RT]) applyPreviousConstraint(context Context, element *RT) error {
	if service.constraintJSONField != "" {
		// We use index 1 since we want to get not the current
		// element but the PREVIOUS one instead.

		last, err := service.getStackedElement(context, 1)
		if err != nil {
			return err
		}

		fieldName := types.FieldForJSON(service.storage.Mapping(), service.constraintJSONField)
		if fieldName == "" {
			logger.Error(
				"constraint field is not mapped in resource",
				"prefix", service.prefix,
				"field", service.constraintJSONField,
			)
			return types.InternalError{}
		}

		if err := setElementField(element, fieldName, last.GetID()); err != nil {
			logger.Error(
				"could not apply constraint to resource element",
				"prefix", service.prefix,
				"field", service.constraintJSONField,
				"struct_field", fieldName,
				"error", err,
			)
			return types.InternalError{}
		}
	}

	return nil
}

func setElementField(element any, fieldName string, value any) error {
	if element == nil {
		return errors.New("element is nil")
	}

	elementValue := reflect.ValueOf(element)
	if elementValue.Kind() != reflect.Pointer || elementValue.IsNil() {
		return errors.New("element must be a non-nil pointer")
	}

	valueValue := reflect.ValueOf(value)
	if !valueValue.IsValid() {
		return errors.New("constraint value is invalid")
	}

	for elementValue.Kind() == reflect.Pointer {
		if elementValue.IsNil() {
			return errors.New("element contains nil pointer")
		}
		elementValue = elementValue.Elem()
	}

	if elementValue.Kind() != reflect.Struct {
		return errors.New("element must point to a struct")
	}

	fieldValue := elementValue.FieldByName(fieldName)
	if !fieldValue.IsValid() {
		return fmt.Errorf("field %q not found", fieldName)
	}
	if !fieldValue.CanSet() {
		return fmt.Errorf("field %q cannot be set", fieldName)
	}

	if valueValue.Type().AssignableTo(fieldValue.Type()) {
		fieldValue.Set(valueValue)
		return nil
	}
	if valueValue.Kind() == fieldValue.Kind() && valueValue.Type().ConvertibleTo(fieldValue.Type()) {
		fieldValue.Set(valueValue.Convert(fieldValue.Type()))
		return nil
	}

	return fmt.Errorf("value of type %s cannot be assigned to field %q of type %s", valueValue.Type(), fieldName, fieldValue.Type())
}

func (service ResourceService[IDT, RT]) validate(context Context, element RT) error {
	if service.validator != nil {
		return service.validator(context, element)
	}
	return utils.Validate(element)
}

func (service ResourceService[IDT, RT]) renderNotFound(context Context, id IDT) error {
	if service.singleton {
		return renderError(context, types.SingletonNotFoundError{
			ElementName: service.prefix,
		})
	}

	return renderError(context, types.NotFoundError[IDT]{
		ElementName: service.prefix,
		Key:         id,
	})
}

func (service ResourceService[IDT, RT]) renderStorageError(context Context, err error) error {
	return renderErrorOr(context, err, types.InternalError{})
}

func (service ResourceService[IDT, RT]) ensureSingletonCreateAllowed(context Context) (bool, error) {
	if !service.singleton {
		return true, nil
	}

	activeFilter, _, err := service.makeElementFilter(context, false)
	if err != nil {
		return false, renderErrorOr(context, err, types.BadRequestError{})
	}

	if _, found, err := service.storage.GetElement(activeFilter); err != nil {
		return false, service.renderStorageError(context, err)
	} else if found {
		return false, renderError(context, types.SingletonAlreadyExistsError{})
	}

	var zero RT
	if _, ok := any(zero).(types.SoftDeletedResource[IDT]); !ok {
		return true, nil
	}

	deletedFilter, _, err := service.makeElementFilter(context, true)
	if err != nil {
		return false, renderErrorOr(context, err, types.BadRequestError{})
	}

	if _, found, err := service.storage.GetElement(deletedFilter); err != nil {
		return false, service.renderStorageError(context, err)
	} else if found {
		return false, renderError(context, types.SingletonDeletedExistsError{})
	}

	return true, nil
}

// The get function is an endpoint to get a single element.
// Pre-requisites:
// - elementMiddleware(false) middleware for GET.
// - elementMiddleware(true) middleware for GET DELETED.
func (service ResourceService[IDT, RT]) get(context Context) error {
	// 1. Get the element.
	element, err := service.getStackedElement(context, 0)
	if err != nil {
		return err
	}

	// 2. Render it.
	return service.RenderElement(context, 200, element)
}

// The update function is an endpoint to update an existing non-deleted element.
// Pre-requisites: elementMiddleware(false) middleware.
func (service ResourceService[IDT, RT]) update(context Context) error {
	element, err := service.getStackedElement(context, 0)
	if err != nil {
		return err
	}

	id := element.GetID()
	createdAt := element.GetCreationTime()

	if err := service.read(context, &element); err != nil {
		return renderErrorOr(context, err, types.BadRequestError{})
	}

	element.SetID(id)
	element.RestoreCreationTime(createdAt)

	if err := service.applyPreviousConstraint(context, &element); err != nil {
		return renderErrorOr(context, err, types.InternalError{})
	}

	if err := service.validate(context, element); err != nil {
		return renderErrorOr(context, err, types.ValidationError{})
	}

	if notFound, err := service.storage.Save(&element); err != nil {
		return service.renderStorageError(context, err)
	} else if notFound {
		return service.renderNotFound(context, id)
	}

	return service.RenderElement(context, 200, element)
}

// The create function is an endpoint to create a new element.
func (service ResourceService[IDT, RT]) create(context Context) error {
	allowed, err := service.ensureSingletonCreateAllowed(context)
	if err != nil || !allowed {
		return err
	}

	var element RT
	if err := service.read(context, &element); err != nil {
		return renderErrorOr(context, err, types.BadRequestError{})
	}

	var zero IDT
	element.SetID(zero)
	element.SetCreationTime()

	if err := service.applyPreviousConstraint(context, &element); err != nil {
		return renderErrorOr(context, err, types.InternalError{})
	}

	if err := service.validate(context, element); err != nil {
		return renderErrorOr(context, err, types.ValidationError{})
	}

	if notFound, err := service.storage.Save(&element); err != nil {
		return service.renderStorageError(context, err)
	} else if notFound {
		return renderError(context, types.InternalError{})
	}

	return service.RenderElement(context, 201, element)
}

// The delete function is an endpoint to delete an existing non-deleted element.
// Pre-requisites: elementMiddleware(false) middleware.
func (service ResourceService[IDT, RT]) delete(context Context) error {
	element, err := service.getStackedElement(context, 0)
	if err != nil {
		return err
	}

	if notFound, err := service.storage.Delete(&element); err != nil {
		return service.renderStorageError(context, err)
	} else if notFound {
		return service.renderNotFound(context, element.GetID())
	}

	return context.RenderNoContent(204)
}

// The prune function is an endpoint to permanently delete an existing deleted
// element. Pre-requisites: elementMiddleware(true) middleware.
func (service ResourceService[IDT, RT]) prune(context Context) error {
	element, err := service.getStackedElement(context, 0)
	if err != nil {
		return err
	}

	if notFound, err := service.storage.Prune(&element); err != nil {
		return service.renderStorageError(context, err)
	} else if notFound {
		return service.renderNotFound(context, element.GetID())
	}

	return context.RenderNoContent(204)
}

// The restore function is an endpoint to restore an existing deleted element.
// Pre-requisites: elementMiddleware(true) middleware.
func (service ResourceService[IDT, RT]) restore(context Context) error {
	element, err := service.getStackedElement(context, 0)
	if err != nil {
		return err
	}

	if notFound, err := service.storage.Restore(&element); err != nil {
		return service.renderStorageError(context, err)
	} else if notFound {
		return service.renderNotFound(context, element.GetID())
	}

	return service.RenderElement(context, 200, element)
}
