package echo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	echov4 "github.com/labstack/echo/v4"
)

type renderedCaptureResource struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestMakeElementRendererRendersResource(t *testing.T) {
	t.Parallel()

	context, recorder := newRendererContext()
	renderer := MakeElementRenderer[int, captureResource]()

	err := renderer(context, http.StatusCreated, captureResource{ID: 10, Name: "Ada"})
	if err != nil {
		t.Fatalf("expected render to succeed, got error: %v", err)
	}
	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, recorder.Code)
	}

	var got renderedCaptureResource
	decodeResponse(t, recorder, &got)
	if got.ID != 10 || got.Name != "Ada" {
		t.Fatalf("unexpected rendered resource: %#v", got)
	}
}

func TestMakeMappedElementRendererRendersProjection(t *testing.T) {
	t.Parallel()

	context, recorder := newRendererContext()
	renderer := MakeMappedElementRenderer[int, captureResource](func(resource *captureResource) renderedCaptureResource {
		return renderedCaptureResource{ID: resource.ID, Name: resource.Name}
	})

	err := renderer(context, http.StatusOK, &captureResource{ID: 20, Name: "Grace"})
	if err != nil {
		t.Fatalf("expected render to succeed, got error: %v", err)
	}

	var got renderedCaptureResource
	decodeResponse(t, recorder, &got)
	if got.ID != 20 || got.Name != "Grace" {
		t.Fatalf("unexpected rendered projection: %#v", got)
	}
}

func TestMakeCollectionRendererRendersItemsEnvelope(t *testing.T) {
	t.Parallel()

	context, recorder := newRendererContext()
	renderer := MakeCollectionRenderer[int, captureResource]()

	err := renderer(context, http.StatusOK, []captureResource{
		{ID: 10, Name: "Ada"},
		{ID: 20, Name: "Grace"},
	}, 5, 25)
	if err != nil {
		t.Fatalf("expected render to succeed, got error: %v", err)
	}

	var got struct {
		Skip  int64                     `json:"skip"`
		Count int                       `json:"count"`
		Total int64                     `json:"total"`
		Items []renderedCaptureResource `json:"items"`
	}
	decodeResponse(t, recorder, &got)
	if got.Skip != 5 || got.Count != 2 || got.Total != 25 {
		t.Fatalf("unexpected collection envelope: %#v", got)
	}
	if len(got.Items) != 2 || got.Items[0].Name != "Ada" || got.Items[1].Name != "Grace" {
		t.Fatalf("unexpected collection items: %#v", got.Items)
	}
}

func TestMakeMappedCollectionRendererRendersProjectedItems(t *testing.T) {
	t.Parallel()

	context, recorder := newRendererContext()
	renderer := MakeMappedCollectionRenderer[int, captureResource](func(resource *captureResource) renderedCaptureResource {
		return renderedCaptureResource{ID: resource.ID, Name: resource.Name}
	})

	err := renderer(context, http.StatusOK, []*captureResource{
		{ID: 10, Name: "Ada"},
		{ID: 20, Name: "Grace"},
	}, 0, 2)
	if err != nil {
		t.Fatalf("expected render to succeed, got error: %v", err)
	}

	var got struct {
		Skip  int64                     `json:"skip"`
		Count int                       `json:"count"`
		Total int64                     `json:"total"`
		Items []renderedCaptureResource `json:"items"`
	}
	decodeResponse(t, recorder, &got)
	if got.Count != 2 || len(got.Items) != 2 {
		t.Fatalf("unexpected mapped collection response: %#v", got)
	}
	if got.Items[0].Name != "Ada" || got.Items[1].Name != "Grace" {
		t.Fatalf("unexpected mapped collection items: %#v", got.Items)
	}
}

func TestMakeElementRendererRejectsUnsupportedObject(t *testing.T) {
	t.Parallel()

	context, recorder := newRendererContext()
	renderer := MakeElementRenderer[int, captureResource]()

	err := renderer(context, http.StatusOK, "not a resource")
	if err != nil {
		t.Fatalf("expected error response render to succeed, got error: %v", err)
	}
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected internal error status, got %d", recorder.Code)
	}

	var got map[string]any
	decodeResponse(t, recorder, &got)
	if got["detail"] != "internal error" {
		t.Fatalf("unexpected error response: %#v", got)
	}
}

func newRendererContext() (echov4.Context, *httptest.ResponseRecorder) {
	e := echov4.New()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()

	return e.NewContext(request, recorder), recorder
}

func decodeResponse(t *testing.T, recorder *httptest.ResponseRecorder, target any) {
	t.Helper()

	if err := json.Unmarshal(recorder.Body.Bytes(), target); err != nil {
		t.Fatalf("expected JSON response, got error: %v", err)
	}
}
