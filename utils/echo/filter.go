package echo

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/universe-10th/echo-resources/types"
)

const (
	FilterArg string = "filter"
)

// ParseFilter parses a filter from the query. This function only matters
// for List / ListDeleted endpoints. Returns a Bad Request if the parsing
// failed.
func ParseFilter(
	context echo.Context, validator types.FilterValidator,
) (types.FilterExpression, error) {
	// 1. Get the filter argument. Gracefully abort if none.
	rawFilter := context.QueryParam(FilterArg)
	rawFilter = strings.TrimSpace(rawFilter)
	if rawFilter == "" {
		return types.FilterExpression{}, nil
	}

	// 2. Parse+Validate that filter.
	if parsed, err := types.NewFilterParser(validator).Parse(json.NewDecoder(bytes.NewBufferString(rawFilter))); err != nil {
		content, code := types.RenderError(types.BadRequestError{})
		_ = context.JSON(int(code), content)
		return types.FilterExpression{}, err
	} else {
		return parsed, nil
	}
}
