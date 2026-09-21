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

type singletonPruneTestEngine struct {
	calls    []string
	pruneErr error
}

func (e *singletonPruneTestEngine) ApplyDeletedFilter(filter *types.FilterExpression, deleted bool) {
	if deleted {
		e.calls = append(e.calls, "ApplyDeletedFilter:true")
		return
	}
	e.calls = append(e.calls, "ApplyDeletedFilter:false")
}

func (e *singletonPruneTestEngine) ApplyPathConstraints(context echo.Context, filter *types.FilterExpression) error {
	e.calls = append(e.calls, "ApplyPathConstraints")
	return nil
}

func (e *singletonPruneTestEngine) ApplyCustomFilter(context echo.Context, filter *types.FilterExpression) error {
	e.calls = append(e.calls, "ApplyCustomFilter")
	return nil
}

func (e *singletonPruneTestEngine) RetrieveElement(
	filter *types.FilterExpression,
) (singletonTestResource, bool, error) {
	e.calls = append(e.calls, "RetrieveElement")
	return singletonTestResource{id: 42}, true, nil
}

func (e *singletonPruneTestEngine) Prune(id int) error {
	e.calls = append(e.calls, "Prune:"+strconv.Itoa(id))
	return e.pruneErr
}

func (e *singletonPruneTestEngine) RenderEmpty(context echo.Context) error {
	e.calls = append(e.calls, "RenderEmpty")
	return nil
}

func newSingletonPruneTestContext() (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/resource/deleted", nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestPruneEndpointStubFollowsExpectedOrder(t *testing.T) {
	engine := &singletonPruneTestEngine{}
	stub := NewPruneEndpointStub[int, singletonTestResource](engine)
	context, _ := newSingletonPruneTestContext()

	if err := stub.Prune(context); err != nil {
		t.Fatalf("Prune returned error: %v", err)
	}

	expected := []string{
		"ApplyCustomFilter",
		"ApplyPathConstraints",
		"ApplyDeletedFilter:true",
		"RetrieveElement",
		"Prune:42",
		"RenderEmpty",
	}
	if !reflect.DeepEqual(engine.calls, expected) {
		t.Fatalf("calls = %#v, want %#v", engine.calls, expected)
	}
}

func TestPruneEndpointStubPruneErrorReturnsTypedError(t *testing.T) {
	engine := &singletonPruneTestEngine{
		pruneErr: types.NotDeletedError{},
	}
	stub := NewPruneEndpointStub[int, singletonTestResource](engine)
	context, recorder := newSingletonPruneTestContext()

	if err := stub.Prune(context); err == nil {
		t.Fatal("Prune returned nil error")
	}

	if recorder.Code != int(types.ErrConflict) {
		t.Fatalf("status = %d, want %d", recorder.Code, types.ErrConflict)
	}

	expected := []string{
		"ApplyCustomFilter",
		"ApplyPathConstraints",
		"ApplyDeletedFilter:true",
		"RetrieveElement",
		"Prune:42",
	}
	if !reflect.DeepEqual(engine.calls, expected) {
		t.Fatalf("calls = %#v, want %#v", engine.calls, expected)
	}
}
