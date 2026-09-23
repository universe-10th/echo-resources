package singleton

import (
	"github.com/labstack/echo/v4"
	common "github.com/universe-10th/echo-resources/dsl/echo"
	"github.com/universe-10th/echo-resources/utils"
)

// ResourceDSL defines the core of a resource. Of course,
// it should define one of the following interfaces, at
// least, as well. But it is not always needed. It must
// define at least one middleware to fetch elements.
type ResourceDSL interface {
	PrefixName() string
	Middlewares() []echo.MiddlewareFunc
}

// WithGet is a DSL interface to tell that the resource
// to register supports getting a single element.
type WithGet interface {
	Get(c echo.Context) error
}

// WithSoftDeletedGet is a DSL interface to tell that the
// resource to register supports getting a single deleted
// element.
type WithSoftDeletedGet interface {
	GetDeleted(c echo.Context) error
}

// WithCreate is a DSL interface to tell that the resource
// to register supports creating a single element.
type WithCreate interface {
	Create(e echo.Context) error
}

// WithUpdate is a DSL interface to tell that the resource
// to register supports updating a single element.
type WithUpdate interface {
	Update(c echo.Context) error
}

// WithDelete is a DSL interface to tell that the resource
// to register supports deleting a single element.
type WithDelete interface {
	Delete(c echo.Context) error
}

// WithSoftDeletedRestore is a DSL interface to tell that
// the resource to register supports restoring a single
// deleted element.
type WithSoftDeletedRestore interface {
	Restore(c echo.Context) error
}

// WithSoftDeletedPrune is a DSL interface to tell that the
// resource to register supports pruning a single deleted
// element.
type WithSoftDeletedPrune interface {
	Prune(c echo.Context) error
}

// WithCustomRoutes is a DSL interface to tell that the
// resource to register supports creating custom routes
// for the non-deleted element.
type WithCustomRoutes interface {
	InstallCustomRoutes(g *echo.Group)
}

// WithCustomDeletedRoutes is a DSL interface to tell that
// the resource to register supports creating custom routes
// for the deleted element.
type WithCustomDeletedRoutes interface {
	InstallCustomDeletedRoutes(g *echo.Group)
}

// Register tries to register a DSL entry for a singleton
// resource. It returns an error if the prefix name is not
// of the appropriate format.
func Register(g common.EchoLevel, dsl ResourceDSL) (child *echo.Group, err error) {
	prefix := dsl.PrefixName()

	if err := utils.CheckPrefix(prefix); err != nil {
		return nil, err
	}

	withCreate, hasWithCreate := dsl.(WithCreate)
	withGet, hasWithGet := dsl.(WithGet)
	withUpdate, hasWithUpdate := dsl.(WithUpdate)
	withDelete, hasWithDelete := dsl.(WithDelete)
	withRestore, hasWithRestore := dsl.(WithSoftDeletedRestore)
	withDeletedGet, hasWithDeletedGet := dsl.(WithSoftDeletedGet)
	withPrune, hasWithPrune := dsl.(WithSoftDeletedPrune)
	withCustomRoutes, hasWithCustomRoutes := dsl.(WithCustomRoutes)
	withCustomDeletedRoutes, hasWithCustomDeletedRoutes := dsl.(WithCustomDeletedRoutes)

	// First, let's tackle the deleted stuff here.
	if hasWithDeletedGet || hasWithPrune || hasWithRestore || hasWithCustomDeletedRoutes {
		deletedGroup := g.Group("/"+prefix+"/deleted", dsl.Middlewares()...)
		deletedElementGroup := deletedGroup.Group("")

		if hasWithRestore {
			deletedElementGroup.POST("", withRestore.Restore)
		}
		if hasWithDeletedGet {
			deletedElementGroup.GET("", withDeletedGet.GetDeleted)
		}
		if hasWithPrune {
			deletedElementGroup.DELETE("", withPrune.Prune)
		}
		if hasWithCustomDeletedRoutes {
			withCustomDeletedRoutes.InstallCustomDeletedRoutes(deletedElementGroup)
		}
	}

	// Then, track the non-deleted stuff.
	if !(hasWithCreate || hasWithUpdate || hasWithGet || hasWithDelete || hasWithCustomRoutes) {
		return nil, nil
	}

	group := g.Group("/"+prefix, dsl.Middlewares()...)

	if hasWithCreate {
		group.POST("", withCreate.Create)
	}

	if !(hasWithUpdate || hasWithGet || hasWithDelete || hasWithCustomRoutes) {
		return nil, nil
	}

	elementGroup := group.Group("")

	if hasWithUpdate {
		elementGroup.PATCH("", withUpdate.Update)
	}
	if hasWithDelete {
		elementGroup.DELETE("", withDelete.Delete)
	}
	if hasWithGet {
		elementGroup.GET("", withGet.Get)
	}
	if hasWithCustomRoutes {
		withCustomRoutes.InstallCustomRoutes(elementGroup)
	}

	return elementGroup, nil
}

// MustRegister Registers a DSL entry for a singleton resource.
// It panics on error.
func MustRegister(g common.EchoLevel, dsl ResourceDSL) (child *echo.Group) {
	if child, err := Register(g, dsl); err != nil {
		panic(err)
	} else {
		return child
	}
}
