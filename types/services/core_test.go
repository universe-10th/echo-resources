package services

import (
	"errors"
	"testing"
	"time"

	"github.com/universe-10th/echo-resources/types"
)

type coreConstraintResource struct {
	ID       int `json:"id"`
	ParentID int `json:"parent_id"`
}

func (r coreConstraintResource) GetID() int                         { return r.ID }
func (r coreConstraintResource) SetID(int)                          {}
func (r coreConstraintResource) GetIDField() string                 { return "id" }
func (r coreConstraintResource) GetCreationTime() time.Time         { return time.Time{} }
func (r coreConstraintResource) GetLastUpdateTime() time.Time       { return time.Time{} }
func (r coreConstraintResource) SetCreationTime()                   {}
func (r coreConstraintResource) SetCreationTimeIn(*time.Location)   {}
func (r coreConstraintResource) RestoreCreationTime(time.Time)      {}
func (r coreConstraintResource) SetLastUpdateTime()                 {}
func (r coreConstraintResource) SetLastUpdateTimeIn(*time.Location) {}
func (r coreConstraintResource) GetCreationTimeField() string       { return "created_at" }
func (r coreConstraintResource) GetLastUpdateTimeField() string     { return "updated_at" }

type coreConstraintBadResource struct {
	ID       int    `json:"id"`
	ParentID string `json:"parent_id"`
}

func (r coreConstraintBadResource) GetID() int                         { return r.ID }
func (r coreConstraintBadResource) SetID(int)                          {}
func (r coreConstraintBadResource) GetIDField() string                 { return "id" }
func (r coreConstraintBadResource) GetCreationTime() time.Time         { return time.Time{} }
func (r coreConstraintBadResource) GetLastUpdateTime() time.Time       { return time.Time{} }
func (r coreConstraintBadResource) SetCreationTime()                   {}
func (r coreConstraintBadResource) SetCreationTimeIn(*time.Location)   {}
func (r coreConstraintBadResource) RestoreCreationTime(time.Time)      {}
func (r coreConstraintBadResource) SetLastUpdateTime()                 {}
func (r coreConstraintBadResource) SetLastUpdateTimeIn(*time.Location) {}
func (r coreConstraintBadResource) GetCreationTimeField() string       { return "created_at" }
func (r coreConstraintBadResource) GetLastUpdateTimeField() string     { return "updated_at" }

func TestReadBindsJSONBody(t *testing.T) {
	t.Parallel()

	var element coreConstraintResource
	context := coreConstraintContext{
		contentType: "application/json; charset=utf-8",
		bind: func(target any) error {
			resource := target.(*coreConstraintResource)
			resource.ID = 42
			return nil
		},
	}
	service := ResourceService[int, coreConstraintResource]{}

	err := service.read(context, &element)
	if err != nil {
		t.Fatalf("read returned error: %v", err)
	}
	if element.ID != 42 {
		t.Fatalf("expected ID 42, got %d", element.ID)
	}
}

func TestReadUsesCustomReader(t *testing.T) {
	t.Parallel()

	var element coreConstraintResource
	service := ResourceService[int, coreConstraintResource]{}
	service.UsingReader(
		func(context Context, element *coreConstraintResource) error {
			element.ID = 84
			return nil
		},
	)

	err := service.read(coreConstraintContext{contentType: "application/json"}, &element)
	if err != nil {
		t.Fatalf("read returned error: %v", err)
	}
	if element.ID != 84 {
		t.Fatalf("expected ID 84, got %d", element.ID)
	}
}

func TestReadRejectsNonJSONContent(t *testing.T) {
	t.Parallel()

	var element coreConstraintResource
	service := ResourceService[int, coreConstraintResource]{}

	err := service.read(coreConstraintContext{contentType: "text/plain"}, &element)
	var badRequest types.BadRequestError
	if !errors.As(err, &badRequest) {
		t.Fatalf("expected BadRequestError, got %T: %v", err, err)
	}
}

func TestApplyPreviousConstraintSetsMappedField(t *testing.T) {
	t.Parallel()

	parent := coreConstraintResource{ID: 42}
	element := coreConstraintResource{ID: 100}
	service := ResourceService[int, coreConstraintResource]{
		prefix:              "children",
		storage:             newCoreConstraintStorage[int, coreConstraintResource](),
		constraintJSONField: "parent_id",
	}

	err := service.applyPreviousConstraint(coreConstraintContext{element: parent}, &element)
	if err != nil {
		t.Fatalf("applyPreviousConstraint returned error: %v", err)
	}
	if element.ParentID != parent.ID {
		t.Fatalf("expected ParentID %d, got %d", parent.ID, element.ParentID)
	}
}

