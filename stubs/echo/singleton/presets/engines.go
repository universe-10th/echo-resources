package presets

import (
	"github.com/labstack/echo/v4"
	"github.com/universe-10th/echo-resources/types"
)

// DumbReadWriteResourceServiceEngine is a no-op implementation of
// ReadWriteResourceServiceEngine.
type DumbReadWriteResourceServiceEngine[IDT comparable, RT types.Resource[IDT]] struct {
	ResourceFiltering[IDT, RT]
}

// NewDumbReadWriteResourceServiceEngine creates a no-op read-write resource
// service engine.
func NewDumbReadWriteResourceServiceEngine[IDT comparable, RT types.Resource[IDT]]() DumbReadWriteResourceServiceEngine[IDT, RT] {
	return DumbReadWriteResourceServiceEngine[IDT, RT]{}
}

func (engine DumbReadWriteResourceServiceEngine[IDT, RT]) ApplyPathConstraintsToFilter(
	context echo.Context, filter *types.FilterExpression,
) error {
	return nil
}

func (engine DumbReadWriteResourceServiceEngine[IDT, RT]) ApplyPathConstraintsToElement(
	context echo.Context, element *RT,
) error {
	return nil
}

func (engine DumbReadWriteResourceServiceEngine[IDT, RT]) RetrieveElement(
	filter *types.FilterExpression,
) (RT, bool, error) {
	var zero RT
	return zero, false, nil
}

func (engine DumbReadWriteResourceServiceEngine[IDT, RT]) Delete(id IDT) error {
	return nil
}

func (engine DumbReadWriteResourceServiceEngine[IDT, RT]) Restore(id IDT) (RT, error) {
	var zero RT
	return zero, nil
}

func (engine DumbReadWriteResourceServiceEngine[IDT, RT]) Prune(id IDT) error {
	return nil
}

func (engine DumbReadWriteResourceServiceEngine[IDT, RT]) ReadBody(
	context echo.Context, element *RT,
) error {
	return nil
}

func (engine DumbReadWriteResourceServiceEngine[IDT, RT]) Save(element *RT) error {
	return nil
}

// DumbReadWriteSoftDeletedResourceServiceEngine is a no-op implementation of
// ReadWriteSoftDeletedResourceServiceEngine.
type DumbReadWriteSoftDeletedResourceServiceEngine[
	IDT comparable,
	RT types.SoftDeletedResource[IDT],
] struct {
	DumbReadWriteResourceServiceEngine[IDT, RT]
}

// NewDumbReadWriteSoftDeletedResourceServiceEngine creates a no-op read-write
// soft-deleted resource service engine.
func NewDumbReadWriteSoftDeletedResourceServiceEngine[
	IDT comparable,
	RT types.SoftDeletedResource[IDT],
]() DumbReadWriteSoftDeletedResourceServiceEngine[IDT, RT] {
	return DumbReadWriteSoftDeletedResourceServiceEngine[IDT, RT]{}
}
