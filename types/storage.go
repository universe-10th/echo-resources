package types

// The Storage interface defines methods to access the
// underlying stored elements.
type Storage[IDT comparable, RT Resource[IDT]] interface {
	// Mapping returns the mapping used for the resource model
	// in this storage engine. Ideally, it is created by the
	// engine itself.
	Mapping() *FieldsMapping

	// GetElement returns a single element, given a filter.
	GetElement(
		filter *FilterExpression,
	) (element RT, found bool, err error)

	// GetElements returns a page of elements, given a filter,
	// a sort criterion, and limits.
	GetElements(
		filter *FilterExpression,
		sort *SortExpression,
		skip int64,
		limit int64,
	)

	// Save creates or updates an element. If the ID is not set
	// (i.e. it is zero-value), a new element will be created.
	// If it is set, an existing element will be updated (this
	// may cause a not-found error). If the element is logically
	// deleted, this method does nothing and returns a not-found
	// error as well.
	Save(element *RT) (notFound bool, err error)

	// Delete deletes an element. IF the ID is not set or does
	// not belong to any element in database, it will return
	// a not-found error. If the element is logically deleted,
	// this method does nothing and returns a not-found error
	// as well.
	Delete(element *RT) (notFound bool, err error)

	// ValidateFilter performs a filter validation. This validation
	// relates to the mapping returned in the Mapping method. Used
	// proactively in the retrieval middlewares before executing
	// any query.
	ValidateFilter(filter *FilterExpression) (err error)

	// ValidateSort performs a sort validation. This validation
	// relates to the mapping returned in the Mapping method. Used
	// proactively in the retrieval middlewares before executing
	// any query.
	ValidateSort(sort *SortExpression) (err error)

	// AddIDFilter adds a criterion to the filter for a specific
	// id. It extends the existing $and top-level criterion, if
	// such criterion is top-level, or creates a new filter with
	// the current filter, and the new filter, added together in
	// a new top-level $and criterion.
	AddIDFilter(filter *FilterExpression, id IDT)
}

// SoftDeletedStorage is a Storage that, also, considers the possibility
// of deleting the elements logically and queries for deleted elements,
// along with the possibility of restoring or pruning deleted elements.
type SoftDeletedStorage[IDT comparable, RT SoftDeletedResource[IDT]] interface {
	Storage[IDT, RT]

	// Restore undeleted a deleted element. If the element does not
	// exist or is not logically deleted, returns a not-found error.
	Restore(element *RT) (notFound bool, err error)

	// Prune definitely removes a deleted element. If the element
	// does not exist or is not logically deleted, returns a not-found
	// error.
	Prune(element *RT) (notFound bool, err error)

	// AddDeletedFilter adds a criterion to look for deleted, or for
	// not-deleted, elements. It extends the existing $and top-level
	// criterion, if such criterion is top-level, or creates a new
	// filter with the current filter, and the new filter, added
	// together in a new top-level $and criterion.
	AddDeletedFilter(filter *FilterExpression, deleted bool)
}
