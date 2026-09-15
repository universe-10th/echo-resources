package types

import "time"

// The Identified interface defines the ID() method for resources.
type Identified[T any] interface {
	// GetID returns the id of the element. This value should never change,
	// once it's defined. The value should be scalar and the zero value of
	// T should imply the element is not saved in storage.
	GetID() T
}

// The Timestamps interface defined the GetCreationTime() and GetLastUpdateTime() methods for resources.
// When created, both fields match. When updated, the value of GetLastUpdateTime() will be different.
type Timestamps interface {
	// GetCreationTime tells the time it was created. This value should never change.
	GetCreationTime() time.Time

	// GetLastUpdateTime tells the time it was last-updated.
	GetLastUpdateTime() time.Time
}

// The DeletionTimestamp interface tells whether the resource is deleted or not, and at what time.
// The constraint here is that, if GetDeletionTime() returns a non-nil value, IsDeleted() must return
// true. Otherwise, it must return false.
type DeletionTimestamp interface {
	// GetDeletionTime tells the time it was deleted. On nil, this means the element is not deleted.
	GetDeletionTime() *time.Time

	// IsDeleted tells whether the element is deleted. It must be true when GetDeletionTime is not
	// nil, and false otherwise.
	IsDeleted() bool
}

// The Resource interface defines a resource with ID and create/update stamps. On deleted, the
// object disappears and is not reachable anymore.
type Resource[T any] interface {
	Identified[T]
	Timestamps
}

// The SoftDeletedResource interface defines a resource with ID and create/update/delete stamps. On
// deleted, the object is marked on its deletion time to become non-nil. Ideally, this means the object
// still exists in database.
type SoftDeletedResource[T any] interface {
	Identified[T]
	Timestamps
	DeletionTimestamp
}
