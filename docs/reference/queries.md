# Filter and Sort Syntax

List endpoints accept query parameters:

- `filter`: JSON filter object.
- `sort`: comma-separated field list.
- `page`: zero-based page number.

Fields use JSON names, not Go struct names and not storage column names.

## Sort

Ascending is the default:

```text
?sort=name,created_at
```

Use `-` for descending:

```text
?sort=-created_at,name
```

Each field is validated by the active storage adapter and by any
`AllowedFieldsFunc` configured on the service.

## Filter Basics

Filters are JSON objects. URL-encode them when sending through a query string.

Equality:

```json
{"name":{"$eq":"Main"}}
```

Comparison:

```json
{"price":{"$gte":10}}
```

Text containment:

```json
{"name":{"$contains":"main"}}
```

Null check:

```json
{"deleted_at":{"$null":true}}
```

Existence check:

```json
{"metadata":{"$exists":true}}
```

The parser supports:

- `$lt`
- `$lte`
- `$gt`
- `$gte`
- `$eq`
- `$ne`
- `$null`
- `$exists`
- `$contains`

## Logical Filters

AND:

```json
{"$and":[{"price":{"$gte":10}},{"price":{"$lte":20}}]}
```

OR:

```json
{"$or":[{"name":{"$contains":"main"}},{"name":{"$contains":"backup"}}]}
```

NOT:

```json
{"$not":{"archived":{"$eq":true}}}
```

Match nothing:

```json
{"$none":true}
```

`$none` is accepted only at the top level.

## Examples With `curl`

```sh
curl --get http://localhost:8080/stores \
  --data-urlencode 'filter={"name":{"$contains":"main"}}' \
  --data-urlencode 'sort=-created_at,name' \
  --data-urlencode 'page=0'
```

```sh
curl --get http://localhost:8080/products \
  --data-urlencode 'filter={"$and":[{"price":{"$gte":10}},{"price":{"$lte":20}}]}'
```

## Validation Rules

The parser validates syntax first. Then the storage validator checks whether the
field and operation are valid for the mapped resource field.

The service may further restrict fields with `UsingAllowedFields`:

```go
service.UsingAllowedFields(func(context services.Context) ([]string, services.Allowance) {
	return []string{"id", "name", "created_at"}, services.Only
})
```

Use `services.All` to allow all valid fields, `services.Only` to allow only the
listed fields, and `services.Except` to deny only the listed fields.
