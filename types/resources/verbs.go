package resources

import (
	"errors"

	"github.com/universe-10th/echo-resources/utils"
)

var (
	ErrInvalidStorage = errors.New("invalid storage")
)

// ResourceVerb tells the verbs supported by the resource.
type ResourceVerb uint16

const (
	ResourceGet ResourceVerb = iota
	ResourceList
	ResourceCreate
	ResourceUpdate
	ResourceDelete
	ResourceGetDeleted
	ResourceListDeleted
	ResourceRestore
	ResourcePrune
)

// ResourceVerbs is a small set of ResourceVerb values.
type ResourceVerbs utils.Flags[ResourceVerb]

// NewResourceVerbs assembles the chosen verbs into a set
// of flags.
func NewResourceVerbs(verbs ...ResourceVerb) ResourceVerbs {
	return ResourceVerbs(utils.NewFlags[ResourceVerb](verbs...))
}
