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

type pruneTestEngine struct {
	calls    []string
	pruneErr error
}

func (e *pruneTestEngine) URLArg() string {
	return "id"
}

func (e *pruneTestEngine) ParseID(context echo.Context) (int, error) {
	e.calls = append(e.calls, "ParseID")
	return 42, nil
}

func (e *pruneTestEngine) ApplyPathConstraints(context echo.Context, filter *types.FilterExpression) error {
	e.calls = append(e.calls, "ApplyPathConstraints")
	return nil
}

func (e *pruneTestEngine) ApplyDeletedFilter(filter *types.FilterExpression, deleted bool) {
	if deleted {
		e.calls = append(e.calls, "ApplyDeletedFilter:true")
		return
	}
	e.calls = append(e.calls, "ApplyDeletedFilter:false")
}

func (e *pruneTestEngine) ApplyCustomFilter(context echo.Context, filter *types.FilterExpression) error {
	e.calls = append(e.calls, "ApplyCustomFilter")
	return nil
}

func (e *pruneTestEngine) RetrieveElement(id int, filter *types.FilterExpression) (getTestResource, bool, error) {
	e.calls = append(e.calls, "RetrieveElement:"+strconv.Itoa(id))
	return getTestResource{id: 100}, true, nil
}

func (e *pruneTestEngine) Prune(id int) error {
	e.calls = append(e.calls, "Prune:"+strconv.Itoa(id))
	return e.pruneErr
}

func (e *pruneTestEngine) RenderEmpty(context echo.Context) error {
	e.calls = append(e.calls, "RenderEmpty")
	return nil
}

func newPruneTestContext() (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/resources/42/prune", nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestPruneEndpointStubFollowsExpectedOrder(t *testing.T) {
	engine := &pruneTestEngine{}
	stub := PruneEndpointStub[int, getTestResource]{engine: engine}
	context, _ := newPruneTestContext()

	if err := stub.Prune(context); err != nil {
		t.Fatalf("Prune returned error: %v", err)
	}

	expected := []string{
		"ParseID",
		"ApplyCustomFilter",
		"ApplyPathConstraints",
		"ApplyDeletedFilter:true",
		"RetrieveElement:42",
		"Prune:100",
		"RenderEmpty",
	}
	if !reflect.DeepEqual(engine.calls, expected) {
		t.Fatalf("calls = %#v, want %#v", engine.calls, expected)
	}
}

func TestPruneEndpointStubPruneErrorReturnsTypedError(t *testing.T) {
	engine := &pruneTestEngine{
		pruneErr: types.NotDeletedError{},
	}
	stub := PruneEndpointStub[int, getTestResource]{engine: engine}
	context, recorder := newPruneTestContext()

	if err := stub.Prune(context); err == nil {
		t.Fatal("Prune returned nil error")
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
		"Prune:100",
	}
	if !reflect.DeepEqual(engine.calls, expected) {
		t.Fatalf("calls = %#v, want %#v", engine.calls, expected)
	}
}
