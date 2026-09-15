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

func (r collectionTestResource) GetCreationTime() time.Time {
	return r.createdAt
}

func (r collectionTestResource) GetLastUpdateTime() time.Time {
	return r.updatedAt
}

func (r collectionTestResource) GetDeletionTime() *time.Time {
	return r.deletedAt
}

func (r collectionTestResource) IsDeleted() bool {
	return r.deletedAt != nil
}

type collectionTestStore struct{}

func (collectionTestStore) Get(id string) (collectionTestResource, bool, error) {
	return collectionTestResource{id: id}, true, nil
}

func (collectionTestStore) List(options ListOptions) ([]collectionTestResource, error) {
	return []collectionTestResource{}, nil
}

func (collectionTestStore) GetDeleted(id string) (collectionTestResource, bool, error) {
	now := time.Now()
	return collectionTestResource{id: id, deletedAt: &now}, true, nil
}

func (collectionTestStore) ListDeleted(options ListOptions) ([]collectionTestResource, error) {
	return []collectionTestResource{}, nil
}

func (collectionTestStore) Delete(id string) (bool, error) {
	return true, nil
}

func (collectionTestStore) DeleteMany(ids []string) (int, error) {
	return len(ids), nil
}

func (collectionTestStore) Prune(id string) (bool, error) {
	return true, nil
}

func (collectionTestStore) PruneMany(ids []string) (int, error) {
	return len(ids), nil
}

func (collectionTestStore) Restore(id string) (bool, error) {
	return true, nil
}

func (collectionTestStore) RestoreMany(ids []string) (int, error) {
	return len(ids), nil
}

func (collectionTestStore) UpdateOne(id string, values map[string]any) (bool, error) {
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
			Get(id string) (collectionTestResource, bool, error)
		} = collectionTestStore{}
	)
}
