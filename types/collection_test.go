package types

import (
	"testing"
	"time"
)

type collectionTestResource struct {
	id        string
	createdAt time.Time
	updatedAt time.Time
	deletedAt *time.Time
}

func (r collectionTestResource) GetID() string {
	return r.id
}

func (r collectionTestResource) SetID(id string) {}

func (r collectionTestResource) GetIDField() string {
	return "id"
}

func (r collectionTestResource) GetCreationTime() time.Time {
	return r.createdAt
}

func (r collectionTestResource) GetLastUpdateTime() time.Time {
	return r.updatedAt
}

func (r collectionTestResource) SetCreationTime() {}

func (r collectionTestResource) SetCreationTimeIn(location *time.Location) {}

func (r collectionTestResource) RestoreCreationTime(stamp time.Time) {}

func (r collectionTestResource) SetLastUpdateTime() {}

func (r collectionTestResource) SetLastUpdateTimeIn(location *time.Location) {}

func (r collectionTestResource) GetCreationTimeField() string {
	return "created_at"
}

func (r collectionTestResource) GetLastUpdateTimeField() string {
	return "updated_at"
}

func (r collectionTestResource) GetDeletionTime() *time.Time {
	return r.deletedAt
}

func (r collectionTestResource) IsDeleted() bool {
	return r.deletedAt != nil
}

func (r collectionTestResource) SetDeletionTime() {}

func (r collectionTestResource) SetDeletionTimeIn(location *time.Location) {}

func (r collectionTestResource) UnsetDeletionTime() {}

func (r collectionTestResource) GetDeletionTimeField() string {
	return "deleted_at"
}

type collectionTestStore struct{}

func (collectionTestStore) Get(id string, filter FilterExpression) (collectionTestResource, bool, error) {
	return collectionTestResource{id: id}, true, nil
}

func (collectionTestStore) List(options ListOptions) ([]collectionTestResource, error) {
	return []collectionTestResource{}, nil
}

func (collectionTestStore) Count(options FilterExpression) (int64, error) {
	return 0, nil
}

func (collectionTestStore) GetDeleted(id string, filter FilterExpression) (collectionTestResource, bool, error) {
	now := time.Now()
	return collectionTestResource{id: id, deletedAt: &now}, true, nil
}

func (collectionTestStore) ListDeleted(options ListOptions) ([]collectionTestResource, error) {
	return []collectionTestResource{}, nil
}

func (collectionTestStore) CountDeleted(options FilterExpression) (int64, error) {
	return 0, nil
}

func (collectionTestStore) Delete(id string, filter FilterExpression) (bool, error) {
	return true, nil
}

func (collectionTestStore) DeleteMany(ids []string, filter FilterExpression) (int, error) {
	return len(ids), nil
}

func (collectionTestStore) Prune(id string, filter FilterExpression) (bool, error) {
	return true, nil
}

func (collectionTestStore) PruneMany(ids []string, filter FilterExpression) (int, error) {
	return len(ids), nil
}

func (collectionTestStore) Restore(id string, filter FilterExpression) (bool, error) {
	return true, nil
}

func (collectionTestStore) RestoreMany(ids []string, filter FilterExpression) (int, error) {
	return len(ids), nil
}

func (collectionTestStore) UpdateOne(id string, filter FilterExpression, values map[string]any) (bool, error) {
	return true, nil
}

func (collectionTestStore) CreateOne(element collectionTestResource) (string, error) {
	return element.GetID(), nil
}

func TestCollectionInterfaces(t *testing.T) {
	t.Parallel()

	var (
		_ CollectionList[string, collectionTestResource]            = collectionTestStore{}
		_ CollectionSoftDeletedList[string, collectionTestResource] = collectionTestStore{}
		_ CollectionDelete[string]                                  = collectionTestStore{}
		_ CollectionSoftDeletedPrune[string]                        = collectionTestStore{}
		_ CollectionSoftDeletedRestore[string]                      = collectionTestStore{}
		_ CollectionUpdate[string, collectionTestResource]          = collectionTestStore{}
		_ CollectionCreate[string, collectionTestResource]          = collectionTestStore{}
		_ SoftDeletedResource[string]                               = collectionTestResource{}
		_ Resource[string]                                          = collectionTestResource{}
		_ Identified[string]                                        = collectionTestResource{}
		_ interface {
			Get(id string, filter FilterExpression) (collectionTestResource, bool, error)
		} = collectionTestStore{}
	)
}
