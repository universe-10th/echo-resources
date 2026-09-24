package services

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
