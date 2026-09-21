package singleton

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/universe-10th/echo-resources/types"
)

type singletonRestoreTestEngine struct {
	calls      []string
	restoreErr error
}

func (e *singletonRestoreTestEngine) ApplyDeletedFilter(filter *types.FilterExpression, deleted bool) {
	if deleted {
		e.calls = append(e.calls, "ApplyDeletedFilter:true")
		return
	}
	e.calls = append(e.calls, "ApplyDeletedFilter:false")
}

func (e *singletonRestoreTestEngine) ApplyPathConstraints(context echo.Context, filter *types.FilterExpression) error {
	e.calls = append(e.calls, "ApplyPathConstraints")
	return nil
}

func (e *singletonRestoreTestEngine) ApplyCustomFilter(context echo.Context, filter *types.FilterExpression) error {
	e.calls = append(e.calls, "ApplyCustomFilter")
	return nil
}

func (e *singletonRestoreTestEngine) RetrieveElement(
	filter *types.FilterExpression,
) (singletonTestResource, bool, error) {
	e.calls = append(e.calls, "RetrieveElement")
	return singletonTestResource{id: 42}, true, nil
}

func (e *singletonRestoreTestEngine) Restore(id int) (singletonTestResource, error) {
	e.calls = append(e.calls, "Restore:"+strconv.Itoa(id))
	return singletonTestResource{id: 84}, e.restoreErr
}

func (e *singletonRestoreTestEngine) RenderElement(context echo.Context, element singletonTestResource) error {
	e.calls = append(e.calls, "RenderElement:"+strconv.Itoa(element.id))
	return nil
}

func newSingletonRestoreTestContext() (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/resource/deleted", nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestRestoreEndpointStubFollowsExpectedOrder(t *testing.T) {
	engine := &singletonRestoreTestEngine{}
	stub := NewRestoreEndpointStub[int, singletonTestResource](engine)
	context, _ := newSingletonRestoreTestContext()

	if err := stub.Restore(context); err != nil {
		t.Fatalf("Restore returned error: %v", err)
	}

	expected := []string{
		"ApplyCustomFilter",
		"ApplyPathConstraints",
		"ApplyDeletedFilter:true",
		"RetrieveElement",
		"Restore:42",
		"RenderElement:84",
	}
	if !reflect.DeepEqual(engine.calls, expected) {
		t.Fatalf("calls = %#v, want %#v", engine.calls, expected)
	}
}

func TestRestoreEndpointStubRestoreErrorReturnsTypedError(t *testing.T) {
	engine := &singletonRestoreTestEngine{
		restoreErr: types.NotDeletedError{},
	}
	stub := NewRestoreEndpointStub[int, singletonTestResource](engine)
	context, recorder := newSingletonRestoreTestContext()

	if err := stub.Restore(context); err == nil {
		t.Fatal("Restore returned nil error")
	}

	if recorder.Code != int(types.ErrConflict) {
		t.Fatalf("status = %d, want %d", recorder.Code, types.ErrConflict)
	}

	expected := []string{
		"ApplyCustomFilter",
		"ApplyPathConstraints",
		"ApplyDeletedFilter:true",
		"RetrieveElement",
		"Restore:42",
	}
	if !reflect.DeepEqual(engine.calls, expected) {
		t.Fatalf("calls = %#v, want %#v", engine.calls, expected)
	}
}
