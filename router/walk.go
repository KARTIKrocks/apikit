package router

// Walk iterates over all registered routes in registration order, calling fn for each.
// If fn returns a non-nil error, Walk stops and returns that error.
func (r *Router) Walk(fn func(RouteInfo) error) error {
	for _, ri := range r.routes {
		if err := fn(ri); err != nil {
			return err
		}
	}
	return nil
}

// Routes returns a copy of all registered routes.
// The returned slice is a snapshot; modifying it has no effect on the router.
func (r *Router) Routes() []RouteInfo {
	return append([]RouteInfo(nil), r.routes...)
}
