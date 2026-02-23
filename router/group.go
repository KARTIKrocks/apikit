package router

import (
	"net/http"
	"strings"

	"github.com/KARTIKrocks/apikit/middleware"
)

// Group represents a collection of routes that share a common prefix and middleware.
type Group struct {
	prefix      string
	middlewares []middleware.Middleware
	router      *Router
	parent      *Group
}

// Get registers an error-returning handler for GET requests.
func (g *Group) Get(pattern string, fn HandlerFunc) {
	g.register("GET", pattern, g.router.wrapError(fn))
}

// GetFunc registers a standard http.HandlerFunc for GET requests.
func (g *Group) GetFunc(pattern string, fn http.HandlerFunc) {
	g.register("GET", pattern, fn)
}

// Post registers an error-returning handler for POST requests.
func (g *Group) Post(pattern string, fn HandlerFunc) {
	g.register("POST", pattern, g.router.wrapError(fn))
}

// PostFunc registers a standard http.HandlerFunc for POST requests.
func (g *Group) PostFunc(pattern string, fn http.HandlerFunc) {
	g.register("POST", pattern, fn)
}

// Put registers an error-returning handler for PUT requests.
func (g *Group) Put(pattern string, fn HandlerFunc) {
	g.register("PUT", pattern, g.router.wrapError(fn))
}

// PutFunc registers a standard http.HandlerFunc for PUT requests.
func (g *Group) PutFunc(pattern string, fn http.HandlerFunc) {
	g.register("PUT", pattern, fn)
}

// Patch registers an error-returning handler for PATCH requests.
func (g *Group) Patch(pattern string, fn HandlerFunc) {
	g.register("PATCH", pattern, g.router.wrapError(fn))
}

// PatchFunc registers a standard http.HandlerFunc for PATCH requests.
func (g *Group) PatchFunc(pattern string, fn http.HandlerFunc) {
	g.register("PATCH", pattern, fn)
}

// Delete registers an error-returning handler for DELETE requests.
func (g *Group) Delete(pattern string, fn HandlerFunc) {
	g.register("DELETE", pattern, g.router.wrapError(fn))
}

// DeleteFunc registers a standard http.HandlerFunc for DELETE requests.
func (g *Group) DeleteFunc(pattern string, fn http.HandlerFunc) {
	g.register("DELETE", pattern, fn)
}

// Handle registers an http.Handler for the given pattern.
// The pattern may include a method prefix (e.g. "GET /path").
func (g *Group) Handle(pattern string, handler http.Handler) {
	method, path := splitPattern(pattern)
	fullPath := g.fullPrefix() + path
	fullPattern := fullPath
	if method != "" {
		fullPattern = method + " " + fullPath
	}

	chain := g.collectMiddleware()
	if len(chain) > 0 {
		handler = middleware.Chain(chain...)(handler)
	}
	g.router.mux.Handle(fullPattern, handler)
}

// HandleFunc registers an http.HandlerFunc for the given pattern.
// The pattern may include a method prefix (e.g. "GET /path").
func (g *Group) HandleFunc(pattern string, fn http.HandlerFunc) {
	g.Handle(pattern, fn)
}

// Use appends middleware to this group. Middleware added via Use only
// applies to routes registered after the call.
func (g *Group) Use(mw ...middleware.Middleware) {
	g.middlewares = append(g.middlewares, mw...)
}

// Group creates a child group with the given prefix and optional middleware.
func (g *Group) Group(prefix string, mw ...middleware.Middleware) *Group {
	return &Group{
		prefix:      prefix,
		middlewares: mw,
		router:      g.router,
		parent:      g,
	}
}

// register builds the full pattern and registers the handler on the mux.
func (g *Group) register(method, pattern string, handler http.Handler) {
	fullPath := g.fullPrefix() + pattern
	fullPattern := method + " " + fullPath

	chain := g.collectMiddleware()
	if len(chain) > 0 {
		handler = middleware.Chain(chain...)(handler)
	}
	g.router.mux.Handle(fullPattern, handler)
}

// collectMiddleware walks the parent chain from root to current group
// and returns the accumulated middleware slice in order.
func (g *Group) collectMiddleware() []middleware.Middleware {
	// Build parent chain (current → root).
	var groups []*Group
	for cur := g; cur != nil; cur = cur.parent {
		groups = append(groups, cur)
	}

	// Reverse to get root → current order.
	var mws []middleware.Middleware
	for i := len(groups) - 1; i >= 0; i-- {
		mws = append(mws, groups[i].middlewares...)
	}
	return mws
}

// fullPrefix returns the concatenated prefix from root to this group.
func (g *Group) fullPrefix() string {
	var groups []*Group
	for cur := g; cur != nil; cur = cur.parent {
		groups = append(groups, cur)
	}

	var b strings.Builder
	for i := len(groups) - 1; i >= 0; i-- {
		b.WriteString(groups[i].prefix)
	}
	return b.String()
}

// splitPattern separates a Go 1.22 pattern into method and path parts.
// "GET /users" → ("GET", "/users"), "/users" → ("", "/users")
func splitPattern(pattern string) (method, path string) {
	if i := strings.IndexByte(pattern, ' '); i != -1 {
		return pattern[:i], pattern[i+1:]
	}
	return "", pattern
}
