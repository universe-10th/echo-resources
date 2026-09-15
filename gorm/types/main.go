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

// GetCreationTime returns the time when the model was first persisted.
func (r Resource[T]) GetCreationTime() time.Time {
	return r.CreatedAt
}

// GetLastUpdateTime returns the time when the model was last updated.
func (r Resource[T]) GetLastUpdateTime() time.Time {
	return r.UpdatedAt
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

var (
	_ resourcetypes.Resource[int]                  = Resource[int]{}
	_ resourcetypes.Resource[int64]                = Resource[int64]{}
	_ resourcetypes.Resource[uint]                 = Resource[uint]{}
	_ resourcetypes.Resource[string]               = Resource[string]{}
	_ resourcetypes.Resource[time.Time]            = Resource[time.Time]{}
	_ resourcetypes.Resource[uuid.UUID]            = Resource[uuid.UUID]{}
	_ resourcetypes.SoftDeletedResource[int]       = SoftDeletedResource[int]{}
	_ resourcetypes.SoftDeletedResource[string]    = SoftDeletedResource[string]{}
	_ resourcetypes.SoftDeletedResource[uuid.UUID] = SoftDeletedResource[uuid.UUID]{}
)
