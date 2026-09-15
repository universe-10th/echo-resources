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

// GetCreationTime returns the time when the document was first persisted.
func (r Resource) GetCreationTime() time.Time {
	return r.CreatedAt
}

// GetLastUpdateTime returns the time when the document was last updated.
func (r Resource) GetLastUpdateTime() time.Time {
	return r.UpdatedAt
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

// SoftDeleteIndexModel returns a reusable index for querying soft-deleted
// documents. It also keeps this package's MongoDB dependency explicit.
func SoftDeleteIndexModel() mongo.IndexModel {
	return mongo.IndexModel{
		Keys: bson.D{{Key: "deleted_at", Value: 1}},
	}
}

var (
	_ resourcetypes.Resource[bson.ObjectID]            = Resource{}
	_ resourcetypes.SoftDeletedResource[bson.ObjectID] = SoftDeletedResource{}
)
