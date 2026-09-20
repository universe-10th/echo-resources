package collection

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/universe-10th/echo-resources/types"
)

type restoreTestEngine struct {
	calls      []string
	restoreErr error
}

func (e *restoreTestEngine) ParseID(context echo.Context) (int, error) {
	e.calls = append(e.calls, "ParseID")
	return 42, nil
}

func (e *restoreTestEngine) ApplyPathConstraints(context echo.Context, filter *types.FilterExpression) error {
	e.calls = append(e.calls, "ApplyPathConstraints")
	return nil
}

func (e *restoreTestEngine) ApplyDeletedFilter(filter *types.FilterExpression, deleted bool) {
	if deleted {
		e.calls = append(e.calls, "ApplyDeletedFilter:true")
		return
	}
	e.calls = append(e.calls, "ApplyDeletedFilter:false")
}

func (e *restoreTestEngine) ApplyCustomFilter(context echo.Context, filter *types.FilterExpression) error {
	e.calls = append(e.calls, "ApplyCustomFilter")
	return nil
}

func (e *restoreTestEngine) RetrieveElement(id int, filter *types.FilterExpression) (getTestResource, bool, error) {
	e.calls = append(e.calls, "RetrieveElement:"+strconv.Itoa(id))
	return getTestResource{id: 100}, true, nil
}

func (e *restoreTestEngine) Restore(id int) (getTestResource, error) {
	e.calls = append(e.calls, "Restore:"+strconv.Itoa(id))
	return getTestResource{id: 200}, e.restoreErr
}

func (e *restoreTestEngine) RenderElement(context echo.Context, element getTestResource) error {
	e.calls = append(e.calls, "RenderElement:"+strconv.Itoa(element.id))
	return nil
}

func newRestoreTestContext() (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/resources/42/restore", nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestRestoreEndpointStubFollowsExpectedOrder(t *testing.T) {
	engine := &restoreTestEngine{}
	stub := RestoreEndpointStub[int, getTestResource]{engine: engine}
	context, _ := newRestoreTestContext()

	if err := stub.Restore(context); err != nil {
		t.Fatalf("Restore returned error: %v", err)
	}

	expected := []string{
		"ParseID",
		"ApplyCustomFilter",
		"ApplyPathConstraints",
		"ApplyDeletedFilter:true",
		"RetrieveElement:42",
		"Restore:100",
		"RenderElement:200",
	}
	if !reflect.DeepEqual(engine.calls, expected) {
		t.Fatalf("calls = %#v, want %#v", engine.calls, expected)
	}
}

func TestRestoreEndpointStubRestoreErrorReturnsTypedError(t *testing.T) {
	engine := &restoreTestEngine{
		restoreErr: types.NotDeletedError{},
	}
	stub := RestoreEndpointStub[int, getTestResource]{engine: engine}
	context, recorder := newRestoreTestContext()

	if err := stub.Restore(context); err == nil {
		t.Fatal("Restore returned nil error")
	}

	if recorder.Code != int(types.ErrConflict) {
		t.Fatalf("status = %d, want %d", recorder.Code, types.ErrConflict)
	}

	expected := []string{
		"ParseID",
		"ApplyCustomFilter",
		"ApplyPathConstraints",
		"ApplyDeletedFilter:true",
		"RetrieveElement:42",
		"Restore:100",
	}
	if !reflect.DeepEqual(engine.calls, expected) {
		t.Fatalf("calls = %#v, want %#v", engine.calls, expected)
	}
}
