package types

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestResourceReturnsObjectIDAndTimestamps(t *testing.T) {
	t.Parallel()

	id := bson.NewObjectID()
	createdAt := time.Date(2026, 9, 15, 10, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)
	resource := Resource{
		ID:        id,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	if got := resource.GetID(); got != id {
		t.Fatalf("expected ID %s, got %s", id.Hex(), got.Hex())
	}

	if got := resource.GetCreationTime(); !got.Equal(createdAt) {
		t.Fatalf("expected creation time %s, got %s", createdAt, got)
	}

	if got := resource.GetLastUpdateTime(); !got.Equal(updatedAt) {
		t.Fatalf("expected update time %s, got %s", updatedAt, got)
	}
}

func TestSoftDeletedResourceReportsActiveDocument(t *testing.T) {
	t.Parallel()

	resource := SoftDeletedResource{
		Resource: Resource{ID: bson.NewObjectID()},
	}

	if resource.IsDeleted() {
		t.Fatal("expected active resource to not be deleted")
	}

	if got := resource.GetDeletionTime(); got != nil {
		t.Fatalf("expected nil deletion time, got %s", got)
	}
}

func TestSoftDeletedResourceReportsDeletedDocument(t *testing.T) {
	t.Parallel()

	deletedAt := time.Date(2026, 9, 15, 11, 0, 0, 0, time.UTC)
	resource := SoftDeletedResource{
		Resource:  Resource{ID: bson.NewObjectID()},
		DeletedAt: &deletedAt,
	}

	if !resource.IsDeleted() {
		t.Fatal("expected resource to be deleted")
	}

	got := resource.GetDeletionTime()
	if got == nil {
		t.Fatal("expected deletion time")
	}

	if !got.Equal(deletedAt) {
		t.Fatalf("expected deletion time %s, got %s", deletedAt, *got)
	}
}

func TestSoftDeleteIndexModel(t *testing.T) {
	t.Parallel()

	index := SoftDeleteIndexModel()
	keys, ok := index.Keys.(bson.D)
	if !ok {
		t.Fatalf("expected bson.D keys, got %T", index.Keys)
	}

	if len(keys) != 1 || keys[0].Key != "deleted_at" || keys[0].Value != 1 {
		t.Fatalf("unexpected soft delete index keys: %#v", keys)
	}
}
