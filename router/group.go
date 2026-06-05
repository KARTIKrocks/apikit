package router

import (
	"fmt"
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
func (g *Group) Get(pattern string, fn HandlerFunc) *RouteEntry {
	return g.register("GET", pattern, g.router.wrapError(fn), fn)
}

// GetFunc registers a standard http.HandlerFunc for GET requests.
func (g *Group) GetFunc(pattern string, fn http.HandlerFunc) *RouteEntry {
	return g.register("GET", pattern, fn, fn)
}

// Post registers an error-returning handler for POST requests.
func (g *Group) Post(pattern string, fn HandlerFunc) *RouteEntry {
	return g.register("POST", pattern, g.router.wrapError(fn), fn)
}

// PostFunc registers a standard http.HandlerFunc for POST requests.
func (g *Group) PostFunc(pattern string, fn http.HandlerFunc) *RouteEntry {
	return g.register("POST", pattern, fn, fn)
}

// Put registers an error-returning handler for PUT requests.
func (g *Group) Put(pattern string, fn HandlerFunc) *RouteEntry {
	return g.register("PUT", pattern, g.router.wrapError(fn), fn)
}

// PutFunc registers a standard http.HandlerFunc for PUT requests.
func (g *Group) PutFunc(pattern string, fn http.HandlerFunc) *RouteEntry {
	return g.register("PUT", pattern, fn, fn)
}

// Patch registers an error-returning handler for PATCH requests.
func (g *Group) Patch(pattern string, fn HandlerFunc) *RouteEntry {
	return g.register("PATCH", pattern, g.router.wrapError(fn), fn)
}

// PatchFunc registers a standard http.HandlerFunc for PATCH requests.
func (g *Group) PatchFunc(pattern string, fn http.HandlerFunc) *RouteEntry {
	return g.register("PATCH", pattern, fn, fn)
}

// Delete registers an error-returning handler for DELETE requests.
func (g *Group) Delete(pattern string, fn HandlerFunc) *RouteEntry {
	return g.register("DELETE", pattern, g.router.wrapError(fn), fn)
}

// DeleteFunc registers a standard http.HandlerFunc for DELETE requests.
func (g *Group) DeleteFunc(pattern string, fn http.HandlerFunc) *RouteEntry {
	return g.register("DELETE", pattern, fn, fn)
}

// Head registers an error-returning handler for HEAD requests.
func (g *Group) Head(pattern string, fn HandlerFunc) *RouteEntry {
	return g.register("HEAD", pattern, g.router.wrapError(fn), fn)
}

// HeadFunc registers a standard http.HandlerFunc for HEAD requests.
func (g *Group) HeadFunc(pattern string, fn http.HandlerFunc) *RouteEntry {
	return g.register("HEAD", pattern, fn, fn)
}

// Options registers an error-returning handler for OPTIONS requests.
func (g *Group) Options(pattern string, fn HandlerFunc) *RouteEntry {
	return g.register("OPTIONS", pattern, g.router.wrapError(fn), fn)
}

// OptionsFunc registers a standard http.HandlerFunc for OPTIONS requests.
func (g *Group) OptionsFunc(pattern string, fn http.HandlerFunc) *RouteEntry {
	return g.register("OPTIONS", pattern, fn, fn)
}

// Handle registers an http.Handler for the given pattern.
// The pattern may include a method prefix (e.g. "GET /path").
func (g *Group) Handle(pattern string, handler http.Handler) *RouteEntry {
	method, path := splitPattern(pattern)
	prefix, chain := g.resolve()
	fullPath := joinPath(prefix, path)
	fullPattern := fullPath
	if method != "" {
		fullPattern = method + " " + fullPath
	}

	origHandler := handler
	if len(chain) > 0 {
		handler = middleware.Chain(chain...)(handler)
	}
	g.router.mux.Handle(fullPattern, markMatched(handler))

	idx := len(g.router.routes)
	g.router.routes = append(g.router.routes, RouteInfo{
		Method:      method,
		Pattern:     fullPath,
		HandlerName: handlerName(origHandler),
	})
	return &RouteEntry{router: g.router, index: idx}
}

// HandleFunc registers an http.HandlerFunc for the given pattern.
// The pattern may include a method prefix (e.g. "GET /path").
func (g *Group) HandleFunc(pattern string, fn http.HandlerFunc) *RouteEntry {
	return g.Handle(pattern, fn)
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

// Route creates a child group with the given prefix and optional middleware,
// then calls fn to register routes on it. This is syntactic sugar for inline sub-routing:
//
//	r.Route("/users", func(sub *Group) {
//	    sub.Get("/", listUsers)
//	    sub.Get("/{id}", getUser)
//	})
func (g *Group) Route(prefix string, fn func(*Group), mw ...middleware.Middleware) *Group {
	sub := g.Group(prefix, mw...)
	fn(sub)
	return sub
}

// Static serves files from the given filesystem directory under the URL prefix.
// Group middleware is applied to all requests.
//
//	r.Static("/assets", "./public")
func (g *Group) Static(prefix, dir string) {
	groupPrefix, chain := g.resolve()
	fullPrefix := joinPath(groupPrefix, prefix)
	fs := http.StripPrefix(fullPrefix, http.FileServer(http.Dir(dir)))
	if len(chain) > 0 {
		fs = middleware.Chain(chain...)(fs)
	}

	g.router.mux.Handle(fullPrefix+"/", markMatched(fs))

	g.router.routes = append(g.router.routes, RouteInfo{
		Method:  "GET",
		Pattern: fullPrefix + "/{file...}",
	})
}

// File registers a handler that serves a single file for GET requests.
//
//	r.File("/favicon.ico", "./public/favicon.ico")
func (g *Group) File(pattern, filePath string) {
	g.register("GET", pattern, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filePath)
	}), nil)
}

