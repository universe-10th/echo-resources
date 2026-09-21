package singleton

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/universe-10th/echo-resources/types"
)

type singletonTestResource struct {
	id      int
	deleted bool
}

func (r singletonTestResource) GetID() int {
	return r.id
}

func (r singletonTestResource) SetID(id int) {}

func (r singletonTestResource) GetIDField() string {
	return "id"
}

func (r singletonTestResource) GetCreationTime() time.Time {
	return time.Time{}
}

func (r singletonTestResource) GetLastUpdateTime() time.Time {
	return time.Time{}
}

func (r singletonTestResource) SetCreationTime() {}

func (r singletonTestResource) SetCreationTimeIn(location *time.Location) {}

func (r singletonTestResource) SetLastUpdateTime() {}

func (r singletonTestResource) SetLastUpdateTimeIn(location *time.Location) {}

func (r singletonTestResource) GetCreationTimeField() string {
	return "created_at"
}

func (r singletonTestResource) GetLastUpdateTimeField() string {
	return "updated_at"
}

func (r singletonTestResource) GetDeletionTime() *time.Time {
	if !r.deleted {
		return nil
	}
	t := time.Time{}
	return &t
}

func (r singletonTestResource) IsDeleted() bool {
	return r.deleted
}

func (r singletonTestResource) SetDeletionTime() {}

func (r singletonTestResource) SetDeletionTimeIn(location *time.Location) {}

func (r singletonTestResource) UnsetDeletionTime() {}

func (r singletonTestResource) GetDeletionTimeField() string {
	return "deleted_at"
}

type singletonCreateTestEngine struct {
	calls         []string
	existing      singletonTestResource
	existingFound bool
}

func (e *singletonCreateTestEngine) ApplyDeletedFilter(filter *types.FilterExpression, deleted bool) {
	if deleted {
		e.calls = append(e.calls, "ApplyDeletedFilter:true")
		return
	}
	e.calls = append(e.calls, "ApplyDeletedFilter:false")
}

func (e *singletonCreateTestEngine) ApplyPathConstraints(context echo.Context, filter *types.FilterExpression) error {
	e.calls = append(e.calls, "ApplyPathConstraints")
	return nil
}

func (e *singletonCreateTestEngine) ApplyCustomFilter(context echo.Context, filter *types.FilterExpression) error {
	e.calls = append(e.calls, "ApplyCustomFilter")
	return nil
}

func (e *singletonCreateTestEngine) RetrieveElement(
	filter *types.FilterExpression,
) (singletonTestResource, bool, error) {
	e.calls = append(e.calls, "RetrieveElement")
	return e.existing, e.existingFound, nil
}

func (e *singletonCreateTestEngine) ReadBody(context echo.Context, element *singletonTestResource) error {
	e.calls = append(e.calls, "ReadBody")
	element.id = 100
	return nil
}

func (e *singletonCreateTestEngine) ApplyPathConstraintsToElement(
	context echo.Context, element *singletonTestResource,
) error {
	e.calls = append(e.calls, "ApplyPathConstraintsToElement:"+strconv.Itoa(element.id))
	element.id = 42
	return nil
}

func (e *singletonCreateTestEngine) Validate(element *singletonTestResource) error {
	e.calls = append(e.calls, "Validate:"+strconv.Itoa(element.id))
	return nil
}

func (e *singletonCreateTestEngine) Save(element *singletonTestResource) error {
	e.calls = append(e.calls, "Save:"+strconv.Itoa(element.id))
	return nil
}

func (e *singletonCreateTestEngine) RenderElement(
	context echo.Context, element singletonTestResource, created bool,
) error {
	e.calls = append(e.calls, "RenderElement:"+strconv.Itoa(element.id)+":"+strconv.FormatBool(created))
	return nil
}

func newSingletonCreateTestContext() (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/resource", nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestCreateEndpointStubCreatesWhenScopedSingletonDoesNotExist(t *testing.T) {
	engine := &singletonCreateTestEngine{}
	stub := CreateEndpointStub[int, singletonTestResource]{engine: engine}
	context, _ := newSingletonCreateTestContext()

	if err := stub.Create(context); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	expected := []string{
		"ApplyCustomFilter",
		"ApplyPathConstraints",
		"RetrieveElement",
		"ReadBody",
		"ApplyPathConstraintsToElement:100",
		"Validate:42",
		"Save:42",
		"RenderElement:42:true",
	}
	if !reflect.DeepEqual(engine.calls, expected) {
		t.Fatalf("calls = %#v, want %#v", engine.calls, expected)
	}
}

func TestCreateEndpointStubActiveExistingSingletonReturnsConflict(t *testing.T) {
	engine := &singletonCreateTestEngine{
		existing:      singletonTestResource{id: 1},
		existingFound: true,
	}
	stub := CreateEndpointStub[int, singletonTestResource]{engine: engine}
	context, recorder := newSingletonCreateTestContext()

	if err := stub.Create(context); err == nil {
		t.Fatal("Create returned nil error")
	}

	if recorder.Code != int(types.ErrConflict) {
		t.Fatalf("status = %d, want %d", recorder.Code, types.ErrConflict)
	}

	expected := []string{
		"ApplyCustomFilter",
		"ApplyPathConstraints",
		"RetrieveElement",
	}
	if !reflect.DeepEqual(engine.calls, expected) {
		t.Fatalf("calls = %#v, want %#v", engine.calls, expected)
	}
}

func TestCreateEndpointStubDeletedExistingSingletonSuggestsRestore(t *testing.T) {
	engine := &singletonCreateTestEngine{
		existing:      singletonTestResource{id: 1, deleted: true},
		existingFound: true,
	}
	stub := CreateEndpointStub[int, singletonTestResource]{engine: engine}
	context, recorder := newSingletonCreateTestContext()

	if err := stub.Create(context); err == nil {
		t.Fatal("Create returned nil error")
	}

	if recorder.Code != int(types.ErrConflict) {
		t.Fatalf("status = %d, want %d", recorder.Code, types.ErrConflict)
	}

	expected := []string{
		"ApplyCustomFilter",
		"ApplyPathConstraints",
		"RetrieveElement",
	}
	if !reflect.DeepEqual(engine.calls, expected) {
		t.Fatalf("calls = %#v, want %#v", engine.calls, expected)
	}
}
