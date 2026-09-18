package reflection

import (
	"reflect"
	"testing"
)

type FieldToStorageEmbedded struct {
	ID string `bson:"_id,omitempty"`
}

type fieldToStorageModel struct {
	FieldToStorageEmbedded `bson:",inline"`
	Name                   string `bson:"product_name"`
	Ignored                string `bson:"-"`
	Count                  int
}

func TestFieldToStorageUsesBSONTagsAndInlineFields(t *testing.T) {
	t.Parallel()

	got := FieldToStorage(fieldToStorageModel{})
	want := map[string]string{
		"ID":    "_id",
		"Name":  "product_name",
		"Count": "count",
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
		"ID":    "_id",
		"Name":  "product_name",
		"Count": "count",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected mapping\nwant: %#v\n got: %#v", want, got)
	}
}
