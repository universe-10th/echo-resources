package echo

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	echov4 "github.com/labstack/echo/v4"
)

type captureResource struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type captureInput struct {
	Name string `json:"name"`
}

func (r captureResource) GetID() int {
	return r.ID
}

func (r captureResource) SetID(id int) {}

func (r captureResource) GetIDField() string {
	return "id"
}

func (r captureResource) GetCreationTime() time.Time {
	return r.CreatedAt
}

func (r captureResource) GetLastUpdateTime() time.Time {
	return r.UpdatedAt
}

func (r captureResource) SetCreationTime() {}

func (r captureResource) SetCreationTimeIn(location *time.Location) {}

func (r captureResource) SetLastUpdateTime() {}

func (r captureResource) SetLastUpdateTimeIn(location *time.Location) {}

func (r captureResource) GetCreationTimeField() string {
	return "created_at"
}

func (r captureResource) GetLastUpdateTimeField() string {
	return "updated_at"
}

func TestMakeBodyCapturerBindsJSONBody(t *testing.T) {
	t.Parallel()

	context, recorder := newCaptureContext(http.MethodPost, `{"id": 10, "name": "Ada"}`, echov4.MIMEApplicationJSON)
	capturer := MakeBodyCapturer[int, captureResource]()

	got, err := capturer(context)
	if err != nil {
		t.Fatalf("expected body capture to succeed, got error: %v", err)
	}
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected no error response to be written, got status %d", recorder.Code)
	}
	if got.ID != 10 || got.Name != "Ada" {
		t.Fatalf("unexpected captured resource: %#v", got)
	}
}

func TestMakeBodyCapturerRejectsNonJSONContentType(t *testing.T) {
	t.Parallel()

	context, recorder := newCaptureContext(http.MethodPost, `{"id": 10}`, echov4.MIMETextPlain)
	capturer := MakeBodyCapturer[int, captureResource]()

	if _, err := capturer(context); err == nil {
		t.Fatal("expected non-JSON request to fail")
	}
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request response, got status %d", recorder.Code)
	}
}

func TestMakeBodyCapturerRejectsInvalidJSONBody(t *testing.T) {
	t.Parallel()

	context, recorder := newCaptureContext(http.MethodPost, `{"id":`, echov4.MIMEApplicationJSON)
	capturer := MakeBodyCapturer[int, captureResource]()

	if _, err := capturer(context); err == nil {
		t.Fatal("expected invalid JSON request to fail")
	}
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request response, got status %d", recorder.Code)
	}
}

func TestMakeMappedBodyCapturerBindsAndMapsJSONBody(t *testing.T) {
	t.Parallel()

	context, recorder := newCaptureContext(http.MethodPost, `{"name": "Ada"}`, echov4.MIMEApplicationJSON)
	capturer := MakeMappedBodyCapturer[int, captureResource](func(input *captureInput) captureResource {
		return captureResource{ID: 20, Name: input.Name}
	})

	got, err := capturer(context)
	if err != nil {
		t.Fatalf("expected mapped body capture to succeed, got error: %v", err)
	}
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected no error response to be written, got status %d", recorder.Code)
	}
	if got.ID != 20 || got.Name != "Ada" {
		t.Fatalf("unexpected mapped resource: %#v", got)
	}
}

func TestMakeMappedBodyCapturerRejectsInvalidJSONBody(t *testing.T) {
	t.Parallel()

	context, recorder := newCaptureContext(http.MethodPost, `{"name":`, echov4.MIMEApplicationJSON)
	capturer := MakeMappedBodyCapturer[int, captureResource](func(input *captureInput) captureResource {
		return captureResource{Name: input.Name}
	})

	if _, err := capturer(context); err == nil {
		t.Fatal("expected invalid JSON request to fail")
	}
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request response, got status %d", recorder.Code)
	}
}

func newCaptureContext(method, body, contentType string) (echov4.Context, *httptest.ResponseRecorder) {
	e := echov4.New()
	request := httptest.NewRequest(method, "/", strings.NewReader(body))
	request.Header.Set(echov4.HeaderContentType, contentType)
	recorder := httptest.NewRecorder()

	return e.NewContext(request, recorder), recorder
}
