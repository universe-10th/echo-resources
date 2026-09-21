package types

import "time"

// The Identified interface defines the ID() method for resources.
type Identified[T comparable] interface {
	// GetID returns the id of the element. This value should never change,
	// once it's defined. The value should be scalar and the zero value of
	// T should imply the element is not saved in storage.
	GetID() T

	// SetID sets the id of the element. This method should be set only on
	// creation, if used.
	SetID(id T)

	// GetIDField tells which (JSON, not struct nor underlying storage) field
	// is used to track the id field of this resource.
	GetIDField() string
}

// The Timestamps interface defined the GetCreationTime() and GetLastUpdateTime() methods for resources.
// When created, both fields match. When updated, the value of GetLastUpdateTime() will be different.
type Timestamps interface {
	// GetCreationTime tells the time it was created. This value should never change.
	GetCreationTime() time.Time

	// GetLastUpdateTime tells the time it was last-updated.
	GetLastUpdateTime() time.Time

	// SetCreationTime sets the creation time of a resource. Implementors must use
	// the current time, expressed in UTC.
	SetCreationTime()

	// SetCreationTimeIn sets the creation time of a resource, using a specific
	// timezone (a null pointer must be coalesced to UTC).
	SetCreationTimeIn(location *time.Location)

	// SetLastUpdateTime sets the last update time of a resource. Implementors must
	// use the current time, expressed in UTC.
	SetLastUpdateTime()

	// SetLastUpdateTimeIn sets the last update time of a resource, using a specific
	// timezone (a null pointer must be coalesced to UTC).
	SetLastUpdateTimeIn(location *time.Location)

	// GetCreationTimeField tells which (JSON, not struct nor underlying storage)
	// field is used to track the date the resource was created on.
	GetCreationTimeField() string

	// GetLastUpdateTimeField tells which (JSON, not struct nor underlying storage)
	// field is used to track the date the resource was last updated on.
	GetLastUpdateTimeField() string
}

// The DeletionTimestamp interface tells whether the resource is deleted or not, and at what time.
// The constraint here is that, if GetDeletionTime() returns a non-nil value, IsDeleted() must return
// true. Otherwise, it must return false.
type DeletionTimestamp interface {
	// GetDeletionTime tells the time it was deleted. On nil, this means the element
	// is not deleted.
	GetDeletionTime() *time.Time

	// IsDeleted tells whether the element is deleted. It must be true when GetDeletionTime
	// is not nil, and false otherwise.
	IsDeleted() bool

	// SetDeletionTime sets the deletion time of a resource. Implementors must use
	// the current time, expressed in UTC. This marks the object as being deleted.
	SetDeletionTime()

	// SetDeletionTimeIn sets the deletion time of a resource, using a specific
	// timezone (a null pointer must be coalesced to UTC). This marks the object
	// as being deleted.
	SetDeletionTimeIn(location *time.Location)

	// UnsetDeletionTime unsets the deletion time of a resource. This marks the
	// object as being not deleted.
	UnsetDeletionTime()

	// GetDeletionTimeField tells which (JSON, not struct nor underlying storage)
	// field is used to track the date the resource was deleted on.
	GetDeletionTimeField() string
}

// The Resource interface defines a resource with ID and create/update stamps. On deleted, the
// object disappears and is not reachable anymore.
type Resource[T comparable] interface {
	Identified[T]
	Timestamps
}

// The SoftDeletedResource interface defines a resource with ID and create/update/delete stamps. On
// deleted, the object is marked on its deletion time to become non-nil. Ideally, this means the object
// still exists in database.
type SoftDeletedResource[T comparable] interface {
	Identified[T]
	Timestamps
	DeletionTimestamp
}
