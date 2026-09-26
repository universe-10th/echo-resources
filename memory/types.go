// Package memory provides in-process resource model building blocks and storage.
package memory

import (
	"time"

	resourcetypes "github.com/universe-10th/rest-resources/types"
)

// PrimaryKey is the set of scalar ID types supported by the memory resource
// helpers and built-in ID generator.
type PrimaryKey interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~string
}

// Resource is a reusable memory model fragment for records with an ID and
// creation/update timestamps.
type Resource[T PrimaryKey] struct {
	ID        T         `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (r Resource[T]) GetID() T {
	return r.ID
}

func (r *Resource[T]) SetID(id T) {
	r.ID = id
}

func (r Resource[T]) GetIDField() string {
	return "id"
}

func (r Resource[T]) GetCreationTime() time.Time {
	return r.CreatedAt
}

func (r Resource[T]) GetLastUpdateTime() time.Time {
	return r.UpdatedAt
}

func (r *Resource[T]) SetCreationTime() {
	r.SetCreationTimeIn(time.UTC)
}

func (r *Resource[T]) SetCreationTimeIn(location *time.Location) {
	if location == nil {
		location = time.UTC
	}
	r.CreatedAt = time.Now().In(location)
}

func (r *Resource[T]) RestoreCreationTime(stamp time.Time) {
	r.CreatedAt = stamp
}

func (r *Resource[T]) SetLastUpdateTime() {
	r.SetLastUpdateTimeIn(time.UTC)
}

func (r *Resource[T]) SetLastUpdateTimeIn(location *time.Location) {
	if location == nil {
		location = time.UTC
	}
	r.UpdatedAt = time.Now().In(location)
}

func (r Resource[T]) GetCreationTimeField() string {
	return "created_at"
}

func (r Resource[T]) GetLastUpdateTimeField() string {
	return "updated_at"
}

// SoftDeletedResource is a reusable memory model fragment for records that keep
// a deletion timestamp instead of disappearing from storage.
type SoftDeletedResource[T PrimaryKey] struct {
	Resource[T]
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

func (r SoftDeletedResource[T]) GetDeletionTime() *time.Time {
	return r.DeletedAt
}

func (r SoftDeletedResource[T]) IsDeleted() bool {
	return r.DeletedAt != nil
}

func (r *SoftDeletedResource[T]) SetDeletionTime() {
	r.SetDeletionTimeIn(time.UTC)
}

func (r *SoftDeletedResource[T]) SetDeletionTimeIn(location *time.Location) {
	if location == nil {
		location = time.UTC
	}
	deletedAt := time.Now().In(location)
	r.DeletedAt = &deletedAt
}

func (r *SoftDeletedResource[T]) UnsetDeletionTime() {
	r.DeletedAt = nil
}

func (r SoftDeletedResource[T]) GetDeletionTimeField() string {
	return "deleted_at"
}

var (
	_ resourcetypes.Resource[int]               = (*Resource[int])(nil)
	_ resourcetypes.Resource[int64]             = (*Resource[int64])(nil)
	_ resourcetypes.Resource[uint]              = (*Resource[uint])(nil)
	_ resourcetypes.Resource[string]            = (*Resource[string])(nil)
	_ resourcetypes.SoftDeletedResource[int]    = (*SoftDeletedResource[int])(nil)
	_ resourcetypes.SoftDeletedResource[string] = (*SoftDeletedResource[string])(nil)
)
