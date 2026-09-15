package types

// The CollectionList interface supports methods to retrieve a single element or a page of elements.
type CollectionList[IDT any, RT Resource[IDT]] interface {
	// Get retrieves one element by its id, or nil on non-existing element.
	// This includes the case of soft-deleted: trying to get a soft-deleted
	// element will return nil.
	Get(id IDT) (*RT, error)

	// List retrieves a page of elements. The skip argument is treated as 0 if negative.
	// The limit argument is not used if 0 or negative. The order argument can be an
	// empty slice. If used, it must consist of strings like "foo" or "-bar" to tell the
	// ascending or descending order (stuff like hash or full-text will not be implemented
	// by this template by default), where "foo" and "-bar" are `camel_case` versions of
	// the declared fields.
	List(skip int, limit int, order []string) ([]RT, error)
}

// The CollectionSoftDeletedList interface supports method to retrieve a single deleted element,
// or a page of deleted elements.
type CollectionSoftDeletedList[IDT any, RT SoftDeletedResource[IDT]] interface {
	// GetDeleted retrieves one DELETED element by its id, or nil on non-existing (or
	// non-deleted) element.
	GetDeleted(id IDT) (*RT, error)

	// ListDeleted retrieves a page of DELETED elements. The skip, limit and order arguments
	// work in the same way as List.
	ListDeleted(skip int, limit int, order []string) ([]RT, error)
}

// The CollectionDelete interface supports methods to delete a single element or a set of elements.
type CollectionDelete[IDT any, RT Resource[IDT]] interface {
	// Delete deletes one element by its id, or nil on non-existing element.
	// This includes the case of soft-deleted: trying to delete a soft-deleted
	// element will return false.
	Delete(id IDT) (bool, error)

	// DeleteMany deletes many records by their ids, and counts how many elements
	// were deleted. This includes the case of soft-deleted: elements are marked
	// as soft-deleted, and soft-deleted elements will not be marked nor counted.
	DeleteMany(ids []IDT) (int, error)
}

// The CollectionSoftDeletedPrune interface supports a method to prune a single deleted element or many
// deleted elements.
type CollectionSoftDeletedPrune[IDT any, RT SoftDeletedResource[IDT]] interface {
	// Prune definitely removes one deleted element by id. It returns false if the
	// element does not exist or is not deleted. Returns true otherwise.
	Prune(id IDT) (bool, error)

	// PruneMany definitely removes many deleted elements by their ids. It returns
	// the amount of pruned elements.
	PruneMany(ids []IDT) (int, error)
}

// The CollectionSoftDeletedRestore interface supports a method to restore (un-delete) a single deleted
// element or many deleted elements.
type CollectionSoftDeletedRestore[IDT any, RT SoftDeletedResource[IDT]] interface {
	// Restore restores one DELETED element by its id. It returns false if the
	// element does not exist or is not deleted. Returns true otherwise.
	Restore(id IDT) (bool, error)

	// RestoreMany restores many DELETED elements by its id. It returns the number
	// of elements restored this way.
	RestoreMany(ids []IDT) (int, error)
}

// The CollectionUpdate interface supports a method to update a single element. The update process must
// validate the new state of the object prior to updating it.
type CollectionUpdate[IDT any, RT Resource[IDT]] interface {
	// UpdateOne updates a single element by their values in the map. All the fields
	// (and their values) must exist and be valid for the intended resource type. The
	// ID is never allowed (i.e. silently discarded) on update.
	//
	// Here, the error is relevant: it may involve validation errors or index errors,
	// not just plain storage errors. This includes index error (e.g. unique constraint
	// was violated somehow).
	UpdateOne(id IDT, values map[string]any) (bool, error)
}

// The CollectionCreate interface supports a method to create a single element. The create process must
// validate the entire new object prior to inserting it. The insertion may fail due to constraints that
// are violated in deleted objects.
type CollectionCreate[IDT any, RT Resource[IDT]] interface {
	// CreateOne creates a single element by their entire object. All the fields must
	// exist and be valid for the intended resource type (certain fields will be allowed
	// empty in order to have automated setup on insertion).
	//
	// Here, the error is relevant: it may involve validation errors or index errors,
	// not just plain storage errors. This includes index error (e.g. unique constraint
	// was violated somehow).
	CreateOne(element RT) (IDT, error)
}
