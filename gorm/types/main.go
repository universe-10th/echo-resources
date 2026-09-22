// Package types provides GORM-ready resource model building blocks.
package types

import (
	"time"

	"github.com/google/uuid"
	resourcetypes "github.com/universe-10th/echo-resources/types"
	"gorm.io/gorm"
)

// PrimaryKey is the set of scalar ID types supported by the GORM resource
// helpers. Each type in this set can be represented as a single database column.
type PrimaryKey interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~string | time.Time | uuid.UUID
}

// Resource is a reusable GORM model fragment for records with an ID and
// creation/update timestamps.
type Resource[T PrimaryKey] struct {
	ID        T         `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetID returns the model primary key.
func (r Resource[T]) GetID() T {
	return r.ID
}

// SetID sets the model primary key.
func (r *Resource[T]) SetID(id T) {
	r.ID = id
}

// GetIDField returns the JSON field used for the model primary key.
func (r Resource[T]) GetIDField() string {
	return "id"
}

// GetCreationTime returns the time when the model was first persisted.
func (r Resource[T]) GetCreationTime() time.Time {
	return r.CreatedAt
}

// GetLastUpdateTime returns the time when the model was last updated.
func (r Resource[T]) GetLastUpdateTime() time.Time {
	return r.UpdatedAt
}

// SetCreationTime sets the creation time in UTC.
func (r *Resource[T]) SetCreationTime() {
	r.SetCreationTimeIn(time.UTC)
}

// SetCreationTimeIn sets the creation time in a specific location. Nil uses UTC.
func (r *Resource[T]) SetCreationTimeIn(location *time.Location) {
	if location == nil {
		location = time.UTC
	}
	r.CreatedAt = time.Now().In(location)
}

// RestoreCreationTime sets the creation time to a specific value.
func (r *Resource[T]) RestoreCreationTime(time time.Time) {
	r.CreatedAt = time
}

// SetLastUpdateTime sets the last update time in UTC.
func (r *Resource[T]) SetLastUpdateTime() {
	r.SetLastUpdateTimeIn(time.UTC)
}

// SetLastUpdateTimeIn sets the last update time in a specific location. Nil uses UTC.
func (r *Resource[T]) SetLastUpdateTimeIn(location *time.Location) {
	if location == nil {
		location = time.UTC
	}
	r.UpdatedAt = time.Now().In(location)
}

// GetCreationTimeField returns the JSON field used for the creation timestamp.
func (r Resource[T]) GetCreationTimeField() string {
	return "created_at"
}

// GetLastUpdateTimeField returns the JSON field used for the update timestamp.
func (r Resource[T]) GetLastUpdateTimeField() string {
	return "updated_at"
}

// SoftDeletedResource is a reusable GORM model fragment for records that keep a
// deletion timestamp instead of disappearing from storage.
type SoftDeletedResource[T PrimaryKey] struct {
	Resource[T]
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// GetDeletionTime returns the deletion time, or nil when the model is active.
func (r SoftDeletedResource[T]) GetDeletionTime() *time.Time {
	if !r.DeletedAt.Valid {
		return nil
	}

	deletedAt := r.DeletedAt.Time
	return &deletedAt
}

// IsDeleted reports whether the model has been soft deleted.
func (r SoftDeletedResource[T]) IsDeleted() bool {
	return r.DeletedAt.Valid
}

// SetDeletionTime sets the deletion time in UTC.
func (r *SoftDeletedResource[T]) SetDeletionTime() {
	r.SetDeletionTimeIn(time.UTC)
}

// SetDeletionTimeIn sets the deletion time in a specific location. Nil uses UTC.
func (r *SoftDeletedResource[T]) SetDeletionTimeIn(location *time.Location) {
	if location == nil {
		location = time.UTC
	}
	r.DeletedAt = gorm.DeletedAt{
		Time:  time.Now().In(location),
		Valid: true,
	}
}

// UnsetDeletionTime clears the deletion time.
func (r *SoftDeletedResource[T]) UnsetDeletionTime() {
	r.DeletedAt = gorm.DeletedAt{}
}

// GetDeletionTimeField returns the JSON field used for the deletion timestamp.
func (r SoftDeletedResource[T]) GetDeletionTimeField() string {
	return "deleted_at"
}

var (
	_ resourcetypes.Resource[int]                  = (*Resource[int])(nil)
	_ resourcetypes.Resource[int64]                = (*Resource[int64])(nil)
	_ resourcetypes.Resource[uint]                 = (*Resource[uint])(nil)
	_ resourcetypes.Resource[string]               = (*Resource[string])(nil)
	_ resourcetypes.Resource[time.Time]            = (*Resource[time.Time])(nil)
	_ resourcetypes.Resource[uuid.UUID]            = (*Resource[uuid.UUID])(nil)
	_ resourcetypes.SoftDeletedResource[int]       = (*SoftDeletedResource[int])(nil)
	_ resourcetypes.SoftDeletedResource[string]    = (*SoftDeletedResource[string])(nil)
	_ resourcetypes.SoftDeletedResource[uuid.UUID] = (*SoftDeletedResource[uuid.UUID])(nil)
)
