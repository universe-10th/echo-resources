package services

import (
	"errors"
	"fmt"
	"log/slog"

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

	// The constraintDBField is used when the resource is child of
	// another resource: it's the JSON name of a field to look up,
	// as part of the current filter lookup.
	constraintDBField string

	// The verbs field tells which verbs will be considered for
	// the resource.
	verbs ResourceVerbs

	// The filter field keeps a custom filter applier. By default,
	// no extra filter is applied.
	filter FilterFunc

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
func (service *ResourceService[IDT, RT]) AllowedFields() AllowedFieldsFunc {
	return service.allowedFields
}

// UsingValidator sets what's the validator to use for RT.
func (service *ResourceService[IDT, RT]) UsingValidator(validator ValidatorFunc[IDT, RT]) *ResourceService[IDT, RT] {
	service.validator = validator
	return service
}

// Validator returns the validator being used.
func (service *ResourceService[IDT, RT]) Validator() ValidatorFunc[IDT, RT] {
	return service.validator
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

// MakeElementFilter assembles a filter from the current request.
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
	if service.constraintDBField != "" {
		lastRaw, exists := context.PopElement()
		if !exists {
			logger.Error("unbalanced context stack operation (PopElement without previous matching PushElement)")
			return nil, id, types.InternalError{}
		}

		last, ok := lastRaw.(*RT)
		if !ok {
			logger.Error("PopElement retrieved a invalid element (possible PushElement/PopElement imbalance)")
			return nil, id, types.InternalError{}
		}

		filter.Restrict(&types.FilterExpression{
			Operator:    types.FilterEQ,
			Field:       service.constraintDBField,
			Value:       (*last).GetID(),
			Expressions: nil,
		})
	}

	// Then, add the per-deleted filter.
	service.storage.AddDeletedFilter(&filter, deleted)

	// Finally, validate the filter.
	if err := service.storage.ValidateFilter(&filter); err != nil {
		return nil, id, types.BadRequestError{}
	}

	// And return.
	return &filter, id, nil
}
