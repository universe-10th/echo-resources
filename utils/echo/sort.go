package echo

import (
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/universe-10th/echo-resources/types"
)

const (
	SortArg string = "sort"
)

// ParseSort parses a sort from the query. This function only matters
// for List / ListDeleted endpoints. Returns a Bad Request if the parsing
// failed.
func ParseSort(
	context echo.Context, validator types.SortValidator,
) (types.SortExpression, error) {
	// 1. Get the sort argument. Gracefully abort if none.
	rawSort := context.QueryParam(SortArg)
	rawSort = strings.TrimSpace(rawSort)
	if rawSort == "" {
		return types.SortExpression{}, nil
	}

	// 2. Parse+Validate that filter.
	if parsed, err := types.NewSortParser(validator).Parse(rawSort); err != nil {
		content, code := types.RenderError(types.BadRequestError{})
		_ = context.JSON(int(code), content)
		return types.SortExpression{}, err
	} else {
		return parsed, nil
	}
}
