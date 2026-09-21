package types

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestResourceReturnsPrimaryKeyAndTimestamps(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	createdAt := time.Date(2026, 9, 15, 10, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)
	resource := Resource[uuid.UUID]{
		ID:        id,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	if got := resource.GetID(); got != id {
		t.Fatalf("expected ID %s, got %s", id, got)
	}

	if got := resource.GetCreationTime(); !got.Equal(createdAt) {
		t.Fatalf("expected creation time %s, got %s", createdAt, got)
	}

	if got := resource.GetLastUpdateTime(); !got.Equal(updatedAt) {
		t.Fatalf("expected update time %s, got %s", updatedAt, got)
	}

	if got := resource.GetIDField(); got != "id" {
		t.Fatalf("expected id field, got %q", got)
	}

	if got := resource.GetCreationTimeField(); got != "created_at" {
		t.Fatalf("expected created_at field, got %q", got)
	}

	if got := resource.GetLastUpdateTimeField(); got != "updated_at" {
		t.Fatalf("expected updated_at field, got %q", got)
	}
}

func TestResourceSettersMutatePrimaryKeyAndTimestamps(t *testing.T) {
	t.Parallel()

	resource := Resource[int]{}
	resource.SetID(7)
	resource.SetCreationTime()
	resource.SetLastUpdateTime()

	if got := resource.GetID(); got != 7 {
		t.Fatalf("expected ID 7, got %d", got)
	}

	if resource.GetCreationTime().IsZero() {
		t.Fatal("expected creation time to be set")
	}

	if resource.GetLastUpdateTime().IsZero() {
		t.Fatal("expected update time to be set")
	}
}

func TestSoftDeletedResourceReportsActiveRecord(t *testing.T) {
	t.Parallel()

	resource := SoftDeletedResource[int]{
		Resource: Resource[int]{ID: 1},
	}

	if resource.IsDeleted() {
		t.Fatal("expected active resource to not be deleted")
	}

	if got := resource.GetDeletionTime(); got != nil {
		t.Fatalf("expected nil deletion time, got %s", got)
	}
}

func TestSoftDeletedResourceReportsDeletedRecord(t *testing.T) {
	t.Parallel()

	deletedAt := time.Date(2026, 9, 15, 11, 0, 0, 0, time.UTC)
	resource := SoftDeletedResource[string]{
		Resource: Resource[string]{ID: "user-1"},
		DeletedAt: gorm.DeletedAt{
			Time:  deletedAt,
			Valid: true,
		},
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

	if got := resource.GetDeletionTimeField(); got != "deleted_at" {
		t.Fatalf("expected deleted_at field, got %q", got)
	}
}

func TestSoftDeletedResourceSettersMutateDeletionTime(t *testing.T) {
	t.Parallel()

	resource := SoftDeletedResource[int]{}
	resource.SetDeletionTime()

	if !resource.IsDeleted() {
		t.Fatal("expected deletion time to mark resource as deleted")
	}

	if resource.GetDeletionTime() == nil {
		t.Fatal("expected deletion time to be set")
	}

	resource.UnsetDeletionTime()

	if resource.IsDeleted() {
		t.Fatal("expected deletion time to be unset")
	}
}
