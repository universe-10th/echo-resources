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

type singletonDeleteTestEngine struct {
	calls     []string
	deleteErr error
}

func (e *singletonDeleteTestEngine) ApplyDeletedFilter(filter *types.FilterExpression, deleted bool) {
	if deleted {
		e.calls = append(e.calls, "ApplyDeletedFilter:true")
		return
	}
	e.calls = append(e.calls, "ApplyDeletedFilter:false")
}

func (e *singletonDeleteTestEngine) ApplyPathConstraintsToFilter(context echo.Context, filter *types.FilterExpression) error {
	e.calls = append(e.calls, "ApplyPathConstraints")
	return nil
}

func (e *singletonDeleteTestEngine) ApplyCustomFilter(context echo.Context, filter *types.FilterExpression) error {
	e.calls = append(e.calls, "ApplyCustomFilter")
	return nil
}

func (e *singletonDeleteTestEngine) RetrieveElement(
	filter *types.FilterExpression,
) (singletonTestResource, bool, error) {
	e.calls = append(e.calls, "RetrieveElement")
	return singletonTestResource{id: 42}, true, nil
}

func (e *singletonDeleteTestEngine) Delete(id int) error {
	e.calls = append(e.calls, "Delete:"+strconv.Itoa(id))
	return e.deleteErr
}

func (e *singletonDeleteTestEngine) RenderEmpty(context echo.Context) error {
	e.calls = append(e.calls, "RenderEmpty")
	return nil
}

func newSingletonDeleteTestContext() (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/resource", nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestDeleteEndpointStubFollowsExpectedOrder(t *testing.T) {
	engine := &singletonDeleteTestEngine{}
	stub := NewDeleteEndpointStub[int, singletonTestResource](engine)
	context, _ := newSingletonDeleteTestContext()

	if err := stub.Delete(context); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	expected := []string{
		"ApplyCustomFilter",
		"ApplyPathConstraints",
		"ApplyDeletedFilter:false",
		"RetrieveElement",
		"Delete:42",
		"RenderEmpty",
	}
	if !reflect.DeepEqual(engine.calls, expected) {
		t.Fatalf("calls = %#v, want %#v", engine.calls, expected)
	}
}

func TestDeleteEndpointStubDeleteErrorReturnsTypedError(t *testing.T) {
	engine := &singletonDeleteTestEngine{
		deleteErr: types.AlreadyUsedError{
			KeyFields: []any{"id"},
			Values:    []any{42},
		},
	}
	stub := NewDeleteEndpointStub[int, singletonTestResource](engine)
	context, recorder := newSingletonDeleteTestContext()

	if err := stub.Delete(context); err == nil {
		t.Fatal("Delete returned nil error")
	}

	if recorder.Code != int(types.ErrConflict) {
		t.Fatalf("status = %d, want %d", recorder.Code, types.ErrConflict)
	}

	expected := []string{
		"ApplyCustomFilter",
		"ApplyPathConstraints",
		"ApplyDeletedFilter:false",
		"RetrieveElement",
		"Delete:42",
	}
	if !reflect.DeepEqual(engine.calls, expected) {
		t.Fatalf("calls = %#v, want %#v", engine.calls, expected)
	}
}
