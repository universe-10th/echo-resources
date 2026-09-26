package echo

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	echov4 "github.com/labstack/echo/v4"
	"github.com/universe-10th/echo-resources/types/services"
	"github.com/universe-10th/echo-resources/utils"
)

type testService struct {
	prefix      string
	urlArg      string
	singleton   bool
	softDeleted bool
	verbs       utils.Flags[services.ResourceVerb]
	parent      services.Service
	children    []services.Service
	middlewares []services.MiddlewareFunc
}

func (service *testService) Prefix() string                            { return service.prefix }
func (service *testService) URLArg() string                            { return service.urlArg }
func (service *testService) IsSingleton() bool                         { return service.singleton }
func (service *testService) Verbs() utils.Flags[services.ResourceVerb] { return service.verbs }
func (service *testService) CanHaveChildren() bool                     { return service.Verbs().Has(services.ResourceGet) }
func (service *testService) IsSoftDeleted() bool                       { return service.softDeleted }
func (service *testService) Children() []services.Service              { return service.children }
func (service *testService) Parent() services.Service                  { return service.parent }
func (service *testService) Middlewares() []services.MiddlewareFunc {
	return service.middlewares
}
func (service *testService) ElementMiddleware(bool) services.MiddlewareFunc {
	return func(next services.HandlerFunc) services.HandlerFunc {
		return func(context services.Context) error {
			context.SetData("element", true)
			context.PushElement("element")
			defer context.PopElement()
			return next(context)
		}
	}
}
func (service *testService) List(context services.Context, deleted bool) error {
	endpointType, verb, name := context.CurrentEndpoint()
	middlewareValue, _ := context.GetData("middleware")
	return context.RenderJSON(http.StatusOK, map[string]any{
		"deleted":     deleted,
		"endpoint":    endpointType,
		"verb":        verb,
		"name":        name,
		"middleware":  middlewareValue,
		"native_echo": context.Native() != nil,
	})
}
func (service *testService) Create(context services.Context) error {
	return context.RenderJSON(http.StatusCreated, map[string]any{"created": true})
}
func (service *testService) Get(context services.Context) error {
	_, hasElement := context.PeekElement(0)
	return context.RenderJSON(http.StatusOK, map[string]any{"element": hasElement})
}
func (service *testService) Update(context services.Context) error {
	return context.RenderNoContent(http.StatusNoContent)
}
func (service *testService) Delete(context services.Context) error {
	return context.RenderNoContent(http.StatusNoContent)
}
func (service *testService) Prune(context services.Context) error {
	return context.RenderNoContent(http.StatusNoContent)
}
func (service *testService) Restore(context services.Context) error {
	return context.RenderJSON(http.StatusOK, map[string]any{"restored": true})
}

func TestInstallRejectsInvalidRootInputs(t *testing.T) {
	t.Parallel()

	if err := Install(nil, &testService{}); !errors.Is(err, ErrInvalidEchoApp) {
		t.Fatalf("expected ErrInvalidEchoApp, got %v", err)
	}

	app := echov4.New()
	if err := Install(app, nil); !errors.Is(err, ErrInvalidService) {
		t.Fatalf("expected ErrInvalidService, got %v", err)
	}

	parent := &testService{prefix: "parents"}
	child := &testService{prefix: "children", parent: parent}
	if err := Install(app, child); !errors.Is(err, ErrInvalidRootService) {
		t.Fatalf("expected ErrInvalidRootService, got %v", err)
	}
}

func TestInstallWrapsMiddlewareAndHandlerContext(t *testing.T) {
	t.Parallel()

	app := echov4.New()
	service := &testService{
		prefix: "items",
		urlArg: "item_id",
		verbs:  utils.NewFlags(services.ResourceList),
		middlewares: []services.MiddlewareFunc{
			func(next services.HandlerFunc) services.HandlerFunc {
				return func(context services.Context) error {
					context.SetData("middleware", "seen")
					return next(context)
				}
			},
		},
	}

	if err := Install(app, service); err != nil {
		t.Fatalf("Install returned error: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/items", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d with body %s", response.Code, response.Body.String())
	}
	if body := response.Body.String(); body == "" || !containsAll(body, `"middleware":"seen"`, `"verb":1`, `"native_echo":true`) {
		t.Fatalf("unexpected response body: %s", body)
	}
}

func TestInstallWrapsElementMiddleware(t *testing.T) {
	t.Parallel()

	app := echov4.New()
	service := &testService{
		prefix: "items",
		urlArg: "item_id",
		verbs:  utils.NewFlags(services.ResourceGet),
	}

	if err := Install(app, service); err != nil {
		t.Fatalf("Install returned error: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/items/42", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d with body %s", response.Code, response.Body.String())
	}
	if body := response.Body.String(); body == "" || !containsAll(body, `"element":true`) {
		t.Fatalf("unexpected response body: %s", body)
	}
}

func containsAll(value string, fragments ...string) bool {
	for _, fragment := range fragments {
		if !contains(value, fragment) {
			return false
		}
	}
	return true
}

func contains(value string, fragment string) bool {
	for start := 0; start+len(fragment) <= len(value); start++ {
		if value[start:start+len(fragment)] == fragment {
			return true
		}
	}
	return false
}
