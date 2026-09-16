package types

/**
 * Filters are a special topic here. They are parsed from a serialized format which
 * looks like this:
 *
 * field  : a valid camel_cased golang identifier. It will stick to a regex. By this
 *          point, we'll assume as solved that camel_cased field names are unique and
 *          they all satisfy: [a-z0-9][a-z)-9]*.
 *
 * cmp    : "$lt" | "$lte" | "$gt" | "$gte" | "$eq" | "$ne"
 * value  : A JSON-serialized value. It will be parsed against the field. For example,
 *          a string might actually represent a time.Time value if the involved field
 *          is a time.Time field.
 *
 * Filter : {"$and": []Filter}
 *        | {"$or": []Filter}
 *        | {"$not": Filter}
 *        | {field: {cmp: value}} <--- comparing a field to a value
 *        | {field: {"$null": true|false}} <--- testing whether the field is null
 *        | {field: {"$exists": true|false}} <--- testing whether the field exists.
 *        | {field: {"$contains": string}} <--- testing for text containment
 *
 * This is NOT extensive in what the useful filters might imply, but these are the only
 * ones we will support today.
 *
 * So, first, I need a special class we can name FilterParser. It takes the filter as
 * a map[string]any (since this may come from user input, but always in the format I
 * explained above) and produces a VALID filter or raises an error. What a VALID filter
 * is... involves:
 *
 * - The referenced fields are valid.
 * - The values provided for comparisons are valid for the respective fields.
 * - The format of each allowed clause is valid.
 */

// The Filter interface has a method to serialize itself into an engine-specific query.
// The Query type is either a map (for MongoDB engine) or string (for GORM engine), and
// new engines might use their own types.
type Filter[Query any] interface {
	// Serialize produces a query values to be used in the underlying database engine.
	Serialize() Query
}

// The FilterValidator interface has methods to tell whether the fields and values involved
// in the parsing can be effectively parsed / identified as valid data.
type FilterValidator interface {
	// IsValidCmpFilter takes the name of a field and tells whether it is valid (for the
	// current filtering) and then takes a value: the value (which is a JSON-decoded value)
	// must be interpretable by the field. This means: if the field accepts the idea of
	// being populated by the value (this might involve decoding, e.g. a time.Time field
	// being populated from a formatted string), then this method returns true. Otherwise,
	// this method must return false.
	IsValidCmpFilter(filter string, value any) bool

	// IsNullCheckable takes the name of a field and tells whether it is valid (for
	// the current filtering) for a $null check (i.e. a nullable field).
	IsNullCheckable(filter string) bool

	// IsExistenceCheckable takes the name of a field and tells whether it is valid (for
	// the current filtering) for an $exists check (i.e. the field supports existence checks
	// in the underlying engine - typically false on GORM and typically true on MongoDB,
	// regardless of the field name, unless additional constraints are set).
	IsExistenceCheckable(filter string) bool

	// IsContainsCheckable takes the name of a field and tells whether it is valid (for
	// the current filtering) for a $contains check (i.e. a string field).
	IsContainsCheckable(field string) bool
}
