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

type singletonUpdateTestEngine struct {
	calls         []string
	validationErr error
}

func (e *singletonUpdateTestEngine) ApplyDeletedFilter(filter *types.FilterExpression, deleted bool) {
	if deleted {
		e.calls = append(e.calls, "ApplyDeletedFilter:true")
		return
	}
	e.calls = append(e.calls, "ApplyDeletedFilter:false")
}

func (e *singletonUpdateTestEngine) ApplyPathConstraintsToFilter(context echo.Context, filter *types.FilterExpression) error {
	e.calls = append(e.calls, "ApplyPathConstraints")
	return nil
}

func (e *singletonUpdateTestEngine) ApplyCustomFilter(context echo.Context, filter *types.FilterExpression) error {
	e.calls = append(e.calls, "ApplyCustomFilter")
	return nil
}

func (e *singletonUpdateTestEngine) RetrieveElement(
	filter *types.FilterExpression,
) (singletonTestResource, bool, error) {
	e.calls = append(e.calls, "RetrieveElement")
	return singletonTestResource{id: 42}, true, nil
}

func (e *singletonUpdateTestEngine) PreserveStampsAndConstraints(
	context echo.Context, element *singletonTestResource,
) (time.Time, any) {
	e.calls = append(e.calls, "PreserveStampsAndConstraints:"+strconv.Itoa(element.id))
	return time.Time{}, element.id
}

func (e *singletonUpdateTestEngine) ReadBody(context echo.Context, element *singletonTestResource) error {
	e.calls = append(e.calls, "ReadBody")
	element.id = 100
	return nil
}

func (e *singletonUpdateTestEngine) RestoreIDStampsAndConstraints(
	element *singletonTestResource, id int, createdAt time.Time, constraints any,
) {
	e.calls = append(e.calls, "RestoreIDStampsAndConstraints")
	element.id = id
}

func (e *singletonUpdateTestEngine) ApplyPathConstraintsToElement(
	context echo.Context, element *singletonTestResource,
) error {
	e.calls = append(e.calls, "ApplyPathConstraintsToElement:"+strconv.Itoa(element.id))
	element.id = 84
	return nil
}

func (e *singletonUpdateTestEngine) Validate(element *singletonTestResource) error {
	e.calls = append(e.calls, "Validate:"+strconv.Itoa(element.id))
	return e.validationErr
}

func (e *singletonUpdateTestEngine) Save(element *singletonTestResource) error {
	e.calls = append(e.calls, "Save:"+strconv.Itoa(element.id))
	return nil
}

func (e *singletonUpdateTestEngine) RenderElement(
	context echo.Context, element singletonTestResource, created bool,
) error {
	e.calls = append(e.calls, "RenderElement:"+strconv.Itoa(element.id)+":"+strconv.FormatBool(created))
	return nil
}

func newSingletonUpdateTestContext() (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/resource", nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestUpdateEndpointStubFollowsExpectedOrder(t *testing.T) {
	engine := &singletonUpdateTestEngine{}
	stub := NewUpdateEndpointStub[int, singletonTestResource](engine)
	context, _ := newSingletonUpdateTestContext()

	if err := stub.Update(context); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	expected := []string{
		"ApplyCustomFilter",
		"ApplyPathConstraints",
		"ApplyDeletedFilter:false",
		"RetrieveElement",
		"PreserveStampsAndConstraints:42",
		"ReadBody",
		"RestoreIDStampsAndConstraints",
		"ApplyPathConstraintsToElement:42",
		"Validate:84",
		"Save:84",
		"RenderElement:84:false",
	}
	if !reflect.DeepEqual(engine.calls, expected) {
		t.Fatalf("calls = %#v, want %#v", engine.calls, expected)
	}
}

func TestUpdateEndpointStubValidationErrorReturnsInvalid(t *testing.T) {
	engine := &singletonUpdateTestEngine{
		validationErr: types.ValidationError{Errors: map[string]any{"name": "required"}},
	}
	stub := NewUpdateEndpointStub[int, singletonTestResource](engine)
	context, recorder := newSingletonUpdateTestContext()

	if err := stub.Update(context); err == nil {
		t.Fatal("Update returned nil error")
	}

	if recorder.Code != int(types.ErrInvalid) {
		t.Fatalf("status = %d, want %d", recorder.Code, types.ErrInvalid)
	}

	expected := []string{
		"ApplyCustomFilter",
		"ApplyPathConstraints",
		"ApplyDeletedFilter:false",
		"RetrieveElement",
		"PreserveStampsAndConstraints:42",
		"ReadBody",
		"RestoreIDStampsAndConstraints",
		"ApplyPathConstraintsToElement:42",
		"Validate:84",
	}
	if !reflect.DeepEqual(engine.calls, expected) {
		t.Fatalf("calls = %#v, want %#v", engine.calls, expected)
	}
}
