package types

// The SingletonGet interface supports a method to retrieve the singleton element.
type SingletonGet[IDT comparable, RT Resource[IDT]] interface {
	// Get retrieves the singleton element, or found=false if it does not exist.
	// This includes the case of soft-deleted: trying to get a soft-deleted
	// element will return found=false.
	//
	// Getting one element is done after applying the filter.
	Get(filter FilterExpression) (element RT, found bool, err error)
}

// The SingletonDeletedGet interface supports a method to retrieve the deleted singleton element.
type SingletonDeletedGet[IDT comparable, RT SoftDeletedResource[IDT]] interface {
	// GetDeleted retrieves the DELETED singleton element, or found=false if it
	// does not exist or is not deleted.
	//
	// Getting one deleted element is done after applying the filter.
	GetDeleted(filter FilterExpression) (element RT, found bool, err error)
}

// The SingletonDelete interface supports a method to delete the singleton element.
type SingletonDelete interface {
	// Delete deletes the singleton element. It returns false if the element does
	// not exist. This includes the case of soft-deleted: trying to delete a
	// soft-deleted element will return false.
	//
	// Getting one element to delete is done after applying the filter.
	Delete(filter FilterExpression) (bool, error)
}

// The SingletonSoftDeletedPrune interface supports a method to prune the deleted singleton element.
type SingletonSoftDeletedPrune interface {
	// Prune definitely removes the deleted singleton element. It returns false if
	// the element does not exist or is not deleted. Returns true otherwise.
	//
	// Getting one deleted element to prune is done after applying the filter.
	Prune(filter FilterExpression) (bool, error)
}

// The SingletonSoftDeletedRestore interface supports a method to restore (un-delete) the singleton element.
type SingletonSoftDeletedRestore interface {
	// Restore restores the DELETED singleton element. It returns false if the
	// element does not exist or is not deleted. Returns true otherwise.
	//
	// Getting one deleted element to restore is done after applying the filter.
	Restore(filter FilterExpression) (bool, error)
}

// The SingletonUpdate interface supports a method to update the singleton element. The update process must
// validate the new state of the object prior to updating it.
type SingletonUpdate[IDT comparable, RT Resource[IDT]] interface {
	// Update updates the singleton element by their values in the map. All the
	// fields (and their values) must exist and be valid for the intended resource
	// type. The ID is never allowed and should cause validation to fail.
	//
	// Here, the error is relevant: it may involve validation errors or index
	// errors, not just plain storage errors. This includes index error
	// (e.g. unique constraint was violated somehow).
	//
	// Getting one element to update is done after applying the filter.
	Update(filter FilterExpression, values map[string]any) (bool, error)
}

// The SingletonCreate interface supports a method to create the singleton element. The create process must
// validate the entire new object prior to inserting it. The insertion may fail due to constraints that
// are violated in deleted objects.
type SingletonCreate[IDT comparable, RT Resource[IDT]] interface {
	// Create creates the singleton element by their entire object. All the fields
	// must exist and be valid for the intended resource type (certain fields will
	// be allowed empty in order to have automated setup on insertion).
	//
	// Aside from error, it returns false if an object already exists.
	//
	// Here, the error is relevant: it may involve validation errors or index
	// errors, not just plain storage errors. This includes index error
	// (e.g. unique constraint was violated somehow).
	Create(element RT) (bool, error)
}
