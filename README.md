# echo-resources

Small JSON resource helpers for the [Echo](https://echo.labstack.com/) Web/HTTP framework.

This module is intended to be imported by other Go services that want a simple,
consistent response envelope for Echo handlers.

## Requirements

- Go 1.25 or newer
- Echo v4

The module uses Go toolchain support, so older compatible Go installations can
download the required toolchain automatically when `GOTOOLCHAIN=auto` is enabled.

## Installation

```sh
go get github.com/universe-10th/echo-resources
```

## Usage

```go
package main

import (
	"net/http"

	echoresources "github.com/universe-10th/echo-resources"
	"github.com/labstack/echo/v4"
)

type user struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func main() {
	e := echo.New()

	e.GET("/users/:id", func(c echo.Context) error {
		return echoresources.OK(c, user{ID: 1, Name: "Ada"})
	})

	e.POST("/users", func(c echo.Context) error {
		created := user{ID: 2, Name: "Grace"}
		return echoresources.Created(c, "/users/2", created)
	})

	e.GET("/missing", func(c echo.Context) error {
		return echoresources.Fail(c, http.StatusNotFound, "user not found")
	})

	e.Logger.Fatal(e.Start(":8080"))
}
```

Example responses:

```json
{"data":{"id":1,"name":"Ada"}}
```

```json
{"error":"user not found"}
```

## API

- `New(data)` creates a `{"data": ...}` envelope without writing a response.
- `NewCollection(data)` creates a `{"data": [...]}` collection envelope.
- `OK(c, data)` writes a `200` JSON resource response.
- `Created(c, location, data)` writes a `201` JSON resource response and an optional `Location` header.
- `List(c, data)` writes a `200` JSON collection response.
- `NoContent(c)` writes a `204` empty response.
- `Text(c, message)` writes a `200` message response.
- `Fail(c, code, message)` writes a JSON error response.

## Resource Types

The `types` package defines small interfaces for domain models:

- `types.Resource[T]` requires `GetID`, `GetCreationTime`, and `GetLastUpdateTime`.
- `types.SoftDeletedResource[T]` also requires `GetDeletionTime` and `IsDeleted`.

Optional integration packages provide reusable model fragments:

```go
import gormtypes "github.com/universe-10th/echo-resources/gorm/types"
```

The GORM package provides:

- `gormtypes.Resource[T]`
- `gormtypes.SoftDeletedResource[T]`
- `gormtypes.PrimaryKey`

`gormtypes.PrimaryKey` allows scalar primary-key values such as signed/unsigned
integers, `string`, `time.Time`, and `github.com/google/uuid.UUID`.

```go
import mongotypes "github.com/universe-10th/echo-resources/mongo/types"
```

The MongoDB package provides:

- `mongotypes.Resource`
- `mongotypes.SoftDeletedResource`
- `mongotypes.SoftDeleteIndexModel`

MongoDB resources implement `types.Resource[bson.ObjectID]` and
`types.SoftDeletedResource[bson.ObjectID]`.

## Filters

Filters are JSON objects parsed into `types.FilterExpression` and serialized by
the integration packages for their backing engines.

Supported top-level and nested logical filters:

```json
{"$and":[{"price":{"$gte":10}},{"name":{"$contains":"Ada"}}]}
{"$or":[{"price":{"$lt":10}},{"price":{"$gt":100}}]}
{"$not":{"archived":{"$exists":true}}}
```

Supported field operations are `$lt`, `$lte`, `$gt`, `$gte`, `$eq`, `$ne`,
`$null`, `$exists`, and `$contains`, depending on field and engine support.

The special top-level-only filter `{"$none":true}` is also supported. Its only
accepted value is `true`, and it serializes to an always-empty query predicate:
`1 = 0` for GORM and `{"$expr": false}` for MongoDB.

## Development

```sh
go test ./...
go vet ./...
gofmt -w .
```

## Versioning

This project should use semantic version tags for releases:

```sh
git tag v0.1.0
git push origin v0.1.0
```

Consumers can then pin a version:

```sh
go get github.com/universe-10th/echo-resources@v0.1.0
```

## License

MIT. See [LICENSE](LICENSE).
