package echo

import (
	"errors"

	"github.com/labstack/echo/v4"
	"github.com/universe-10th/echo-resources/types"
)

// A CollectionElementRetrieverFunc is a function that retrieves an item or
// returns an error. When an error occurs, the error will be returned to
// the caller, and the caller MUST stop.
type CollectionElementRetrieverFunc[IDT comparable, RT types.Resource[IDT]] func(
	context echo.Context, retriever types.CollectionList[IDT, RT],
) (RT, error)

// A CollectionListRetrieverFunc is a function that retrieves an item or
// returns an error. When an error occurs, the error will be returned to
// the caller, and the caller MUST stop.
type CollectionListRetrieverFunc[IDT comparable, RT types.Resource[IDT]] func(
	context echo.Context, retriever types.CollectionList[IDT, RT],
) ([]RT, int64, int64, error)

// A CollectionSoftDeletedElementRetrieverFunc is a function that retrieves
// an item or returns an error. When an error occurs, the error will be returned
// to the caller, and the caller MUST stop. This works on DELETED items.
type CollectionSoftDeletedElementRetrieverFunc[IDT comparable, RT types.SoftDeletedResource[IDT]] func(
	context echo.Context, retriever types.CollectionSoftDeletedList[IDT, RT],
) (RT, error)

// A CollectionSoftDeletedListRetrieverFunc is a function that retrieves an item
// or returns an error. When an error occurs, the error will be returned to the
// caller, and the caller MUST stop. This works on DELETED items.
type CollectionSoftDeletedListRetrieverFunc[IDT comparable, RT types.SoftDeletedResource[IDT]] func(
	context echo.Context, retriever types.CollectionSoftDeletedList[IDT, RT],
) ([]RT, int64, int64, error)

// MakeCollectionElementRetriever creates a collection retriever function.
// This function uses a specific retriever logic based on ID lookup.
func MakeCollectionElementRetriever[IDT comparable, RT types.Resource[IDT]](
	urlArg string, elementName string,
) CollectionElementRetrieverFunc[IDT, RT] {
	return func(context echo.Context, retriever types.CollectionList[IDT, RT]) (RT, error) {
		var result RT
		var found bool
		var err error

		id, err := echo.PathParam[IDT](context, urlArg)
		if err == nil {
			result, found, err = retriever.Get(id)
		}

		if !found {
			err = types.NotFoundError[IDT]{
				ElementName: elementName,
				Key:         id,
			}
		}

		if err != nil {
			var err_ types.Error
			if errors.As(err, &err_) {
				serializedErr, code := types.RenderError(err_)
				_ = context.JSON(int(code), serializedErr)
			}
		}
		return result, err
	}
}

// MakeCollectionListRetriever creates a collection retriever function.
// This function uses a specific retriever logic based on a validator
// for sort and a validator for filter.
func MakeCollectionListRetriever[IDT comparable, RT types.Resource[IDT]](
	filterValidator types.FilterValidator, sortValidator types.SortValidator,
) CollectionListRetrieverFunc[IDT, RT] {
	return func(
		context echo.Context, retriever types.CollectionList[IDT, RT],
	) ([]RT, int64, int64, error) {
		var result []RT
		var count int64
		var err error
		var page struct {
			Skip  int64
			Limit int64
		}

		// Parse the sort, if present.
		var sort types.SortExpression
		sort, err = ParseSort(context, sortValidator)
		if err != nil {
			return nil, 0, 0, err
		}

		// Parse the filter, if present.
		var filter types.FilterExpression
		filter, err = ParseFilter(context, filterValidator)
		if err != nil {
			return nil, 0, 0, err
		}

		// Compute the count of total elements.
		count, err = retriever.Count(filter)
		if err != nil {
			return nil, 0, 0, err
		}

		// Compute the page of elements.
		result, err = retriever.List(types.ListOptions{
			Skip:   page.Skip,
			Limit:  page.Limit,
			Sort:   sort,
			Filter: filter,
		})

		if err != nil {
			var err_ types.Error
			if errors.As(err, &err_) {
				serializedErr, code := types.RenderError(err_)
				_ = context.JSON(int(code), serializedErr)
			}
		}
		return result, page.Skip, count, err
	}
}

// MakeCollectionSoftDeletedElementRetriever creates a collection retriever function.
// This function uses a specific retriever logic based on ID lookup. This works on
// DELETED items.
func MakeCollectionSoftDeletedElementRetriever[IDT comparable, RT types.SoftDeletedResource[IDT]](
	urlArg string, elementName string,
) CollectionSoftDeletedElementRetrieverFunc[IDT, RT] {
	return func(context echo.Context, retriever types.CollectionSoftDeletedList[IDT, RT]) (RT, error) {
		var result RT
		var found bool
		var err error

		id, err := echo.PathParam[IDT](context, urlArg)
		if err == nil {
			result, found, err = retriever.GetDeleted(id)
		}

		if !found {
			err = types.NotFoundError[IDT]{
				ElementName: elementName,
				Key:         id,
			}
		}

		if err != nil {
			var err_ types.Error
			if errors.As(err, &err_) {
				serializedErr, code := types.RenderError(err_)
				_ = context.JSON(int(code), serializedErr)
			}
		}
		return result, err
	}
}

// MakeCollectionSoftDeletedListRetriever creates a collection retriever function.
// This function uses a specific retriever logic based on a validator for sort and
// a validator for filter. This works on DELETED items.
func MakeCollectionSoftDeletedListRetriever[IDT comparable, RT types.SoftDeletedResource[IDT]](
	filterValidator types.FilterValidator, sortValidator types.SortValidator,
) CollectionSoftDeletedListRetrieverFunc[IDT, RT] {
	return func(
		context echo.Context, retriever types.CollectionSoftDeletedList[IDT, RT],
	) ([]RT, int64, int64, error) {
		var result []RT
		var count int64
		var err error
		var page struct {
			Skip  int64
			Limit int64
		}

		// Parse the sort, if present.
		var sort types.SortExpression
		sort, err = ParseSort(context, sortValidator)
		if err != nil {
			return nil, 0, 0, err
		}

		// Parse the filter, if present.
		var filter types.FilterExpression
		filter, err = ParseFilter(context, filterValidator)
		if err != nil {
			return nil, 0, 0, err
		}

		// Compute the count of total elements.
		count, err = retriever.CountDeleted(filter)
		if err != nil {
			return nil, 0, 0, err
		}

		// Compute the page of elements.
		result, err = retriever.ListDeleted(types.ListOptions{
			Skip:   page.Skip,
			Limit:  page.Limit,
			Sort:   sort,
			Filter: filter,
		})

		if err != nil {
			var err_ types.Error
			if errors.As(err, &err_) {
				serializedErr, code := types.RenderError(err_)
				_ = context.JSON(int(code), serializedErr)
			}
		}
		return result, page.Skip, count, err
	}
}
