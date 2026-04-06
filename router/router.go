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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/KARTIKrocks/apikit/errors"
	"github.com/KARTIKrocks/apikit/middleware"
)

var probeWriterPool = sync.Pool{
	New: func() any { return &probeWriter{} },
}

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
	mux                     *http.ServeMux
	group                   Group
	errorHandler            ErrorHandler
	notFoundHandler         http.Handler
	methodNotAllowedHandler http.Handler
	stripSlash              bool
	redirectSlash           bool
	routes                  []RouteInfo
	namedRoutes             map[string]int // name → index into routes
}

// New creates a new Router with the given options.
func New(opts ...Option) *Router {
	r := &Router{
		mux:          http.NewServeMux(),
		errorHandler: DefaultErrorHandler,
		namedRoutes:  make(map[string]int),
	}
	r.group = Group{
		router: r,
	}
	for _, opt := range opts {
		opt(r)
	}
	if r.stripSlash && r.redirectSlash {
		panic("router: WithStripSlash and WithRedirectSlash are mutually exclusive")
	}
	return r
}

// WithErrorHandler sets a custom error handler for the router.
func WithErrorHandler(fn ErrorHandler) Option {
	return func(r *Router) {
		r.errorHandler = fn
	}
}

// WithNotFound sets a custom handler for 404 Not Found responses.
// When set, this handler is called instead of the ErrorHandler for unmatched routes.
func WithNotFound(handler http.Handler) Option {
	return func(r *Router) {
		r.notFoundHandler = handler
	}
}

// WithMethodNotAllowed sets a custom handler for 405 Method Not Allowed responses.
// When set, this handler is called instead of the ErrorHandler for disallowed methods.
func WithMethodNotAllowed(handler http.Handler) Option {
	return func(r *Router) {
		r.methodNotAllowedHandler = handler
	}
}

// WithStripSlash silently removes trailing slashes from request paths before routing.
// "/users/" becomes "/users". The root path "/" is never modified.
func WithStripSlash() Option {
	return func(r *Router) {
		r.stripSlash = true
	}
}

// WithRedirectSlash sends a 301 Moved Permanently redirect for requests with a trailing slash.
// "/users/" redirects to "/users". The root path "/" is never redirected.
// Mutually exclusive with WithStripSlash.
func WithRedirectSlash() Option {
	return func(r *Router) {
		r.redirectSlash = true
	}
}

// ServeHTTP implements http.Handler.
// It intercepts 404 and 405 responses from the underlying ServeMux
// and routes them through the router's ErrorHandler for consistent error format.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Handle trailing slashes before routing.
	path := req.URL.Path
	if path != "/" && strings.HasSuffix(path, "/") {
		trimmed := strings.TrimRight(path, "/")
		if trimmed == "" {
			trimmed = "/"
		}
		if r.redirectSlash {
			target := trimmed
			if req.URL.RawQuery != "" {
				target += "?" + req.URL.RawQuery
			}
			http.Redirect(w, req, target, http.StatusMovedPermanently)
			return
		}
		if r.stripSlash {
			req.URL.Path = trimmed
		}
	}

	// Use a pooled probe writer to detect 404/405 from ServeMux before writing to the real response.
	pw := probeWriterPool.Get().(*probeWriter)
	pw.ResponseWriter = w
	pw.code = 0
	pw.intercepted = false
	pw.wroteBody = false
	pw.matched = false

	r.mux.ServeHTTP(pw, req)

	if !pw.intercepted {
		probeWriterPool.Put(pw)
		return
	}

	// ServeMux returned 404 or 405 — use dedicated handler if set, otherwise ErrorHandler.
	switch pw.code {
	case http.StatusNotFound:
		if r.notFoundHandler != nil {
			r.notFoundHandler.ServeHTTP(w, req)
		} else {
			r.errorHandler(w, req, errors.NotFound(""))
		}
	case http.StatusMethodNotAllowed:
		if r.methodNotAllowedHandler != nil {
			r.methodNotAllowedHandler.ServeHTTP(w, req)
		} else {
			r.errorHandler(w, req, &errors.Error{
				StatusCode: http.StatusMethodNotAllowed,
				Code:       errors.CodeMethodNotAllowed,
				Message:    "Method not allowed",
			})
		}
	}
	probeWriterPool.Put(pw)
}

// probeWriter intercepts WriteHeader calls to detect 404/405 from ServeMux.
// If the status is 404 or 405 and no user handler has been matched,
// it suppresses the write so the router's ErrorHandler can produce a consistent response.
//
// Registered handlers set pw.matched = true via a thin wrapper so that intentional
// 404/405 responses from user handlers are forwarded correctly.
type probeWriter struct {
	http.ResponseWriter
	code        int
	intercepted bool
	wroteBody   bool
	matched     bool // true when a registered handler was matched by ServeMux
}

func (pw *probeWriter) WriteHeader(code int) {
	// Only intercept unmatched routes (ServeMux's own 404/405).
	// If a registered handler was matched, forward the status as-is.
	if !pw.matched && !pw.wroteBody && (code == http.StatusNotFound || code == http.StatusMethodNotAllowed) {
		pw.code = code
		pw.intercepted = true
		return
	}
	pw.ResponseWriter.WriteHeader(code)
}

func (pw *probeWriter) Write(b []byte) (int, error) {
	if pw.intercepted {
		// Suppress the body written by ServeMux's default handler.
		return len(b), nil
	}
	pw.wroteBody = true
	return pw.ResponseWriter.Write(b)
}

