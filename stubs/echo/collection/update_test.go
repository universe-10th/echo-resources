package collection

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

type updateTestEngine struct {
	calls         []string
	validationErr error
}

func (e *updateTestEngine) URLArg() string {
	return "id"
}

func (e *updateTestEngine) ParseID(context echo.Context) (int, error) {
	e.calls = append(e.calls, "ParseID")
	return 42, nil
}

func (e *updateTestEngine) ApplyPathConstraints(context echo.Context, filter *types.FilterExpression) error {
	e.calls = append(e.calls, "ApplyPathConstraints")
	return nil
}

func (e *updateTestEngine) ApplyDeletedFilter(filter *types.FilterExpression, deleted bool) {
	if deleted {
		e.calls = append(e.calls, "ApplyDeletedFilter:true")
		return
	}
	e.calls = append(e.calls, "ApplyDeletedFilter:false")
}

func (e *updateTestEngine) ApplyCustomFilter(context echo.Context, filter *types.FilterExpression) error {
	e.calls = append(e.calls, "ApplyCustomFilter")
	return nil
}

func (e *updateTestEngine) RetrieveElement(id int, filter *types.FilterExpression) (getTestResource, bool, error) {
	e.calls = append(e.calls, "RetrieveElement")
	return getTestResource{id: id}, true, nil
}

func (e *updateTestEngine) PreserveStampsAndConstraints(
	context echo.Context, element *getTestResource,
) (time.Time, any) {
	e.calls = append(e.calls, "PreserveStampsAndConstraints:"+strconv.Itoa(element.id))
	return time.Time{}, element.id
}

func (e *updateTestEngine) ReadBody(context echo.Context, element *getTestResource) error {
	e.calls = append(e.calls, "ReadBody")
	element.id = 100
	return nil
}

func (e *updateTestEngine) Validate(element *getTestResource) error {
	e.calls = append(e.calls, "Validate:"+strconv.Itoa(element.id))
	return e.validationErr
}

func (e *updateTestEngine) RestoreIDStampsAndConstraints(
	element *getTestResource, id int, createdAt time.Time, constraints any,
) {
	e.calls = append(e.calls, "RestoreIDStampsAndConstraints")
	element.id = id
}

func (e *updateTestEngine) ApplyPathConstraintsToElement(context echo.Context, element *getTestResource) error {
	e.calls = append(e.calls, "ApplyPathConstraintsToElement:"+strconv.Itoa(element.id))
	element.id = 84
	return nil
}

func (e *updateTestEngine) Save(element *getTestResource) error {
	e.calls = append(e.calls, "Save:"+strconv.Itoa(element.id))
	return nil
}

func (e *updateTestEngine) RenderElement(context echo.Context, element getTestResource) error {
	e.calls = append(e.calls, "RenderElement:"+strconv.Itoa(element.id))
	return nil
}

func newUpdateTestContext() (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/resources/42", nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestUpdateEndpointStubFollowsExpectedOrder(t *testing.T) {
	engine := &updateTestEngine{}
	stub := UpdateEndpointStub[int, getTestResource]{engine: engine}
	context, _ := newUpdateTestContext()

	if err := stub.Update(context); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	expected := []string{
		"ParseID",
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
		"RenderElement:84",
	}
	if !reflect.DeepEqual(engine.calls, expected) {
		t.Fatalf("calls = %#v, want %#v", engine.calls, expected)
	}
}

func TestUpdateEndpointStubValidationErrorReturnsInvalid(t *testing.T) {
	engine := &updateTestEngine{
		validationErr: types.ValidationError{Errors: map[string]any{"name": "required"}},
	}
	stub := UpdateEndpointStub[int, getTestResource]{engine: engine}
	context, recorder := newUpdateTestContext()

	if err := stub.Update(context); err == nil {
		t.Fatal("Update returned nil error")
	}

	if recorder.Code != int(types.ErrInvalid) {
		t.Fatalf("status = %d, want %d", recorder.Code, types.ErrInvalid)
	}

	expected := []string{
		"ParseID",
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
