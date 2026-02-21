// Package middleware provides production-ready HTTP middleware
// that works with any net/http compatible router.
//
// All middleware follows the standard func(http.Handler) http.Handler signature.
//
// Usage:
//
//	stack := middleware.Chain(
//	    middleware.RequestID(),
//	    middleware.Logger(logger),
//	    middleware.Recover(),
//	    middleware.CORS(middleware.DefaultCORSConfig()),
//	    middleware.Timeout(30 * time.Second),
//	)
//
//	http.ListenAndServe(":8080", stack(mux))
package middleware

import "net/http"

// Middleware is a function that wraps an http.Handler.
type Middleware func(http.Handler) http.Handler

// Chain composes multiple middleware into a single middleware.
// Middleware is applied in the order provided (first middleware is outermost).
//
//	stack := middleware.Chain(first, second, third)
//	// Request flow: first → second → third → handler
//	// Response flow: third → second → first → client
func Chain(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// Then is an alias for applying middleware to a handler.
// Equivalent to Chain(middlewares...)(handler).
func Then(handler http.Handler, middlewares ...Middleware) http.Handler {
	return Chain(middlewares...)(handler)
}
