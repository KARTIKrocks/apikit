package router

import (
	"fmt"
	"strings"
)

// URL builds a URL path for the named route, substituting path parameters.
// Parameters are provided as key-value pairs: r.URL("get-user", "id", "42") returns "/users/42".
//
// It panics if the route name is not found, if the number of params is odd,
// if a {placeholder} in the pattern has no matching key in params,
// or if extra params are provided that don't match any placeholder.
func (r *Router) URL(name string, params ...string) string {
	idx, ok := r.namedRoutes[name]
	if !ok {
		panic(fmt.Sprintf("router: unknown route name %q", name))
	}
	if len(params)%2 != 0 {
		panic("router: URL params must be key-value pairs")
	}

	pattern := r.routes[idx].Pattern

	// Build param map.
	m := make(map[string]string, len(params)/2)
	for i := 0; i < len(params); i += 2 {
		m[params[i]] = params[i+1]
	}

	// Replace {name} and {name...} placeholders, tracking which params are used.
	used := 0
	var b strings.Builder
	b.Grow(len(pattern))
	for i := 0; i < len(pattern); {
		if pattern[i] == '{' {
			end := strings.IndexByte(pattern[i:], '}')
			if end == -1 {
				b.WriteByte(pattern[i])
				i++
				continue
			}
			placeholder := pattern[i+1 : i+end]
			key := strings.TrimSuffix(placeholder, "...")
			val, found := m[key]
			if !found {
				panic(fmt.Sprintf("router: missing param %q for route %q", key, name))
			}
			used++
			b.WriteString(val)
			i += end + 1
		} else {
			b.WriteByte(pattern[i])
			i++
		}
	}

	if used != len(m) {
		panic(fmt.Sprintf("router: %d extra param(s) provided for route %q", len(m)-used, name))
	}

	return b.String()
}