func TestApplyConstraintReturnsInternalErrorForUnmappedField(t *testing.T) {
	t.Parallel()

	parent := coreConstraintResource{ID: 42}
	element := coreConstraintResource{ID: 100}
	service := ResourceService[int, coreConstraintResource]{
		prefix:              "children",
		storage:             newCoreConstraintStorage[int, coreConstraintResource](),
		constraintJSONField: "missing",
	}

	err := service.applyPreviousConstraint(coreConstraintContext{element: parent}, &element)
	var internal types.InternalError
	if !errors.As(err, &internal) {
		t.Fatalf("expected InternalError, got %T: %v", err, err)
	}
}

func TestApplyConstraintReturnsInternalErrorForIncompatibleField(t *testing.T) {
	t.Parallel()

	parent := coreConstraintBadResource{ID: 42}
	element := coreConstraintBadResource{ID: 100}
	service := ResourceService[int, coreConstraintBadResource]{
		prefix:              "children",
		storage:             newCoreConstraintStorage[int, coreConstraintBadResource](),
		constraintJSONField: "parent_id",
	}

	err := service.applyPreviousConstraint(coreConstraintContext{element: parent}, &element)
	var internal types.InternalError
	if !errors.As(err, &internal) {
		t.Fatalf("expected InternalError, got %T: %v", err, err)
	}
}

type coreConstraintStorage[IDT comparable, RT types.Resource[IDT]] struct {
	mapping *types.FieldsMapping
}

func newCoreConstraintStorage[IDT comparable, RT types.Resource[IDT]]() coreConstraintStorage[IDT, RT] {
	return coreConstraintStorage[IDT, RT]{
		mapping: types.NewFieldsMapping[IDT, RT](func(any) map[string]string {
			return map[string]string{
				"ID":       "id",
				"ParentID": "parent_id",
			}
		}),
	}
}

func (s coreConstraintStorage[IDT, RT]) Mapping() *types.FieldsMapping { return s.mapping }
func (s coreConstraintStorage[IDT, RT]) GetElement(*types.FilterExpression) (RT, bool, error) {
	var zero RT
	return zero, false, nil
}
func (s coreConstraintStorage[IDT, RT]) GetElements(*types.FilterExpression, *types.SortExpression, int64, int64) ([]RT, int64, error) {
	return nil, 0, nil
}
func (s coreConstraintStorage[IDT, RT]) Save(*RT) (bool, error)    { return false, nil }
func (s coreConstraintStorage[IDT, RT]) Delete(*RT) (bool, error)  { return false, nil }
func (s coreConstraintStorage[IDT, RT]) Restore(*RT) (bool, error) { return false, nil }
func (s coreConstraintStorage[IDT, RT]) Prune(*RT) (bool, error)   { return false, nil }
func (s coreConstraintStorage[IDT, RT]) ValidateFilter(*types.FilterExpression) error {
	return nil
}
func (s coreConstraintStorage[IDT, RT]) ValidateSort(*types.SortExpression) error {
	return nil
}
func (s coreConstraintStorage[IDT, RT]) AddIDFilter(*types.FilterExpression, IDT)       {}
func (s coreConstraintStorage[IDT, RT]) AddDeletedFilter(*types.FilterExpression, bool) {}

type coreConstraintContext struct {
	element     any
	contentType string
	bind        func(any) error
}

func (c coreConstraintContext) Native() any                             { return nil }
func (c coreConstraintContext) GetPathParam(string) (string, error)     { return "", nil }
func (c coreConstraintContext) GetQueryParam(string) (string, error)    { return "", nil }
func (c coreConstraintContext) GetQueryParams(string) ([]string, error) { return nil, nil }
func (c coreConstraintContext) GetHeader(name string) (string, error) {
	if name == "Content-Type" {
		return c.contentType, nil
	}
	return "", nil
}
func (c coreConstraintContext) GetHeaders(string) ([]string, error) { return nil, nil }
func (c coreConstraintContext) GetCookie(string) (Cookie, error)    { return Cookie{}, nil }
func (c coreConstraintContext) BindJSON(target any) error {
	if c.bind != nil {
		return c.bind(target)
	}
	return nil
}
func (c coreConstraintContext) SetHeader(string, string)    {}
func (c coreConstraintContext) SetCookie(Cookie)            {}
func (c coreConstraintContext) GetData(string) (any, bool)  { return nil, false }
func (c coreConstraintContext) SetData(string, any)         {}
func (c coreConstraintContext) PushElement(any)             {}
func (c coreConstraintContext) PopElement() (any, bool)     { return nil, false }
func (c coreConstraintContext) PeekElement(int) (any, bool) { return c.element, c.element != nil }
func (c coreConstraintContext) RenderJSON(int, any) error   { return nil }
func (c coreConstraintContext) RenderNoContent(int) error   { return nil }
func (c coreConstraintContext) CurrentService() any         { return nil }
func (c coreConstraintContext) CurrentEndpoint() (EndpointType, ResourceVerb, string) {
	return EndpointVerb, ResourceGet, ""
}
func (c coreConstraintContext) Setup(any, EndpointType, ResourceVerb, string) {}
