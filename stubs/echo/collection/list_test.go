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

type listTestEngine struct {
	calls []string
}

func (e *listTestEngine) GetAllowedFields() ([]string, Allowance, error) {
	e.calls = append(e.calls, "GetAllowedFields")
	return []string{"id"}, Only, nil
}

func (e *listTestEngine) ApplyPathConstraints(context echo.Context, filter *types.FilterExpression) error {
	e.calls = append(e.calls, "ApplyPathConstraints")
	return nil
}

func (e *listTestEngine) ApplyDeletedFilter(filter *types.FilterExpression, deleted bool) {
	if deleted {
		e.calls = append(e.calls, "ApplyDeletedFilter:true")
		return
	}
	e.calls = append(e.calls, "ApplyDeletedFilter:false")
}

func (e *listTestEngine) ParseFilterAndSort(
	context echo.Context, fields []string, allowance Allowance,
) (*types.FilterExpression, *types.SortExpression, error) {
	e.calls = append(e.calls, "ParseFilterAndSort")
	if !reflect.DeepEqual(fields, []string{"id"}) {
		e.calls = append(e.calls, "unexpected-fields")
	}
	if allowance != Only {
		e.calls = append(e.calls, "unexpected-allowance")
	}
	return &types.FilterExpression{}, &types.SortExpression{}, nil
}

func (e *listTestEngine) ApplyCustomFilterAndSort(
	context echo.Context, filter *types.FilterExpression, sort *types.SortExpression,
) error {
	e.calls = append(e.calls, "ApplyCustomFilterAndSort")
	return nil
}

func (e *listTestEngine) PageSize() int64 {
	e.calls = append(e.calls, "PageSize")
	return 10
}

func (e *listTestEngine) RetrieveList(
	filter *types.FilterExpression, sort *types.SortExpression, page int64,
) ([]getTestResource, int64, error) {
	e.calls = append(e.calls, "RetrieveList:"+strconv.FormatInt(page, 10))
	return []getTestResource{{id: 1}, {id: 2}}, 7, nil
}

func (e *listTestEngine) RenderList(
	context echo.Context, elements []getTestResource, page int64, totalPages int64,
) error {
	e.calls = append(e.calls, "RenderList:"+strconv.FormatInt(page, 10)+":"+strconv.FormatInt(totalPages, 10))
	if len(elements) != 2 {
		e.calls = append(e.calls, "unexpected-elements")
	}
	return nil
}

func newListTestContext() echo.Context {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/resources?page=3", nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}

func TestListEndpointStubFollowsExpectedOrder(t *testing.T) {
	engine := &listTestEngine{}
	stub := ListEndpointStub[int, getTestResource]{engine: engine}

	if err := stub.List(newListTestContext()); err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	expected := []string{
		"GetAllowedFields",
		"ParseFilterAndSort",
		"ApplyCustomFilterAndSort",
		"ApplyPathConstraints",
		"ApplyDeletedFilter:false",
		"RetrieveList:3",
		"RenderList:3:7",
	}
	if !reflect.DeepEqual(engine.calls, expected) {
		t.Fatalf("calls = %#v, want %#v", engine.calls, expected)
	}
}

func TestSoftDeletedListEndpointStubFiltersDeletedElements(t *testing.T) {
	engine := &listTestEngine{}
	stub := SoftDeletedListEndpointStub[int, getTestResource]{engine: engine}

	if err := stub.ListDeleted(newListTestContext()); err != nil {
		t.Fatalf("ListDeleted returned error: %v", err)
	}

	expected := []string{
		"GetAllowedFields",
		"ParseFilterAndSort",
		"ApplyCustomFilterAndSort",
		"ApplyPathConstraints",
		"ApplyDeletedFilter:true",
		"RetrieveList:3",
		"RenderList:3:7",
	}
	if !reflect.DeepEqual(engine.calls, expected) {
		t.Fatalf("calls = %#v, want %#v", engine.calls, expected)
	}
}