// Mount attaches an http.Handler at the given prefix, stripping the prefix
// from the request path before passing it to the handler. Group middleware is applied.
// If the handler is a *Router, its routes are merged into the parent's route table for introspection.
//
//	admin := router.New()
//	admin.Get("/stats", statsHandler)
//	r.Mount("/admin", admin)
func (g *Group) Mount(prefix string, handler http.Handler) {
	groupPrefix, chain := g.resolve()
	fullPrefix := joinPath(groupPrefix, prefix)

	// Order: markMatched → middleware → StripPrefix → handler
	// This ensures parent middleware sees the full path (consistent with Static/regular routes)
	// and markMatched always runs regardless of StripPrefix outcome.
	h := http.StripPrefix(fullPrefix, handler)
	if len(chain) > 0 {
		h = middleware.Chain(chain...)(h)
	}

	g.router.mux.Handle(fullPrefix+"/", markMatched(h))
	g.router.mux.Handle(fullPrefix, markMatched(h))

	// Merge sub-router routes for introspection; otherwise record a single mount entry.
	if sub, ok := handler.(*Router); ok {
		for _, ri := range sub.routes {
			idx := len(g.router.routes)
			g.router.routes = append(g.router.routes, RouteInfo{
				Method:      ri.Method,
				Pattern:     joinPath(fullPrefix, ri.Pattern),
				Name:        ri.Name,
				HandlerName: ri.HandlerName,
			})
			if ri.Name != "" {
				if _, exists := g.router.namedRoutes[ri.Name]; exists {
					panic(fmt.Sprintf("router: duplicate route name %q (from mounted sub-router)", ri.Name))
				}
				g.router.namedRoutes[ri.Name] = idx
			}
		}
	} else {
		g.router.routes = append(g.router.routes, RouteInfo{
			Pattern:     fullPrefix + "/",
			HandlerName: handlerName(handler),
		})
	}
}

// With returns a child group that shares this group's prefix but adds
// the given middleware. It is intended for per-route middleware:
//
//	r.With(authMW).Get("/admin", adminHandler)
func (g *Group) With(mw ...middleware.Middleware) *Group {
	return &Group{
		prefix:      "",
		middlewares: mw,
		router:      g.router,
		parent:      g,
	}
}

// register builds the full pattern, registers the handler on the mux, and records the route.
func (g *Group) register(method, pattern string, handler http.Handler, origFn any) *RouteEntry {
	prefix, chain := g.resolve()
	fullPath := joinPath(prefix, pattern)
	fullPattern := method + " " + fullPath
	if len(chain) > 0 {
		handler = middleware.Chain(chain...)(handler)
	}
	g.router.mux.Handle(fullPattern, markMatched(handler))

	idx := len(g.router.routes)
	g.router.routes = append(g.router.routes, RouteInfo{
		Method:      method,
		Pattern:     fullPath,
		HandlerName: handlerName(origFn),
	})
	return &RouteEntry{router: g.router, index: idx}
}

// resolve walks the parent chain once and returns both the full prefix
// and the accumulated middleware slice (root → current order).
// It pre-sizes slices and tracks the last written byte to avoid
// intermediate string allocations.
func (g *Group) resolve() (prefix string, mws []middleware.Middleware) {
	// Count depth for pre-sizing.
	depth := 0
	for cur := g; cur != nil; cur = cur.parent {
		depth++
	}

	// Collect groups in root → current order.
	groups := make([]*Group, depth)
	i := depth - 1
	for cur := g; cur != nil; cur = cur.parent {
		groups[i] = cur
		i--
	}

	// Build prefix, tracking last byte to avoid b.String() in the loop.
	var b strings.Builder
	var lastByte byte
	for _, grp := range groups {
		p := grp.prefix
		if len(p) == 0 {
			continue
		}
		if b.Len() > 0 && lastByte == '/' && p[0] == '/' {
			p = p[1:]
		}
		if len(p) > 0 {
			b.WriteString(p)
			lastByte = p[len(p)-1]
		}
	}
	prefix = b.String()

	// Collect middleware in root → current order.
	for _, grp := range groups {
		mws = append(mws, grp.middlewares...)
	}
	return
}

// joinPath concatenates a prefix and path, normalizing double slashes at the join point.
// joinPath("/api/", "/users") → "/api/users"
// joinPath("/api", "/users")  → "/api/users"
// joinPath("/api", "users")   → "/apiusers" (caller should ensure leading slash on path)
func joinPath(prefix, path string) string {
	if strings.HasSuffix(prefix, "/") && strings.HasPrefix(path, "/") {
		return prefix + path[1:]
	}
	return prefix + path
}

// splitPattern separates a Go 1.22 pattern into method and path parts.
// "GET /users" → ("GET", "/users"), "/users" → ("", "/users")
// It trims extra whitespace between method and path.
func splitPattern(pattern string) (method, path string) {
	method, path, found := strings.Cut(pattern, " ")
	if !found {
		return "", pattern
	}
	path = strings.TrimLeft(path, " ") // handle extra spaces: "GET  /users"
	return method, path
}
