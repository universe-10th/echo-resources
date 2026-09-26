package echo

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	echov4 "github.com/labstack/echo/v4"
	"github.com/universe-10th/echo-resources/memory"
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

type integrationStore struct {
	memory.SoftDeletedResource[int]
	Name string `json:"name"`
}

type integrationCatalog struct {
	memory.Resource[int]
	StoreID int    `json:"store_id"`
	Name    string `json:"name"`
}

type integrationProduct struct {
	memory.SoftDeletedResource[int]
	CatalogID int    `json:"catalog_id"`
	Name      string `json:"name"`
	Rank      int    `json:"rank"`
}

type integrationSetting struct {
	memory.Resource[int]
	Version string `json:"version"`
}

type integrationHardItem struct {
	memory.Resource[int]
	Name string `json:"name"`
}

/*
Route documentation for this platform integration.

The test installs three root services:
  - /stores: soft-deleted collection. Its live element group is /stores/:store_id.
  - /platform: singleton.
  - /hard-items: hard-deleted collection. Its live element group is /hard-items/:hard_item_id.

Then it attaches:
  - /stores/:store_id/catalogs: collection constrained by catalog.store_id == store.id.
  - /stores/:store_id/catalogs/:catalog_id/products: soft-deleted collection constrained by product.catalog_id == catalog.id.

Middleware names below are the service-neutral middleware functions wrapped for Echo:
  - setup(S, V): setupMiddleware(service=S, endpointType=EndpointVerb, verb=V, name="").
  - element(S, deleted): services.ElementMiddleware for service S, loading the current element and pushing it on the context stack.
  - service middlewares: service.Middlewares(); this test leaves them empty for every service.

Created URLs and middleware chains:

  - POST /platform
    setup(platform, ResourceCreate)
    Accessed by this test to create the singleton.

  - GET /platform
    setup(platform, ResourceGet) -> element(platform, false)
    Accessed by this test to read the singleton.

  - PATCH /platform
    setup(platform, ResourceUpdate) -> element(platform, false)

  - DELETE /platform
    setup(platform, ResourceDelete) -> element(platform, false)

  - GET /stores
    setup(stores, ResourceList)

  - POST /stores
    setup(stores, ResourceCreate)
    Accessed by this test to create the parent store.

  - GET /stores/:store_id
    setup(stores, ResourceGet) -> element(stores, false)

  - PATCH /stores/:store_id
    setup(stores, ResourceUpdate) -> element(stores, false)

  - DELETE /stores/:store_id
    setup(stores, ResourceDelete) -> element(stores, false)

  - GET /stores/deleted
    setup(stores, ResourceListDeleted)

  - GET /stores/deleted/:store_id
    setup(stores, ResourceGetDeleted) -> element(stores, true)

  - POST /stores/deleted/:store_id
    setup(stores, ResourceRestore) -> element(stores, true)

  - DELETE /stores/deleted/:store_id
    setup(stores, ResourcePrune) -> element(stores, true)

  - GET /stores/:store_id/catalogs
    setup(stores, ResourceGet) -> element(stores, false) ->
    setup(catalogs, ResourceList)

  - POST /stores/:store_id/catalogs
    setup(stores, ResourceGet) -> element(stores, false) ->
    setup(catalogs, ResourceCreate)

  - GET /stores/:store_id/catalogs/:catalog_id
    setup(stores, ResourceGet) -> element(stores, false) ->
    setup(catalogs, ResourceGet) -> element(catalogs, false)

  - PATCH /stores/:store_id/catalogs/:catalog_id
    setup(stores, ResourceGet) -> element(stores, false) ->
    setup(catalogs, ResourceUpdate) -> element(catalogs, false)

  - DELETE /stores/:store_id/catalogs/:catalog_id
    setup(stores, ResourceGet) -> element(stores, false) ->
    setup(catalogs, ResourceDelete) -> element(catalogs, false)

  - GET /stores/:store_id/catalogs/:catalog_id/products
    setup(stores, ResourceGet) -> element(stores, false) ->
    setup(catalogs, ResourceGet) -> element(catalogs, false) ->
    setup(products, ResourceList)
    Accessed by this test with ?sort=rank to list live products.

  - POST /stores/:store_id/catalogs/:catalog_id/products
    setup(stores, ResourceGet) -> element(stores, false) ->
    setup(catalogs, ResourceGet) -> element(catalogs, false) ->
    setup(products, ResourceCreate)

  - GET /stores/:store_id/catalogs/:catalog_id/products/:product_id
    setup(stores, ResourceGet) -> element(stores, false) ->
    setup(catalogs, ResourceGet) -> element(catalogs, false) ->
    setup(products, ResourceGet) -> element(products, false)

  - PATCH /stores/:store_id/catalogs/:catalog_id/products/:product_id
    setup(stores, ResourceGet) -> element(stores, false) ->
    setup(catalogs, ResourceGet) -> element(catalogs, false) ->
    setup(products, ResourceUpdate) -> element(products, false)

  - DELETE /stores/:store_id/catalogs/:catalog_id/products/:product_id
    setup(stores, ResourceGet) -> element(stores, false) ->
    setup(catalogs, ResourceGet) -> element(catalogs, false) ->
    setup(products, ResourceDelete) -> element(products, false)
    Accessed by this test to soft-delete a product.

  - GET /stores/:store_id/catalogs/:catalog_id/products/deleted
    setup(stores, ResourceGet) -> element(stores, false) ->
    setup(catalogs, ResourceGet) -> element(catalogs, false) ->
    setup(products, ResourceListDeleted)
    Accessed by this test to list deleted products.

  - GET /stores/:store_id/catalogs/:catalog_id/products/deleted/:product_id
    setup(stores, ResourceGet) -> element(stores, false) ->
    setup(catalogs, ResourceGet) -> element(catalogs, false) ->
    setup(products, ResourceGetDeleted) -> element(products, true)

  - POST /stores/:store_id/catalogs/:catalog_id/products/deleted/:product_id
    setup(stores, ResourceGet) -> element(stores, false) ->
    setup(catalogs, ResourceGet) -> element(catalogs, false) ->
    setup(products, ResourceRestore) -> element(products, true)
    Accessed by this test to restore a deleted product.

  - DELETE /stores/:store_id/catalogs/:catalog_id/products/deleted/:product_id
    setup(stores, ResourceGet) -> element(stores, false) ->
    setup(catalogs, ResourceGet) -> element(catalogs, false) ->
    setup(products, ResourcePrune) -> element(products, true)
    Accessed by this test first to prove restored products are no longer in the
    deleted route, then again to prune a deleted product.

  - POST /hard-items
    setup(hard-items, ResourceCreate)
    Accessed by this test to create a hard-deleted resource.

  - GET /hard-items/:hard_item_id
    setup(hard-items, ResourceGet) -> element(hard-items, false)
    Accessed by this test after hard deletion to assert 404.

  - DELETE /hard-items/:hard_item_id
    setup(hard-items, ResourceDelete) -> element(hard-items, false)
    Accessed by this test to permanently delete the hard-deleted resource.

The hard-items collection also creates GET /hard-items, PATCH /hard-items/:hard_item_id,
and no /deleted routes because integrationHardItem does not implement SoftDeletedResource.
Catalogs also create no /deleted routes because integrationCatalog is not soft-deleted.
*/
func TestEchoPlatformWithMemoryStorageNestedSingletonsCollectionsAndDeletes(t *testing.T) {
	t.Parallel()

	app := echov4.New()

	storeStorage := memory.NewStorage[int, *integrationStore]()
	catalogStorage := memory.NewStorage[int, *integrationCatalog]()
	productStorage := memory.NewStorage[int, *integrationProduct]()
	settingStorage := memory.NewStorage[int, *integrationSetting]()
	hardItemStorage := memory.NewStorage[int, *integrationHardItem]()

	storeService := services.MustCreateCollectionService[int, *integrationStore]("stores", "store_id", storeStorage)
	storeService.UsingPageSize(5)
	catalogService := services.MustCreateCollectionService[int, *integrationCatalog]("catalogs", "catalog_id", catalogStorage)
	catalogService.UsingPageSize(5)
	catalogService.MustAttachTo(storeService, "store_id")
	productService := services.MustCreateCollectionService[int, *integrationProduct]("products", "product_id", productStorage)
	productService.UsingPageSize(5)
	productService.MustAttachTo(catalogService, "catalog_id")
	settingService := services.MustCreateSingletonService[int, *integrationSetting]("platform", settingStorage)
	hardItemService := services.MustCreateCollectionService[int, *integrationHardItem]("hard-items", "hard_item_id", hardItemStorage)

	if err := Install(app, storeService); err != nil {
		t.Fatalf("Install returned error: %v", err)
	}
	if err := Install(app, settingService); err != nil {
		t.Fatalf("Install singleton returned error: %v", err)
	}
	if err := Install(app, hardItemService); err != nil {
		t.Fatalf("Install hard item service returned error: %v", err)
	}

	settingResponse := performJSONRequest(t, app, http.MethodPost, "/platform", map[string]any{"version": "2026.9"})
	requireStatus(t, settingResponse, http.StatusCreated)
	var createdSetting integrationSetting
	decodeJSON(t, settingResponse, &createdSetting)
	if createdSetting.Version != "2026.9" || createdSetting.ID == 0 {
		t.Fatalf("unexpected singleton create response: %#v", createdSetting)
	}

	storeResponse := performJSONRequest(t, app, http.MethodPost, "/stores", map[string]any{"name": "Main"})
	requireStatus(t, storeResponse, http.StatusCreated)
	var createdStore integrationStore
	decodeJSON(t, storeResponse, &createdStore)
	if createdStore.ID == 0 {
		t.Fatal("expected store to receive a generated ID")
	}

	hardItemResponse := performJSONRequest(t, app, http.MethodPost, "/hard-items", map[string]any{"name": "Temporary"})
	requireStatus(t, hardItemResponse, http.StatusCreated)

	catalog := &integrationCatalog{StoreID: createdStore.ID, Name: "Fall"}
	if notFound, err := catalogStorage.Save(&catalog); err != nil || notFound {
		t.Fatalf("catalog Save returned notFound=%v err=%v", notFound, err)
	}
	firstProduct := &integrationProduct{CatalogID: catalog.ID, Name: "Hat", Rank: 2}
	secondProduct := &integrationProduct{CatalogID: catalog.ID, Name: "Scarf", Rank: 1}
	for _, product := range []*integrationProduct{firstProduct, secondProduct} {
		if notFound, err := productStorage.Save(&product); err != nil || notFound {
			t.Fatalf("product Save returned notFound=%v err=%v", notFound, err)
		}
	}

	settingResponse = performJSONRequest(t, app, http.MethodGet, "/platform", nil)
	requireStatus(t, settingResponse, http.StatusOK)
	var loadedSetting integrationSetting
	decodeJSON(t, settingResponse, &loadedSetting)
	if loadedSetting.Version != "2026.9" {
		t.Fatalf("unexpected singleton setting response: %#v", loadedSetting)
	}

	productsResponse := performJSONRequest(t, app, http.MethodGet, "/stores/1/catalogs/1/products?sort=rank", nil)
	requireStatus(t, productsResponse, http.StatusOK)
	var productsPage struct {
		Elements   []integrationProduct `json:"elements"`
		Page       int                  `json:"page"`
		TotalPages int                  `json:"totalPages"`
	}
	decodeJSON(t, productsResponse, &productsPage)
	if productsPage.TotalPages != 1 || len(productsPage.Elements) != 2 || productsPage.Elements[0].Name != "Scarf" {
		t.Fatalf("unexpected products page: %#v", productsPage)
	}

	deleteProductResponse := performJSONRequest(t, app, http.MethodDelete, "/stores/1/catalogs/1/products/1", nil)
	requireStatus(t, deleteProductResponse, http.StatusNoContent)

	deletedProductsResponse := performJSONRequest(t, app, http.MethodGet, "/stores/1/catalogs/1/products/deleted", nil)
	requireStatus(t, deletedProductsResponse, http.StatusOK)
	var deletedProductsPage struct {
		Elements []integrationProduct `json:"elements"`
	}
	decodeJSON(t, deletedProductsResponse, &deletedProductsPage)
	if len(deletedProductsPage.Elements) != 1 || deletedProductsPage.Elements[0].Name != "Hat" {
		t.Fatalf("unexpected deleted products page: %#v", deletedProductsPage)
	}

	restoreProductResponse := performJSONRequest(t, app, http.MethodPost, "/stores/1/catalogs/1/products/deleted/1", nil)
	requireStatus(t, restoreProductResponse, http.StatusOK)
	pruneAfterRestoreResponse := performJSONRequest(t, app, http.MethodDelete, "/stores/1/catalogs/1/products/deleted/1", nil)
	requireStatus(t, pruneAfterRestoreResponse, http.StatusNotFound)

	deleteProductAgainResponse := performJSONRequest(t, app, http.MethodDelete, "/stores/1/catalogs/1/products/1", nil)
	requireStatus(t, deleteProductAgainResponse, http.StatusNoContent)
	pruneProductResponse := performJSONRequest(t, app, http.MethodDelete, "/stores/1/catalogs/1/products/deleted/1", nil)
	requireStatus(t, pruneProductResponse, http.StatusNoContent)

	deleteHardItemResponse := performJSONRequest(t, app, http.MethodDelete, "/hard-items/1", nil)
	requireStatus(t, deleteHardItemResponse, http.StatusNoContent)
	missingHardItemResponse := performJSONRequest(t, app, http.MethodGet, "/hard-items/1", nil)
	requireStatus(t, missingHardItemResponse, http.StatusNotFound)
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

func performJSONRequest(t *testing.T, app *echov4.Echo, method string, target string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var requestBody *bytes.Reader
	if body == nil {
		requestBody = bytes.NewReader(nil)
	} else {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("json.Marshal returned error: %v", err)
		}
		requestBody = bytes.NewReader(encoded)
	}

	request := httptest.NewRequest(method, target, requestBody)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	return response
}

func requireStatus(t *testing.T, response *httptest.ResponseRecorder, status int) {
	t.Helper()

	if response.Code != status {
		t.Fatalf("expected status %d, got %d with body %s", status, response.Code, response.Body.String())
	}
}

func decodeJSON(t *testing.T, response *httptest.ResponseRecorder, target any) {
	t.Helper()

	if err := json.Unmarshal(response.Body.Bytes(), target); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v; response status=%d body=%s", err, response.Code, response.Body.String())
	}
}
