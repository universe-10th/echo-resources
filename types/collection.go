package types

// ListOptions describes pagination and ordering for list operations. By
// the point this structure is read, the filter and sort are not validated
// yet. The List / ListDeleted methods will first try to parse the filter
// and sort values.
type ListOptions struct {
	// Skip is treated as 0 if negative.
	Skip int64

	// Limit is not used if 0 or negative.
	Limit int64

	// Sort can be empty. Implementations should validate fields against an allowlist.
	Sort SortExpression

	// Filter can be empty. Implementations should validate fields against an allowlist.
	Filter FilterExpression
}

// The CollectionList interface supports methods to retrieve a single element or a page of elements.
type CollectionList[IDT comparable, RT Resource[IDT]] interface {
	// Get retrieves one element by its id, or found=false on non-existing element.
	// This includes the case of soft-deleted: trying to get a soft-deleted
	// element will return found=false.
	Get(id IDT, filter FilterExpression) (element RT, found bool, err error)

	// List retrieves a page of elements.
	List(options ListOptions) ([]RT, error)

	// Count retrieves the count of elements, only accounting for the filter.
	Count(options FilterExpression) (int64, error)
}

// The CollectionSoftDeletedList interface supports method to retrieve a single deleted element,
// or a page of deleted elements.
type CollectionSoftDeletedList[IDT comparable, RT SoftDeletedResource[IDT]] interface {
	// GetDeleted retrieves one DELETED element by its id, or found=false on non-existing
	// or non-deleted element.
	GetDeleted(id IDT, filter FilterExpression) (element RT, found bool, err error)

	// ListDeleted retrieves a page of DELETED elements.
	ListDeleted(options ListOptions) ([]RT, error)

	// CountDeleted retrieves the count of elements, only accounting for the filter.
	CountDeleted(options FilterExpression) (int64, error)
}

// The CollectionDelete interface supports methods to delete a single element or a set of elements.
type CollectionDelete[IDT comparable] interface {
	// Delete deletes one element by its id.
	// This includes the case of soft-deleted: trying to delete a soft-deleted
	// element will return false.
	Delete(id IDT, filter FilterExpression) (bool, error)

	// DeleteMany deletes many records by their ids, and counts how many elements
	// were deleted. This includes the case of soft-deleted: elements are marked
	// as soft-deleted, and soft-deleted elements will not be marked nor counted.
	DeleteMany(ids []IDT, filter FilterExpression) (int, error)
}

// The CollectionSoftDeletedPrune interface supports a method to prune a single deleted element or many
// deleted elements.
type CollectionSoftDeletedPrune[IDT comparable] interface {
	// Prune definitely removes one deleted element by id. It returns false if the
	// element does not exist or is not deleted. Returns true otherwise.
	Prune(id IDT, filter FilterExpression) (bool, error)

	// PruneMany definitely removes many deleted elements by their ids. It returns
	// the amount of pruned elements.
	PruneMany(ids []IDT, filter FilterExpression) (int, error)
}

// The CollectionSoftDeletedRestore interface supports a method to restore (un-delete) a single deleted
// element or many deleted elements.
type CollectionSoftDeletedRestore[IDT comparable] interface {
	// Restore restores one DELETED element by its id. It returns false if the
	// element does not exist or is not deleted. Returns true otherwise.
	Restore(id IDT, filter FilterExpression) (bool, error)

	// RestoreMany restores many DELETED elements by its id. It returns the number
	// of elements restored this way.
	RestoreMany(ids []IDT, filter FilterExpression) (int, error)
}

// The CollectionUpdate interface supports a method to update a single element. The update process must
// validate the new state of the object prior to updating it.
type CollectionUpdate[IDT comparable, RT Resource[IDT]] interface {
	// UpdateOne updates a single element by their values in the map. All the fields
	// (and their values) must exist and be valid for the intended resource type. The
	// ID is never allowed and should cause validation to fail.
	//
	// Here, the error is relevant: it may involve validation errors or index errors,
	// not just plain storage errors. This includes index error (e.g. unique constraint
	// was violated somehow).
	UpdateOne(id IDT, filter FilterExpression, values map[string]any) (bool, error)
}

// The CollectionCreate interface supports a method to create a single element. The create process must
// validate the entire new object prior to inserting it. The insertion may fail due to constraints that
// are violated in deleted objects.
type CollectionCreate[IDT comparable, RT Resource[IDT]] interface {
	// CreateOne creates a single element by their entire object. All the fields must
	// exist and be valid for the intended resource type (certain fields will be allowed
	// empty in order to have automated setup on insertion).
	//
	// Here, the error is relevant: it may involve validation errors or index errors,
	// not just plain storage errors. This includes index error (e.g. unique constraint
	// was violated somehow).
	CreateOne(element RT) (IDT, error)
}
