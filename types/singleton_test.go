package types

import "testing"

type singletonTestStore struct{}

func (singletonTestStore) Get(filter FilterExpression) (collectionTestResource, bool, error) {
	return collectionTestResource{id: "singleton"}, true, nil
}

func (singletonTestStore) GetDeleted(filter FilterExpression) (collectionTestResource, bool, error) {
	return collectionTestResource{id: "singleton"}, true, nil
}

func (singletonTestStore) Delete(filter FilterExpression) (bool, error) {
	return true, nil
}

func (singletonTestStore) Prune(filter FilterExpression) (bool, error) {
	return true, nil
}

func (singletonTestStore) Restore(filter FilterExpression) (bool, error) {
	return true, nil
}

func (singletonTestStore) Update(filter FilterExpression, values map[string]any) (bool, error) {
	return true, nil
}

func (singletonTestStore) Create(element collectionTestResource) (bool, error) {
	return true, nil
}

func TestSingletonInterfaces(t *testing.T) {
	t.Parallel()

	var (
		_ SingletonGet[string, collectionTestResource]            = singletonTestStore{}
		_ SingletonSoftDeletedGet[string, collectionTestResource] = singletonTestStore{}
		_ SingletonDelete                                         = singletonTestStore{}
		_ SingletonSoftDeletedPrune                               = singletonTestStore{}
		_ SingletonSoftDeletedRestore                             = singletonTestStore{}
		_ SingletonUpdate[string, collectionTestResource]         = singletonTestStore{}
		_ SingletonCreate[string, collectionTestResource]         = singletonTestStore{}
	)
}
