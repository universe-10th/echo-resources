package singleton

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/universe-10th/echo-resources/types"
)

type singletonGetTestEngine struct {
	calls []string
}

func (e *singletonGetTestEngine) ApplyDeletedFilter(filter *types.FilterExpression, deleted bool) {
	if deleted {
		e.calls = append(e.calls, "ApplyDeletedFilter:true")
		return
	}
	e.calls = append(e.calls, "ApplyDeletedFilter:false")
}

func (e *singletonGetTestEngine) ApplyPathConstraints(context echo.Context, filter *types.FilterExpression) error {
	e.calls = append(e.calls, "ApplyPathConstraints")
	return nil
}

func (e *singletonGetTestEngine) ApplyCustomFilter(context echo.Context, filter *types.FilterExpression) error {
	e.calls = append(e.calls, "ApplyCustomFilter")
	return nil
}

func (e *singletonGetTestEngine) RetrieveElement(
	filter *types.FilterExpression,
) (singletonTestResource, bool, error) {
	e.calls = append(e.calls, "RetrieveElement")
	return singletonTestResource{id: 1}, true, nil
}

func (e *singletonGetTestEngine) RenderElement(context echo.Context, element singletonTestResource) error {
	e.calls = append(e.calls, "RenderElement")
	return nil
}

func newSingletonGetTestContext() echo.Context {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/resource", nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}

func TestGetEndpointStubRetrievesNonDeletedSingleton(t *testing.T) {
	engine := &singletonGetTestEngine{}
	stub := GetEndpointStub[int, singletonTestResource]{engine: engine}

	if err := stub.Get(newSingletonGetTestContext()); err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	expected := []string{
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

func TestSoftDeletedGetEndpointStubRetrievesDeletedSingleton(t *testing.T) {
	engine := &singletonGetTestEngine{}
	stub := SoftDeletedGetEndpointStub[int, singletonTestResource]{engine: engine}

	if err := stub.GetDeleted(newSingletonGetTestContext()); err != nil {
		t.Fatalf("GetDeleted returned error: %v", err)
	}

	expected := []string{
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
