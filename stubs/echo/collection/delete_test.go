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

type deleteTestEngine struct {
	calls     []string
	deleteErr error
}

func (e *deleteTestEngine) ParseID(context echo.Context) (int, error) {
	e.calls = append(e.calls, "ParseID")
	return 42, nil
}

func (e *deleteTestEngine) ApplyPathConstraints(context echo.Context, filter *types.FilterExpression) error {
	e.calls = append(e.calls, "ApplyPathConstraints")
	return nil
}

func (e *deleteTestEngine) ApplyDeletedFilter(filter *types.FilterExpression, deleted bool) {
	if deleted {
		e.calls = append(e.calls, "ApplyDeletedFilter:true")
		return
	}
	e.calls = append(e.calls, "ApplyDeletedFilter:false")
}

func (e *deleteTestEngine) ApplyCustomFilter(context echo.Context, filter *types.FilterExpression) error {
	e.calls = append(e.calls, "ApplyCustomFilter")
	return nil
}

func (e *deleteTestEngine) RetrieveElement(id int, filter *types.FilterExpression) (getTestResource, bool, error) {
	e.calls = append(e.calls, "RetrieveElement:"+strconv.Itoa(id))
	return getTestResource{id: 100}, true, nil
}

func (e *deleteTestEngine) Delete(id int) error {
	e.calls = append(e.calls, "Delete:"+strconv.Itoa(id))
	return e.deleteErr
}

func (e *deleteTestEngine) RenderEmpty(context echo.Context) error {
	e.calls = append(e.calls, "RenderEmpty")
	return nil
}

func newDeleteTestContext() (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/resources/42", nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestDeleteEndpointStubFollowsExpectedOrder(t *testing.T) {
	engine := &deleteTestEngine{}
	stub := DeleteEndpointStub[int, getTestResource]{engine: engine}
	context, _ := newDeleteTestContext()

	if err := stub.Delete(context); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	expected := []string{
		"ParseID",
		"ApplyCustomFilter",
		"ApplyPathConstraints",
		"ApplyDeletedFilter:false",
		"RetrieveElement:42",
		"Delete:100",
		"RenderEmpty",
	}
	if !reflect.DeepEqual(engine.calls, expected) {
		t.Fatalf("calls = %#v, want %#v", engine.calls, expected)
	}
}

func TestDeleteEndpointStubDeleteErrorReturnsTypedError(t *testing.T) {
	engine := &deleteTestEngine{
		deleteErr: types.AlreadyUsedError{
			KeyFields: []any{"id"},
			Values:    []any{100},
		},
	}
	stub := DeleteEndpointStub[int, getTestResource]{engine: engine}
	context, recorder := newDeleteTestContext()

	if err := stub.Delete(context); err == nil {
		t.Fatal("Delete returned nil error")
	}

	if recorder.Code != int(types.ErrConflict) {
		t.Fatalf("status = %d, want %d", recorder.Code, types.ErrConflict)
	}

	expected := []string{
		"ParseID",
		"ApplyCustomFilter",
		"ApplyPathConstraints",
		"ApplyDeletedFilter:false",
		"RetrieveElement:42",
		"Delete:100",
	}
	if !reflect.DeepEqual(engine.calls, expected) {
		t.Fatalf("calls = %#v, want %#v", engine.calls, expected)
	}
}
