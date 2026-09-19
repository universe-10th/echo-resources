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
) ([]RT, error)

// MakeCollectionElementRetriever creates a collection retriever function.
// This function uses a specific retriever logic
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
	) ([]RT, error) {
		var result []RT
		var err error
		var page struct {
			Skip  int
			Limit int
		}

		// Parse the sort, if present.
		var sort types.SortExpression
		sort, err = ParseSort(context, sortValidator)
		if err != nil {
			return nil, err
		}

		// Parse the filter, if present.
		var filter types.FilterExpression
		filter, err = ParseFilter(context, filterValidator)
		if err != nil {
			return nil, err
		}

		if err == nil {
			result, err = retriever.List(types.ListOptions{
				Skip:   page.Skip,
				Limit:  page.Limit,
				Sort:   sort,
				Filter: filter,
			})
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
