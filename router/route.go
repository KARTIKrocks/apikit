package router

import (
	"fmt"
	"reflect"
	"runtime"
)

// RouteInfo holds metadata about a registered route.
type RouteInfo struct {
	Method      string // HTTP method ("GET", "POST", etc.). Empty for method-agnostic Handle() routes.
	Pattern     string // Full path pattern as registered, e.g. "/api/v1/users/{id}".
	Name        string // Optional name set via RouteEntry.Name(), used for URL generation.
	HandlerName string // Runtime function name of the original handler.
}

// RouteEntry is returned by route registration methods to allow optional chaining.
// Callers that ignore the return value get the same behavior as before.
type RouteEntry struct {
	router *Router
	index  int
}

// Name assigns a name to this route for URL generation.
// It panics if the name is already taken (fail-fast, like http.ServeMux on duplicate patterns).
func (re *RouteEntry) Name(name string) *RouteEntry {
	if _, exists := re.router.namedRoutes[name]; exists {
		panic(fmt.Sprintf("router: duplicate route name %q", name))
	}
	re.router.routes[re.index].Name = name
	re.router.namedRoutes[name] = re.index
	return re
}

// handlerName extracts a human-readable function name from a handler using runtime reflection.
func handlerName(fn any) string {
	if fn == nil {
		return ""
	}
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		return ""
	}
	pc := v.Pointer()
	f := runtime.FuncForPC(pc)
	if f == nil {
		return ""
	}
	return f.Name()
}
