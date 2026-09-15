package types

import "testing"

type singletonTestStore struct{}

func (singletonTestStore) Get() (collectionTestResource, bool, error) {
	return collectionTestResource{id: "singleton"}, true, nil
}

func (singletonTestStore) GetDeleted() (collectionTestResource, bool, error) {
	return collectionTestResource{id: "singleton"}, true, nil
}

func (singletonTestStore) Delete() (bool, error) {
	return true, nil
}

func (singletonTestStore) Prune() (bool, error) {
	return true, nil
}

func (singletonTestStore) Restore() (bool, error) {
	return true, nil
}

func (singletonTestStore) Update(values map[string]any) (bool, error) {
	return true, nil
}

func (singletonTestStore) Create(element collectionTestResource) (bool, error) {
	return true, nil
}

func TestSingletonInterfaces(t *testing.T) {
	t.Parallel()

	var (
		_ SingletonGet[string, collectionTestResource]        = singletonTestStore{}
		_ SingletonDeletedGet[string, collectionTestResource] = singletonTestStore{}
		_ SingletonDelete                                     = singletonTestStore{}
		_ SingletonSoftDeletedPrune                           = singletonTestStore{}
		_ SingletonSoftDeletedRestore                         = singletonTestStore{}
		_ SingletonUpdate[string, collectionTestResource]     = singletonTestStore{}
		_ SingletonCreate[string, collectionTestResource]     = singletonTestStore{}
	)
}
