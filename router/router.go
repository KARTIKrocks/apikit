// Package router provides route grouping and method helpers on top of http.ServeMux.
//
// It adds .Get()/.Post() method helpers, Group(prefix, ...middleware) for prefix grouping,
// and per-group middleware — all delegating to a single http.ServeMux underneath.
//
// Usage:
//
//	r := router.New()
//	r.Use(middleware.RequestID(), middleware.Logger(logger))
//
//	r.Get("/health", healthHandler)
//
//	api := r.Group("/api/v1", authMiddleware)
//	api.Get("/users", listUsers)
//	api.Post("/users", createUser)
//
//	srv := server.New(r) // router implements http.Handler
package router

import (
	"encoding/json"
	stderrors "errors"
	"net/http"
	"time"

	"github.com/KARTIKrocks/apikit/errors"
	"github.com/KARTIKrocks/apikit/middleware"
)

// HandlerFunc is an HTTP handler that returns an error.
// Errors are handled by the router's ErrorHandler.
type HandlerFunc func(http.ResponseWriter, *http.Request) error

// ErrorHandler handles errors returned by HandlerFunc handlers.
type ErrorHandler func(http.ResponseWriter, *http.Request, error)

// Option configures a Router.
type Option func(*Router)

// Router is a thin wrapper around http.ServeMux that provides
// method helpers, route grouping, and per-group middleware.
type Router struct {
	mux          *http.ServeMux
	group        Group
	errorHandler ErrorHandler
}

// New creates a new Router with the given options.
func New(opts ...Option) *Router {
	r := &Router{
		mux:          http.NewServeMux(),
		errorHandler: DefaultErrorHandler,
	}
	r.group = Group{
		router: r,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// WithErrorHandler sets a custom error handler for the router.
func WithErrorHandler(fn ErrorHandler) Option {
	return func(r *Router) {
		r.errorHandler = fn
	}
}

// ServeHTTP implements http.Handler.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// Get registers an error-returning handler for GET requests.
func (r *Router) Get(pattern string, fn HandlerFunc) { r.group.Get(pattern, fn) }

// GetFunc registers a standard http.HandlerFunc for GET requests.
func (r *Router) GetFunc(pattern string, fn http.HandlerFunc) { r.group.GetFunc(pattern, fn) }

// Post registers an error-returning handler for POST requests.
func (r *Router) Post(pattern string, fn HandlerFunc) { r.group.Post(pattern, fn) }

// PostFunc registers a standard http.HandlerFunc for POST requests.
func (r *Router) PostFunc(pattern string, fn http.HandlerFunc) { r.group.PostFunc(pattern, fn) }

// Put registers an error-returning handler for PUT requests.
func (r *Router) Put(pattern string, fn HandlerFunc) { r.group.Put(pattern, fn) }

// PutFunc registers a standard http.HandlerFunc for PUT requests.
func (r *Router) PutFunc(pattern string, fn http.HandlerFunc) { r.group.PutFunc(pattern, fn) }

// Patch registers an error-returning handler for PATCH requests.
func (r *Router) Patch(pattern string, fn HandlerFunc) { r.group.Patch(pattern, fn) }

// PatchFunc registers a standard http.HandlerFunc for PATCH requests.
func (r *Router) PatchFunc(pattern string, fn http.HandlerFunc) { r.group.PatchFunc(pattern, fn) }

// Delete registers an error-returning handler for DELETE requests.
func (r *Router) Delete(pattern string, fn HandlerFunc) { r.group.Delete(pattern, fn) }

// DeleteFunc registers a standard http.HandlerFunc for DELETE requests.
func (r *Router) DeleteFunc(pattern string, fn http.HandlerFunc) { r.group.DeleteFunc(pattern, fn) }

// Handle registers an http.Handler for the given pattern.
func (r *Router) Handle(pattern string, handler http.Handler) { r.group.Handle(pattern, handler) }

// HandleFunc registers an http.HandlerFunc for the given pattern.
func (r *Router) HandleFunc(pattern string, fn http.HandlerFunc) { r.group.HandleFunc(pattern, fn) }

// Use adds middleware to the root group.
func (r *Router) Use(mw ...middleware.Middleware) { r.group.Use(mw...) }

// Group creates a new route group with the given prefix and optional middleware.
func (r *Router) Group(prefix string, mw ...middleware.Middleware) *Group {
	return r.group.Group(prefix, mw...)
}

// wrapError wraps an error-returning HandlerFunc into a standard http.HandlerFunc.
func (r *Router) wrapError(fn HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if err := fn(w, req); err != nil {
			r.errorHandler(w, req, err)
		}
	}
}

// errorEnvelope is the JSON structure written by DefaultErrorHandler.
type errorEnvelope struct {
	Success   bool       `json:"success"`
	Error     *errorBody `json:"error"`
	Timestamp int64      `json:"timestamp"`
}

// errorBody is the error portion of the envelope.
type errorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// DefaultErrorHandler writes a JSON error response matching the standard envelope format.
// It uses errors.As to extract *errors.Error; unrecognized errors become 500 Internal Server Error.
func DefaultErrorHandler(w http.ResponseWriter, _ *http.Request, err error) {
	var apiErr *errors.Error
	code := http.StatusInternalServerError
	errCode := "INTERNAL_ERROR"
	message := "An internal error occurred"
	var fields map[string]string

	if stderrors.As(err, &apiErr) {
		code = apiErr.StatusCode
		errCode = apiErr.Code
		message = apiErr.Message
		fields = apiErr.Fields
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)

	_ = json.NewEncoder(w).Encode(errorEnvelope{
		Success: false,
		Error: &errorBody{
			Code:    errCode,
			Message: message,
			Fields:  fields,
		},
		Timestamp: time.Now().Unix(),
	})
}
