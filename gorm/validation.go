package gorm

import (
	"fmt"

	"github.com/universe-10th/echo-resources/types"
)

func validateFilterExpression(filter *types.FilterExpression, validator types.FilterValidator) error {
	if filter == nil || filter.Operator == "" {
		return nil
	}

	switch filter.Operator {
	case types.FilterNone:
		return nil
	case types.FilterAnd, types.FilterOr:
		for _, expression := range filter.Expressions {
			if err := validateFilterExpression(&expression, validator); err != nil {
				return err
			}
		}
		return nil
	case types.FilterNot:
		if len(filter.Expressions) != 1 {
			return fmt.Errorf("%s filter must contain exactly one nested expression", filter.Operator)
		}
		return validateFilterExpression(&filter.Expressions[0], validator)
	case types.FilterLT, types.FilterLTE, types.FilterGT, types.FilterGTE, types.FilterEQ, types.FilterNE:
		if validator.IsValidCmpFilter(filter.Field, filter.Value) {
			return nil
		}
	case types.FilterNull:
		if validator.IsNullCheckable(filter.Field) {
			return nil
		}
	case types.FilterExists:
		if validator.IsExistenceCheckable(filter.Field) {
			return nil
		}
	case types.FilterContains:
		if validator.IsContainsCheckable(filter.Field) {
			return nil
		}
	}

	return fmt.Errorf("invalid filter on field %q", filter.Field)
}

func validateSortExpression(sort *types.SortExpression, validator types.SortValidator) error {
	if sort == nil {
		return nil
	}

	for _, item := range sort.Sort {
		if !validator.IsSortable(item.Field, item.Order) {
			return fmt.Errorf("sort field %q is not allowed", item.Field)
		}
	}

	return nil
}
