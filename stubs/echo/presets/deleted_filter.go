package presets

import "github.com/universe-10th/echo-resources/types"

// ApplyingDeleteFilter applies the standard soft-delete filter constraint.
type ApplyingDeleteFilter[IDT comparable, RT types.SoftDeletedResource[IDT]] struct{}

// ApplyDeletedFilter adds a deletion-state criterion to filter.
//
// The criterion targets the resource's deletion timestamp field. For active
// resources, the timestamp must be null; for deleted resources, the timestamp
// must be non-null. Existing $and filters are extended in-place. Other existing
// filters are wrapped in a parent $and together with the deletion criterion.
func (engine ApplyingDeleteFilter[IDT, RT]) ApplyDeletedFilter(
	filter *types.FilterExpression, deleted bool,
) {
	if filter == nil {
		return
	}

	var m RT
	deletedField := m.GetDeletionTimeField()
	deletedCriterion := types.FilterExpression{
		Operator: types.FilterNull,
		Field:    deletedField,
		Value:    !deleted,
	}

	if filter.Operator == "" {
		*filter = types.FilterExpression{
			Operator:    types.FilterAnd,
			Expressions: []types.FilterExpression{deletedCriterion},
		}
		return
	}

	if filter.Operator == types.FilterAnd {
		filter.Expressions = append(filter.Expressions, deletedCriterion)
		return
	}

	*filter = types.FilterExpression{
		Operator: types.FilterAnd,
		Expressions: []types.FilterExpression{
			*filter,
			deletedCriterion,
		},
	}
}
