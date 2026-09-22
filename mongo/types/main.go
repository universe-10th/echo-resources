// Package types provides MongoDB-ready resource model building blocks.
package types

import (
	"time"

	resourcetypes "github.com/universe-10th/echo-resources/types"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Resource is a reusable MongoDB document fragment for records with an ObjectID
// and creation/update timestamps.
type Resource struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	CreatedAt time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at" json:"updated_at"`
}

// GetID returns the document primary key.
func (r Resource) GetID() bson.ObjectID {
	return r.ID
}

// SetID sets the document primary key.
func (r *Resource) SetID(id bson.ObjectID) {
	r.ID = id
}

// GetIDField returns the JSON field used for the document primary key.
func (r Resource) GetIDField() string {
	return "id"
}

// GetCreationTime returns the time when the document was first persisted.
func (r Resource) GetCreationTime() time.Time {
	return r.CreatedAt
}

// GetLastUpdateTime returns the time when the document was last updated.
func (r Resource) GetLastUpdateTime() time.Time {
	return r.UpdatedAt
}

// SetCreationTime sets the creation time in UTC.
func (r *Resource) SetCreationTime() {
	r.SetCreationTimeIn(time.UTC)
}

// SetCreationTimeIn sets the creation time in a specific location. Nil uses UTC.
func (r *Resource) SetCreationTimeIn(location *time.Location) {
	if location == nil {
		location = time.UTC
	}
	r.CreatedAt = time.Now().In(location)
}

// RestoreCreationTime sets the creation time to a specific value.
func (r *Resource) RestoreCreationTime(time time.Time) {
	r.CreatedAt = time
}

// SetLastUpdateTime sets the last update time in UTC.
func (r *Resource) SetLastUpdateTime() {
	r.SetLastUpdateTimeIn(time.UTC)
}

// SetLastUpdateTimeIn sets the last update time in a specific location. Nil uses UTC.
func (r *Resource) SetLastUpdateTimeIn(location *time.Location) {
	if location == nil {
		location = time.UTC
	}
	r.UpdatedAt = time.Now().In(location)
}

// GetCreationTimeField returns the JSON field used for the creation timestamp.
func (r Resource) GetCreationTimeField() string {
	return "created_at"
}

// GetLastUpdateTimeField returns the JSON field used for the update timestamp.
func (r Resource) GetLastUpdateTimeField() string {
	return "updated_at"
}

// SoftDeletedResource is a reusable MongoDB document fragment for records that
// keep a deletion timestamp instead of disappearing from storage.
type SoftDeletedResource struct {
	Resource  `bson:",inline"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

// GetDeletionTime returns the deletion time, or nil when the document is active.
func (r SoftDeletedResource) GetDeletionTime() *time.Time {
	return r.DeletedAt
}

// IsDeleted reports whether the document has been soft deleted.
func (r SoftDeletedResource) IsDeleted() bool {
	return r.DeletedAt != nil
}

// SetDeletionTime sets the deletion time in UTC.
func (r *SoftDeletedResource) SetDeletionTime() {
	r.SetDeletionTimeIn(time.UTC)
}

// SetDeletionTimeIn sets the deletion time in a specific location. Nil uses UTC.
func (r *SoftDeletedResource) SetDeletionTimeIn(location *time.Location) {
	if location == nil {
		location = time.UTC
	}
	deletedAt := time.Now().In(location)
	r.DeletedAt = &deletedAt
}

// UnsetDeletionTime clears the deletion time.
func (r *SoftDeletedResource) UnsetDeletionTime() {
	r.DeletedAt = nil
}

// GetDeletionTimeField returns the JSON field used for the deletion timestamp.
func (r SoftDeletedResource) GetDeletionTimeField() string {
	return "deleted_at"
}

// SoftDeleteIndexModel returns a reusable index for querying soft-deleted
// documents. It also keeps this package's MongoDB dependency explicit.
func SoftDeleteIndexModel() mongo.IndexModel {
	return mongo.IndexModel{
		Keys: bson.D{{Key: "deleted_at", Value: 1}},
	}
}

var (
	_ resourcetypes.Resource[bson.ObjectID]            = (*Resource)(nil)
	_ resourcetypes.SoftDeletedResource[bson.ObjectID] = (*SoftDeletedResource)(nil)
)
