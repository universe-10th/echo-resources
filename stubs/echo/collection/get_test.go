package collection

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/universe-10th/echo-resources/types"
)

type getTestResource struct {
	id int
}

func (r getTestResource) GetID() int {
	return r.id
}

func (r getTestResource) SetID(id int) {}

func (r getTestResource) GetIDField() string {
	return "id"
}

func (r getTestResource) GetCreationTime() time.Time {
	return time.Time{}
}

func (r getTestResource) GetLastUpdateTime() time.Time {
	return time.Time{}
}

func (r getTestResource) SetCreationTime() {}

func (r getTestResource) SetCreationTimeIn(location *time.Location) {}

func (r getTestResource) SetLastUpdateTime() {}

func (r getTestResource) SetLastUpdateTimeIn(location *time.Location) {}

func (r getTestResource) GetCreationTimeField() string {
	return "created_at"
}

func (r getTestResource) GetLastUpdateTimeField() string {
	return "updated_at"
}

func (r getTestResource) GetDeletionTime() *time.Time {
	return nil
}

func (r getTestResource) IsDeleted() bool {
	return r.GetDeletionTime() != nil
}

func (r getTestResource) SetDeletionTime() {}

func (r getTestResource) SetDeletionTimeIn(location *time.Location) {}

func (r getTestResource) UnsetDeletionTime() {}

func (r getTestResource) GetDeletionTimeField() string {
	return "deleted_at"
}

type getTestEngine struct {
	calls []string
}

func (e *getTestEngine) URLArg() string {
	return "id"
}

func (e *getTestEngine) ParseID(context echo.Context) (int, error) {
	e.calls = append(e.calls, "ParseID")
	return 42, nil
}

func (e *getTestEngine) ApplyPathConstraints(context echo.Context, filter *types.FilterExpression) error {
	e.calls = append(e.calls, "ApplyPathConstraints")
	return nil
}

func (e *getTestEngine) ApplyDeletedFilter(filter *types.FilterExpression, deleted bool) {
	if deleted {
		e.calls = append(e.calls, "ApplyDeletedFilter:true")
		return
	}
	e.calls = append(e.calls, "ApplyDeletedFilter:false")
}

func (e *getTestEngine) ApplyCustomFilter(context echo.Context, filter *types.FilterExpression) error {
	e.calls = append(e.calls, "ApplyCustomFilter")
	return nil
}

func (e *getTestEngine) RetrieveElement(id int, filter *types.FilterExpression) (getTestResource, bool, error) {
	e.calls = append(e.calls, "RetrieveElement")
	return getTestResource{id: id}, true, nil
}

func (e *getTestEngine) RenderElement(context echo.Context, element getTestResource) error {
	e.calls = append(e.calls, "RenderElement")
	return nil
}

func newGetTestContext() echo.Context {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/resources/42", nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}

func TestGetEndpointStubFollowsExpectedOrder(t *testing.T) {
	engine := &getTestEngine{}
	stub := GetEndpointStub[int, getTestResource]{engine: engine}

	if err := stub.Get(newGetTestContext()); err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	expected := []string{
		"ParseID",
		"ApplyCustomFilter",
		"ApplyPathConstraints",
		"ApplyDeletedFilter:false",
		"RetrieveElement",
		"RenderElement",
	}
	if !reflect.DeepEqual(engine.calls, expected) {
		t.Fatalf("calls = %#v, want %#v", engine.calls, expected)
	}
}

func TestSoftDeletedGetEndpointStubFiltersDeletedElements(t *testing.T) {
	engine := &getTestEngine{}
	stub := SoftDeletedGetEndpointStub[int, getTestResource]{engine: engine}

	if err := stub.GetDeleted(newGetTestContext()); err != nil {
		t.Fatalf("GetDeleted returned error: %v", err)
	}

	expected := []string{
		"ParseID",
		"ApplyCustomFilter",
		"ApplyPathConstraints",
		"ApplyDeletedFilter:true",
		"RetrieveElement",
		"RenderElement",
	}
	if !reflect.DeepEqual(engine.calls, expected) {
		t.Fatalf("calls = %#v, want %#v", engine.calls, expected)
	}
}