// Unwrap returns the underlying ResponseWriter so that http.ResponseController
// can reach interfaces like http.Flusher and http.Hijacker through the wrapper.
func (pw *probeWriter) Unwrap() http.ResponseWriter { return pw.ResponseWriter }

// markMatched wraps a handler to set the matched flag on the probeWriter.
// This lets probeWriter distinguish ServeMux's default 404/405 from
// intentional 404/405 responses returned by user handlers.
func markMatched(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if pw, ok := w.(*probeWriter); ok {
			pw.matched = true
		}
		h.ServeHTTP(w, r)
	})
}

// Get registers an error-returning handler for GET requests.
func (r *Router) Get(pattern string, fn HandlerFunc) *RouteEntry { return r.group.Get(pattern, fn) }

// GetFunc registers a standard http.HandlerFunc for GET requests.
func (r *Router) GetFunc(pattern string, fn http.HandlerFunc) *RouteEntry {
	return r.group.GetFunc(pattern, fn)
}

// Post registers an error-returning handler for POST requests.
func (r *Router) Post(pattern string, fn HandlerFunc) *RouteEntry { return r.group.Post(pattern, fn) }

// PostFunc registers a standard http.HandlerFunc for POST requests.
func (r *Router) PostFunc(pattern string, fn http.HandlerFunc) *RouteEntry {
	return r.group.PostFunc(pattern, fn)
}

// Put registers an error-returning handler for PUT requests.
func (r *Router) Put(pattern string, fn HandlerFunc) *RouteEntry { return r.group.Put(pattern, fn) }

// PutFunc registers a standard http.HandlerFunc for PUT requests.
func (r *Router) PutFunc(pattern string, fn http.HandlerFunc) *RouteEntry {
	return r.group.PutFunc(pattern, fn)
}

// Patch registers an error-returning handler for PATCH requests.
func (r *Router) Patch(pattern string, fn HandlerFunc) *RouteEntry {
	return r.group.Patch(pattern, fn)
}

// PatchFunc registers a standard http.HandlerFunc for PATCH requests.
func (r *Router) PatchFunc(pattern string, fn http.HandlerFunc) *RouteEntry {
	return r.group.PatchFunc(pattern, fn)
}

// Delete registers an error-returning handler for DELETE requests.
func (r *Router) Delete(pattern string, fn HandlerFunc) *RouteEntry {
	return r.group.Delete(pattern, fn)
}

// DeleteFunc registers a standard http.HandlerFunc for DELETE requests.
func (r *Router) DeleteFunc(pattern string, fn http.HandlerFunc) *RouteEntry {
	return r.group.DeleteFunc(pattern, fn)
}

// Head registers an error-returning handler for HEAD requests.
func (r *Router) Head(pattern string, fn HandlerFunc) *RouteEntry { return r.group.Head(pattern, fn) }

// HeadFunc registers a standard http.HandlerFunc for HEAD requests.
func (r *Router) HeadFunc(pattern string, fn http.HandlerFunc) *RouteEntry {
	return r.group.HeadFunc(pattern, fn)
}

// Options registers an error-returning handler for OPTIONS requests.
func (r *Router) Options(pattern string, fn HandlerFunc) *RouteEntry {
	return r.group.Options(pattern, fn)
}

// OptionsFunc registers a standard http.HandlerFunc for OPTIONS requests.
func (r *Router) OptionsFunc(pattern string, fn http.HandlerFunc) *RouteEntry {
	return r.group.OptionsFunc(pattern, fn)
}

// Handle registers an http.Handler for the given pattern.
func (r *Router) Handle(pattern string, handler http.Handler) *RouteEntry {
	return r.group.Handle(pattern, handler)
}

// HandleFunc registers an http.HandlerFunc for the given pattern.
func (r *Router) HandleFunc(pattern string, fn http.HandlerFunc) *RouteEntry {
	return r.group.HandleFunc(pattern, fn)
}

// Use adds middleware to the root group.
func (r *Router) Use(mw ...middleware.Middleware) { r.group.Use(mw...) }

// With returns a group that shares the root prefix but adds
// the given middleware for the next registered route(s).
func (r *Router) With(mw ...middleware.Middleware) *Group { return r.group.With(mw...) }

// Route creates a sub-group with the given prefix and calls fn to register routes on it.
func (r *Router) Route(prefix string, fn func(*Group), mw ...middleware.Middleware) *Group {
	return r.group.Route(prefix, fn, mw...)
}

// Mount attaches an http.Handler at the given prefix.
func (r *Router) Mount(prefix string, handler http.Handler) { r.group.Mount(prefix, handler) }

// Static serves files from the given directory under the URL prefix.
func (r *Router) Static(prefix, dir string) { r.group.Static(prefix, dir) }

// File registers a handler that serves a single file for GET requests.
func (r *Router) File(pattern, filePath string) { r.group.File(pattern, filePath) }

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

	body, err := json.Marshal(errorEnvelope{
		Success: false,
		Error: &errorBody{
			Code:    errCode,
			Message: message,
			Fields:  fields,
		},
		Timestamp: time.Now().Unix(),
	})
	if err != nil {
		http.Error(w, `{"success":false,"error":{"code":"INTERNAL_ERROR","message":"An internal error occurred"}}`, http.StatusInternalServerError)
		return
	}
	body = append(body, '\n')

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(code)
	_, _ = w.Write(body)
}
