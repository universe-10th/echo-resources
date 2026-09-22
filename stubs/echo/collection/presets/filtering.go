package presets

import (
	"github.com/labstack/echo/v4"
	"github.com/universe-10th/echo-resources/stubs/echo/collection"
	"github.com/universe-10th/echo-resources/types"
	echo2 "github.com/universe-10th/echo-resources/utils/echo"
)

// AllowedFieldsProvider returns the fields allowed for filtering and sorting.
type AllowedFieldsProvider func() ([]string, collection.Allowance, error)

// FilterAndSortApplier applies extra application-level constraints to parsed
// filter and sort expressions.
type FilterAndSortApplier func(echo.Context, *types.FilterExpression, *types.SortExpression) error

// ResourceFiltering is a wrapper component for collection filter and sort
// parsing.
type ResourceFiltering[IDT comparable, RT types.Resource[IDT]] struct {
	filterValidator types.FilterValidator
	sortValidator   types.SortValidator
	allowedFields   AllowedFieldsProvider
	applier         FilterAndSortApplier
}

// NewResourceFiltering creates a collection filtering component using default
// allowed-fields and custom-filter behavior.
func NewResourceFiltering[IDT comparable, RT types.Resource[IDT]]() ResourceFiltering[IDT, RT] {
	return ResourceFiltering[IDT, RT]{}
}

// UsingDefaultFiltering resets this component to its default behavior.
func (resourceFiltering *ResourceFiltering[IDT, RT]) UsingDefaultFiltering() {
	resourceFiltering.filterValidator = nil
	resourceFiltering.sortValidator = nil
	resourceFiltering.allowedFields = nil
	resourceFiltering.applier = nil
}

// UsingFilterValidator updates the validator used when parsing filter=.
func (resourceFiltering *ResourceFiltering[IDT, RT]) UsingFilterValidator(
	validator types.FilterValidator,
) {
	resourceFiltering.filterValidator = validator
}

// UsingSortValidator updates the validator used when parsing sort=.
func (resourceFiltering *ResourceFiltering[IDT, RT]) UsingSortValidator(
	validator types.SortValidator,
) {
	resourceFiltering.sortValidator = validator
}

// UsingValidators updates both validators used when parsing filter= and sort=.
func (resourceFiltering *ResourceFiltering[IDT, RT]) UsingValidators(
	filterValidator types.FilterValidator, sortValidator types.SortValidator,
) {
	resourceFiltering.filterValidator = filterValidator
	resourceFiltering.sortValidator = sortValidator
}

// UsingCustomAllowedFields updates the provider for allowed filter/sort fields.
// Passing nil restores the default provider.
func (resourceFiltering *ResourceFiltering[IDT, RT]) UsingCustomAllowedFields(
	provider AllowedFieldsProvider,
) {
	resourceFiltering.allowedFields = provider
}

// UsingCustomFilterAndSort updates the custom filter/sort applier. Passing nil
// restores the default no-op applier.
func (resourceFiltering *ResourceFiltering[IDT, RT]) UsingCustomFilterAndSort(
	applier FilterAndSortApplier,
) {
	resourceFiltering.applier = applier
}

// GetAllowedFields returns the allowed filter/sort fields. By default, all
// fields accepted by the validators are allowed.
func (resourceFiltering ResourceFiltering[IDT, RT]) GetAllowedFields() ([]string, collection.Allowance, error) {
	if resourceFiltering.allowedFields == nil {
		return []string{}, collection.All, nil
	}

	return resourceFiltering.allowedFields()
}

// ParseFilterAndSort parses filter= and sort= query parameters using the
// configured validators and allowed-field settings.
func (resourceFiltering ResourceFiltering[IDT, RT]) ParseFilterAndSort(
	context echo.Context, fields []string, allowance collection.Allowance,
) (*types.FilterExpression, *types.SortExpression, error) {
	filter, err := echo2.ParseFilter(
		context,
		allowedFilterValidator{validator: resourceFiltering.filterValidator, fields: fields, allowance: allowance},
	)
	if err != nil {
		return nil, nil, err
	}

	sort, err := echo2.ParseSort(
		context,
		allowedSortValidator{validator: resourceFiltering.sortValidator, fields: fields, allowance: allowance},
	)
	if err != nil {
		return nil, nil, err
	}

	return &filter, &sort, nil
}

// ApplyCustomFilterAndSort applies configured application-specific
// filter/sort constraints. The default behavior is a no-op.
func (resourceFiltering ResourceFiltering[IDT, RT]) ApplyCustomFilterAndSort(
	context echo.Context, filter *types.FilterExpression, sort *types.SortExpression,
) error {
	if resourceFiltering.applier == nil {
		return nil
	}

	return resourceFiltering.applier(context, filter, sort)
}

type allowedFilterValidator struct {
	validator types.FilterValidator
	fields    []string
	allowance collection.Allowance
}

func (v allowedFilterValidator) IsValidCmpFilter(field string, value any) bool {
	return v.isAllowed(field) && v.validator != nil && v.validator.IsValidCmpFilter(field, value)
}

func (v allowedFilterValidator) IsNullCheckable(field string) bool {
	return v.isAllowed(field) && v.validator != nil && v.validator.IsNullCheckable(field)
}

func (v allowedFilterValidator) IsExistenceCheckable(field string) bool {
	return v.isAllowed(field) && v.validator != nil && v.validator.IsExistenceCheckable(field)
}

func (v allowedFilterValidator) IsContainsCheckable(field string) bool {
	return v.isAllowed(field) && v.validator != nil && v.validator.IsContainsCheckable(field)
}

func (v allowedFilterValidator) isAllowed(field string) bool {
	return isAllowedField(field, v.fields, v.allowance)
}

type allowedSortValidator struct {
	validator types.SortValidator
	fields    []string
	allowance collection.Allowance
}

func (v allowedSortValidator) IsSortable(field string, orderType types.OrderType) bool {
	return isAllowedField(field, v.fields, v.allowance) &&
		v.validator != nil &&
		v.validator.IsSortable(field, orderType)
}

func isAllowedField(field string, fields []string, allowance collection.Allowance) bool {
	switch allowance {
	case collection.Only:
		return containsField(fields, field)
	case collection.Except:
		return !containsField(fields, field)
	default:
		return true
	}
}

func containsField(fields []string, field string) bool {
	for _, allowedField := range fields {
		if allowedField == field {
			return true
		}
	}

	return false
}
