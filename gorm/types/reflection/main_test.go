package reflection

import (
	"reflect"
	"testing"
)

type FieldToStorageEmbedded struct {
	ID        int
	CreatedAt string `gorm:"column:created_at"`
}

type fieldToStorageModel struct {
	FieldToStorageEmbedded
	Name    string `gorm:"column:product_name"`
	Ignored string `gorm:"-"`
	Count   int
}

func TestFieldToStorageUsesGORMColumnsAndSnakeCase(t *testing.T) {
	t.Parallel()

	got := FieldToStorage(fieldToStorageModel{})
	want := map[string]string{
		"ID":        "id",
		"CreatedAt": "created_at",
		"Name":      "product_name",
		"Count":     "count",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected mapping\nwant: %#v\n got: %#v", want, got)
	}
}

func TestFieldToStorageAcceptsNilPointers(t *testing.T) {
	t.Parallel()

	var value *fieldToStorageModel

	got := FieldToStorage(value)
	want := map[string]string{
		"ID":        "id",
		"CreatedAt": "created_at",
		"Name":      "product_name",
		"Count":     "count",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected mapping\nwant: %#v\n got: %#v", want, got)
	}
}
