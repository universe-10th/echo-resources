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

type createTestEngine struct {
	calls         []string
	validationErr error
	saveErr       error
}

func (e *createTestEngine) ReadBody(context echo.Context, element *getTestResource) error {
	e.calls = append(e.calls, "ReadBody")
	element.id = 100
	return nil
}

func (e *createTestEngine) Validate(element *getTestResource) error {
	e.calls = append(e.calls, "Validate:"+strconv.Itoa(element.id))
	return e.validationErr
}

func (e *createTestEngine) ApplyPathConstraintsToElement(context echo.Context, element *getTestResource) error {
	e.calls = append(e.calls, "ApplyPathConstraintsToElement:"+strconv.Itoa(element.id))
	element.id = 42
	return nil
}

func (e *createTestEngine) Save(element *getTestResource) error {
	e.calls = append(e.calls, "Save:"+strconv.Itoa(element.id))
	return e.saveErr
}

func (e *createTestEngine) RenderElement(context echo.Context, element getTestResource) error {
	e.calls = append(e.calls, "RenderElement:"+strconv.Itoa(element.id))
	return nil
}

func newCreateTestContext() (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/resources", nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestCreateEndpointStubFollowsExpectedOrder(t *testing.T) {
	engine := &createTestEngine{}
	stub := CreateEndpointStub[int, getTestResource]{engine: engine}
	context, _ := newCreateTestContext()

	if err := stub.Create(context); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	expected := []string{
		"ReadBody",
		"ApplyPathConstraintsToElement:100",
		"Validate:42",
		"Save:42",
		"RenderElement:42",
	}
	if !reflect.DeepEqual(engine.calls, expected) {
		t.Fatalf("calls = %#v, want %#v", engine.calls, expected)
	}
}

func TestCreateEndpointStubValidationErrorReturnsInvalid(t *testing.T) {
	engine := &createTestEngine{
		validationErr: types.ValidationError{Errors: map[string]any{"name": "required"}},
	}
	stub := CreateEndpointStub[int, getTestResource]{engine: engine}
	context, recorder := newCreateTestContext()

	if err := stub.Create(context); err == nil {
		t.Fatal("Create returned nil error")
	}

	if recorder.Code != int(types.ErrInvalid) {
		t.Fatalf("status = %d, want %d", recorder.Code, types.ErrInvalid)
	}

	expected := []string{
		"ReadBody",
		"ApplyPathConstraintsToElement:100",
		"Validate:42",
	}
	if !reflect.DeepEqual(engine.calls, expected) {
		t.Fatalf("calls = %#v, want %#v", engine.calls, expected)
	}
}

func TestCreateEndpointStubSaveErrorReturnsTypedError(t *testing.T) {
	engine := &createTestEngine{
		saveErr: types.AlreadyUsedError{
			KeyFields: []any{"id"},
			Values:    []any{42},
		},
	}
	stub := CreateEndpointStub[int, getTestResource]{engine: engine}
	context, recorder := newCreateTestContext()

	if err := stub.Create(context); err == nil {
		t.Fatal("Create returned nil error")
	}

	if recorder.Code != int(types.ErrConflict) {
		t.Fatalf("status = %d, want %d", recorder.Code, types.ErrConflict)
	}

	expected := []string{
		"ReadBody",
		"ApplyPathConstraintsToElement:100",
		"Validate:42",
		"Save:42",
	}
	if !reflect.DeepEqual(engine.calls, expected) {
		t.Fatalf("calls = %#v, want %#v", engine.calls, expected)
	}
}
