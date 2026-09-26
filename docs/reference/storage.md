# Storage Adapters

Services call the `types.Storage[IDT, RT]` interface. Built-in adapters provide
memory, GORM, and MongoDB implementations.

## Interface

```go
type Storage[IDT comparable, RT Resource[IDT]] interface {
	Mapping() *FieldsMapping
	GetElement(filter *FilterExpression) (RT, bool, error)
	GetElements(filter *FilterExpression, sort *SortExpression, skip, limit int64) ([]RT, int64, error)
	Save(element *RT) (notFound bool, err error)
	Delete(element *RT) (notFound bool, err error)
	ValidateFilter(filter *FilterExpression) error
	ValidateSort(sort *SortExpression) error
	AddIDFilter(filter *FilterExpression, id IDT)
	Restore(element *RT) (notFound bool, err error)
	Prune(element *RT) (notFound bool, err error)
	AddDeletedFilter(filter *FilterExpression, deleted bool)
}
```

`Save` creates when the ID is zero and updates otherwise. `Delete` hard-deletes
ordinary resources and soft-deletes resources that implement
`types.SoftDeletedResource`.

## `memory`

Use for tests, prototypes, and ephemeral storage.

Import paths:

```go
github.com/universe-10th/rest-resources/memory
```

Public model fragments:

- `Resource[T memory.PrimaryKey]`
- `SoftDeletedResource[T memory.PrimaryKey]`

Storage constructors:

- `NewStorage[IDT, RT]() *Storage[IDT, RT]`
- `NewStorageWithIDGenerator[IDT, RT](generator IDGenerator[IDT]) *Storage[IDT, RT]`

Field mapping:

- `FieldToStorage(v any) map[string]string`

## `gorm`

Use with a caller-owned `*gorm.DB`.

Import paths:

```go
github.com/universe-10th/rest-resources/gorm
github.com/universe-10th/rest-resources/gorm/types
```

Storage constructor:

- `NewStorage[IDT, RT](db *gorm.DB) *Storage[IDT, RT]`

The `gorm/types` package provides:

- `Resource[T gormtypes.PrimaryKey]`
- `SoftDeletedResource[T gormtypes.PrimaryKey]`
- `PrimaryKey`
- `FieldToStorage`
- `NewFieldsMapping`
- `NewFilterSource`, `NewFilterValidator`, `NewFilterSerializer`
- `NewSortSource`, `NewSortValidator`, `NewSortSerializer`

GORM soft deletes use `gorm.DeletedAt`.

## `mongo`

Use with a caller-owned MongoDB collection.

Import paths:

```go
github.com/universe-10th/rest-resources/mongo
github.com/universe-10th/rest-resources/mongo/types
```

Storage constructors:

- `NewStorage[IDT, RT](collection *mongo.Collection) *Storage[IDT, RT]`
- `NewStorageWithContext[IDT, RT](collection *mongo.Collection, contextProvider func() context.Context) *Storage[IDT, RT]`

The `mongo/types` package provides:

- `Resource`
- `SoftDeletedResource`
- `SoftDeleteIndexModel`
- `FieldToStorage`
- `NewFieldsMapping`
- `NewFilterSource`, `NewFilterValidator`, `NewFilterSerializer`
- `NewSortSource`, `NewSortValidator`, `NewSortSerializer`

MongoDB resources use `bson.ObjectID`.
