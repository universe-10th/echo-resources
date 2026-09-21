package collection

import (
	"errors"
	"regexp"

	"github.com/labstack/echo/v4"
	common "github.com/universe-10th/echo-resources/dsl/echo"
)

var _prefix = regexp.MustCompile("^[a-zA-Z0-9]+([-_][a-zA-Z0-9]+)*$")
var _urlArg = regexp.MustCompile("^[a-zA-Z0-9]+([-_][a-zA-Z0-9]+)*$") // Yes, same for now.

// ErrInvalidPrefix tells the specified prefix is not of
// the valid format.
var ErrInvalidPrefix = errors.New("invalid prefix")

// ErrInvalidURLArg tells the specified URL arg is not of
// the valid format.
var ErrInvalidURLArg = errors.New("invalid URL arg")

// ResourceDSL defines the core of a resource. Of course,
// it should define one of the following interfaces, at
// least, as well. But it is not always needed. It must
// define at least one middleware to fetch elements. Also,
// defines the URL Arg to use for the elements.
type ResourceDSL interface {
	PrefixName() string
	Middlewares() []echo.MiddlewareFunc
	URLArg() string
}

// WithList is a DSL interface to tell that the resource
// to register supports listing elements.
type WithList interface {
	List(c echo.Context) error
}

// WithSoftDeletedList is a DSL interface to tell that the
// resource to register supports listing DELETED elements.
type WithSoftDeletedList interface {
	ListDeleted(c echo.Context) error
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

// WithCustomCollectionRoutes is a DSL interface to tell
// that the resource to register supports creating custom
// routes for the non-deleted collection.
type WithCustomCollectionRoutes interface {
	InstallCustomCollectionRoutes(g *echo.Group)
}

// WithCustomElementRoutes is a DSL interface to tell that
// the resource to register supports creating custom routes
// for non-deleted items.
type WithCustomElementRoutes interface {
	InstallCustomElementRoutes(g *echo.Group)
}

// WithCustomCollectionDeletedRoutes is a DSL interface to
// tell that the resource to register supports creating
// custom routes for the deleted collection.
type WithCustomCollectionDeletedRoutes interface {
	InstallCustomCollectionDeletedRoutes(g *echo.Group)
}

// WithCustomElementDeletedRoutes is a DSL interface to tell
// that the resource to register supports creating custom
// routes for deleted items.
type WithCustomElementDeletedRoutes interface {
	InstallCustomElementDeletedRoutes(g *echo.Group)
}

// Register tries to register a DSL entry for a collection
// resource. It returns an error if the prefix name is not
// of the appropriate format.
func Register(g common.EchoLevel, dsl ResourceDSL) (child *echo.Group, err error) {
	prefix := dsl.PrefixName()
	if ok := _prefix.MatchString(prefix); !ok {
		return nil, ErrInvalidPrefix
	}

	urlArg := dsl.URLArg()
	if ok := _urlArg.MatchString(urlArg); !ok {
		return nil, ErrInvalidURLArg
	}

	withList, hasWithList := dsl.(WithList)
	withDeletedList, hasWithDeletedList := dsl.(WithSoftDeletedList)
	withCreate, hasWithCreate := dsl.(WithCreate)
	withGet, hasWithGet := dsl.(WithGet)
	withUpdate, hasWithUpdate := dsl.(WithUpdate)
	withDelete, hasWithDelete := dsl.(WithDelete)
	withRestore, hasWithRestore := dsl.(WithSoftDeletedRestore)
	withDeletedGet, hasWithDeletedGet := dsl.(WithSoftDeletedGet)
	withPrune, hasWithPrune := dsl.(WithSoftDeletedPrune)
	withCustomCollectionRoutes, hasWithCustomCollectionRoutes := dsl.(WithCustomCollectionRoutes)
	withCustomElementRoutes, hasWithCustomElementRoutes := dsl.(WithCustomElementRoutes)
	withCustomCollectionDeletedRoutes, hasWithCustomCollectionDeletedRoutes := dsl.(WithCustomCollectionDeletedRoutes)
	withCustomElementDeletedRoutes, hasWithCustomElementDeletedRoutes := dsl.(WithCustomElementDeletedRoutes)

	// First, let's tackle the deleted stuff here.
	if hasWithDeletedList || hasWithDeletedGet || hasWithPrune || hasWithRestore ||
		hasWithCustomCollectionDeletedRoutes || hasWithCustomElementDeletedRoutes {
		collectionDeletedGroup := g.Group("/"+prefix+"/deleted", dsl.Middlewares()...)

		if hasWithDeletedList {
			collectionDeletedGroup.GET("", withDeletedList.ListDeleted)
		}
		if hasWithCustomCollectionDeletedRoutes {
			withCustomCollectionDeletedRoutes.InstallCustomCollectionDeletedRoutes(collectionDeletedGroup)
		}

		if hasWithDeletedGet || hasWithPrune || hasWithRestore || hasWithCustomElementDeletedRoutes {
			itemDeletedGroup := collectionDeletedGroup.Group("/:" + urlArg)

			if hasWithRestore {
				itemDeletedGroup.POST("", withRestore.Restore)
			}
			if hasWithDeletedGet {
				itemDeletedGroup.GET("", withDeletedGet.GetDeleted)
			}
			if hasWithPrune {
				itemDeletedGroup.DELETE("", withPrune.Prune)
			}
			if hasWithCustomElementDeletedRoutes {
				withCustomElementDeletedRoutes.InstallCustomElementDeletedRoutes(itemDeletedGroup)
			}
		}
	}

	// Then, track the non-deleted stuff.
	if !(hasWithList || hasWithCreate || hasWithUpdate || hasWithGet || hasWithDelete ||
		hasWithCustomCollectionRoutes || hasWithCustomElementRoutes) {
		return nil, nil
	}

	collectionGroup := g.Group("/"+prefix, dsl.Middlewares()...)

	if hasWithList {
		collectionGroup.GET("", withList.List)
	}
	if hasWithCreate {
		collectionGroup.POST("", withCreate.Create)
	}
	if hasWithCustomCollectionRoutes {
		withCustomCollectionRoutes.InstallCustomCollectionRoutes(collectionGroup)
	}

	if !(hasWithUpdate || hasWithGet || hasWithDelete || hasWithCustomElementRoutes) {
		return nil, nil
	}

	itemGroup := collectionGroup.Group("/:" + urlArg)

	if hasWithUpdate {
		itemGroup.PATCH("", withUpdate.Update)
	}
	if hasWithDelete {
		itemGroup.DELETE("", withDelete.Delete)
	}
	if hasWithGet {
		itemGroup.GET("", withGet.Get)
	}
	if hasWithCustomElementRoutes {
		withCustomElementRoutes.InstallCustomElementRoutes(itemGroup)
	}

	return itemGroup, nil
}

// MustRegister Registers a DSL entry for a collection resource.
// It panics on error.
func MustRegister(g common.EchoLevel, dsl ResourceDSL) (child *echo.Group) {
	if child, err := Register(g, dsl); err != nil {
		panic(err)
	} else {
		return child
	}
}
