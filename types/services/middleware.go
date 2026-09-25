package services

import (
	"errors"

	"github.com/universe-10th/echo-resources/types"
)

// setupMiddleware is used for ALL the endpoints registered in
// a resource. Useful for particular middleware functions that
// need to detect the current endpoint, and registered prior
// to any other middleware functions given to the resource.
func setupMiddleware(resource any, endpointType EndpointType, verb ResourceVerb, name string) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(context Context) error {
			context.Setup(resource, endpointType, verb, name)
			return next(context)
		}
	}
}

func renderError(context Context, err types.Error) error {
	content, code := types.RenderError(err)
	return context.RenderJSON(int(code), content)
}

func renderErrorOr(context Context, err error, defaultError types.Error) error {
	var err_ types.Error
	if !errors.As(err, &err_) {
		err_ = defaultError
	}
	content, code := types.RenderError(err_)
	return context.RenderJSON(int(code), content)
}

// elementMiddleware will be the FIRST middleware installed
// in the per-element (deleted or not) endpoints (either for
// a collection or for singleton). And it will go down for
// all the derived elements as well (they will exist under
// the GET + non-deleted verb). It will also be used in the
// per-element custom actions.
func elementMiddleware[IDT comparable, RT types.Resource[IDT]](deleted bool) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(context Context) error {
			// 1. Get the service. If, for some reason, it is not properly
			//    set, return an error for bad configuration.
			service, ok := context.CurrentService().(ResourceService[IDT, RT])
			if !ok {
				logger.Error("service is not properly configured!")
				return renderError(context, types.InternalError{})
			}

			// 2. Get the filter for a current element. This will always
			//    involve a non-deleted query.
			filter, id, err := service.makeElementFilter(context, deleted)
			if err != nil {
				return renderErrorOr(context, err, types.BadRequestError{})
			}

			// 3. With the filter, retrieve an element.
			element, found, err := service.Storage().GetElement(filter)
			if err != nil {
				return renderErrorOr(context, err, types.InternalError{})
			} else if !found {
				var err_ types.Error
				if service.IsSingleton() {
					err_ = types.NotFoundError[IDT]{
						ElementName: service.Prefix(),
						Key:         id,
					}
				} else {
					err_ = types.NotFoundError[IDT]{
						ElementName: service.Prefix(),
						Key:         id,
					}
				}
				return renderError(context, err_)
			}

			// 4. Otherwise, push the element.
			context.PushElement(element)
			defer context.PopElement()

			// 5. Then, execute the next things.
			return next(context)
		}
	}
}
