package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ─── Route Introspection ────────────────────────────────────────────────────

func TestRoutesEmpty(t *testing.T) {
	r := New()
	if got := len(r.Routes()); got != 0 {
		t.Errorf("expected 0 routes, got %d", got)
	}
}

func TestRoutesBasic(t *testing.T) {
	r := New()
	r.Get("/users", noopHandler)
	r.Post("/users", noopHandler)
	r.Put("/users/{id}", noopHandler)

	routes := r.Routes()
	if len(routes) != 3 {
		t.Fatalf("expected 3 routes, got %d", len(routes))
	}

	want := []struct{ method, pattern string }{
		{"GET", "/users"},
		{"POST", "/users"},
		{"PUT", "/users/{id}"},
	}
	for i, w := range want {
		if routes[i].Method != w.method {
			t.Errorf("route[%d] method: expected %q, got %q", i, w.method, routes[i].Method)
		}
		if routes[i].Pattern != w.pattern {
			t.Errorf("route[%d] pattern: expected %q, got %q", i, w.pattern, routes[i].Pattern)
		}
	}
}

func TestRoutesWithGroups(t *testing.T) {
	r := New()
	api := r.Group("/api")
	v1 := api.Group("/v1")
	v1.Get("/items", noopHandler)

	routes := r.Routes()
	if len(routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(routes))
	}
	if routes[0].Pattern != "/api/v1/items" {
		t.Errorf("expected /api/v1/items, got %q", routes[0].Pattern)
	}
}

func TestRoutesCopy(t *testing.T) {
	r := New()
	r.Get("/a", noopHandler)

	routes := r.Routes()
	routes[0].Method = "MODIFIED"

	// Original should be unchanged.
	if r.Routes()[0].Method != "GET" {
		t.Error("modifying returned slice should not affect router")
	}
}

func TestWalk(t *testing.T) {
	r := New()
	r.Get("/a", noopHandler)
	r.Post("/b", noopHandler)
	r.Delete("/c", noopHandler)

	var collected []string
	err := r.Walk(func(ri RouteInfo) error {
		collected = append(collected, ri.Method+" "+ri.Pattern)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(collected) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(collected))
	}
	if collected[0] != "GET /a" {
		t.Errorf("first: expected %q, got %q", "GET /a", collected[0])
	}
}

func TestWalkEarlyExit(t *testing.T) {
	r := New()
	r.Get("/a", noopHandler)
	r.Get("/b", noopHandler)
	r.Get("/c", noopHandler)

	stopErr := fmt.Errorf("stop")
	count := 0
	err := r.Walk(func(ri RouteInfo) error {
		count++
		if count == 2 {
			return stopErr
		}
		return nil
	})
	if err != stopErr {
		t.Errorf("expected stop error, got %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 iterations, got %d", count)
	}
}

func TestRouteInfoHandlerName(t *testing.T) {
	r := New()
	r.Get("/test", namedTestHandler)

	routes := r.Routes()
	if !strings.Contains(routes[0].HandlerName, "namedTestHandler") {
		t.Errorf("expected handler name to contain 'namedTestHandler', got %q", routes[0].HandlerName)
	}
}

func TestHandleHandlerNameWithMiddleware(t *testing.T) {
	r := New()
	api := r.Group("/api", headerMiddleware("X-Test", "yes"))
	api.Handle("GET /data", http.HandlerFunc(namedStdlibHandler))

	routes := r.Routes()
	if strings.Contains(routes[0].HandlerName, "Chain") {
		t.Errorf("handler name should be the original, not middleware wrapper, got %q", routes[0].HandlerName)
	}
}

func TestProbeWriterUnwrap(t *testing.T) {
	r := New()
	r.Get("/flush", func(w http.ResponseWriter, req *http.Request) error {
		rc := http.NewResponseController(w)
		if err := rc.Flush(); err != nil {
			return fmt.Errorf("flush failed: %w", err)
		}
		fmt.Fprint(w, "flushed")
		return nil
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/flush", nil)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if !rec.Flushed {
		t.Error("expected response to be flushed")
	}
}

func TestRoutesHandle(t *testing.T) {
	r := New()
	r.Handle("/catch-all", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {}))

	routes := r.Routes()
	if len(routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(routes))
	}
	if routes[0].Method != "" {
		t.Errorf("expected empty method for method-agnostic Handle, got %q", routes[0].Method)
	}
}

func TestRoutesHandleWithMethod(t *testing.T) {
	r := New()
	r.Handle("GET /explicit", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {}))

	routes := r.Routes()
	if routes[0].Method != "GET" {
		t.Errorf("expected method GET, got %q", routes[0].Method)
	}
	if routes[0].Pattern != "/explicit" {
		t.Errorf("expected pattern /explicit, got %q", routes[0].Pattern)
	}
}

func TestRoutesFuncHelpers(t *testing.T) {
	r := New()
	r.GetFunc("/a", func(w http.ResponseWriter, req *http.Request) {})
	r.PostFunc("/b", func(w http.ResponseWriter, req *http.Request) {})

	routes := r.Routes()
	if len(routes) != 2 {
		t.Fatalf("expected 2 routes, got %d", len(routes))
	}
	if routes[0].Method != "GET" || routes[1].Method != "POST" {
		t.Error("unexpected methods for Func helpers")
	}
}

// ─── Named Routes ───────────────────────────────────────────────────────────

func TestNamedRouteBasic(t *testing.T) {
	r := New()
	r.Get("/users", noopHandler).Name("list-users")

	routes := r.Routes()
	if routes[0].Name != "list-users" {
		t.Errorf("expected name 'list-users', got %q", routes[0].Name)
	}
}

func TestNamedRouteDuplicatePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic on duplicate route name")
		}
	}()

	r := New()
	r.Get("/a", noopHandler).Name("dup")
	r.Get("/b", noopHandler).Name("dup")
}

func TestNamedRouteOnGroup(t *testing.T) {
	r := New()
	api := r.Group("/api")
	api.Get("/users/{id}", noopHandler).Name("get-user")

	routes := r.Routes()
	if routes[0].Name != "get-user" {
		t.Errorf("expected name 'get-user', got %q", routes[0].Name)
	}
}

func TestNamedRouteChaining(t *testing.T) {
	r := New()
	entry := r.Get("/test", noopHandler).Name("test-route")
	if entry == nil {
		t.Error("Name() should return non-nil RouteEntry")
	}
}

// ─── URL Generation ─────────────────────────────────────────────────────────

func TestURLBasic(t *testing.T) {
	r := New()
	r.Get("/users/{id}", noopHandler).Name("get-user")

	got := r.URL("get-user", "id", "42")
	if got != "/users/42" {
		t.Errorf("expected /users/42, got %q", got)
	}
}

func TestURLMultipleParams(t *testing.T) {
	r := New()
	r.Get("/users/{userID}/posts/{postID}", noopHandler).Name("get-post")

	got := r.URL("get-post", "userID", "1", "postID", "99")
	if got != "/users/1/posts/99" {
		t.Errorf("expected /users/1/posts/99, got %q", got)
	}
}

func TestURLNoParams(t *testing.T) {
	r := New()
	r.Get("/health", noopHandler).Name("health")

	got := r.URL("health")
	if got != "/health" {
		t.Errorf("expected /health, got %q", got)
	}
}

func TestURLCatchAll(t *testing.T) {
	r := New()
	r.Get("/files/{path...}", noopHandler).Name("files")

	got := r.URL("files", "path", "docs/readme.md")
	if got != "/files/docs/readme.md" {
		t.Errorf("expected /files/docs/readme.md, got %q", got)
	}
}

func TestURLWithGroup(t *testing.T) {
	r := New()
	api := r.Group("/api/v1")
	api.Get("/users/{id}", noopHandler).Name("api-user")

	got := r.URL("api-user", "id", "5")
	if got != "/api/v1/users/5" {
		t.Errorf("expected /api/v1/users/5, got %q", got)
	}
}

func TestURLUnknownNamePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic for unknown route name")
		}
	}()

	r := New()
	r.URL("nonexistent")
}

func TestURLOddParamsPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic for odd param count")
		}
	}()

	r := New()
	r.Get("/test/{id}", noopHandler).Name("test")
	r.URL("test", "id")
}

func TestURLMissingParamPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic for missing param")
		}
	}()

	r := New()
	r.Get("/users/{id}", noopHandler).Name("user")
	r.URL("user", "wrong", "42")
}

func TestURLExtraParamsPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic for extra params")
		}
	}()

	r := New()
	r.Get("/users/{id}", noopHandler).Name("user")
	r.URL("user", "id", "42", "extra", "value")
}

// ─── Parameter Constraints ──────────────────────────────────────────────────

func TestConstraintInt(t *testing.T) {
	c := Int("id")
	if !c.Validate("42") {
		t.Error("expected 42 to be valid int")
	}
	if !c.Validate("-1") {
		t.Error("expected -1 to be valid int")
	}
	if c.Validate("abc") {
		t.Error("expected 'abc' to be invalid int")
	}
	if c.Validate("") {
		t.Error("expected empty string to be invalid int")
	}
}

func TestConstraintUUID(t *testing.T) {
	c := UUID("id")
	if !c.Validate("550e8400-e29b-41d4-a716-446655440000") {
		t.Error("expected valid UUID to pass")
	}
	if c.Validate("not-a-uuid") {
		t.Error("expected invalid UUID to fail")
	}
	if c.Validate("") {
		t.Error("expected empty string to fail")
	}
}

func TestConstraintRegex(t *testing.T) {
	c := Regex("slug", `^[a-z0-9-]+$`)
	if !c.Validate("hello-world") {
		t.Error("expected 'hello-world' to match")
	}
	if c.Validate("Hello World!") {
		t.Error("expected 'Hello World!' not to match")
	}
}

func TestConstraintRegexInvalidPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic on invalid regex")
		}
	}()
	Regex("id", "[invalid")
}

func TestConstraintOneOf(t *testing.T) {
	c := OneOf("status", "active", "inactive", "pending")
	if !c.Validate("active") {
		t.Error("expected 'active' to be valid")
	}
	if c.Validate("deleted") {
		t.Error("expected 'deleted' to be invalid")
	}
}

func TestConstraintMultiple(t *testing.T) {
	r := New()
	r.Get("/users/{id}/status/{status}", ValidateParams(
		func(w http.ResponseWriter, req *http.Request) error {
			fmt.Fprint(w, "ok")
			return nil
		},
		Int("id"),
		OneOf("status", "active", "inactive"),
	))

	// Valid request.
	rec := doRequest(r, "GET", "/users/42/status/active")
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	// Invalid id.
	rec = doRequest(r, "GET", "/users/abc/status/active")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid id, got %d", rec.Code)
	}

	// Invalid status.
	rec = doRequest(r, "GET", "/users/42/status/deleted")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid status, got %d", rec.Code)
	}
}

func TestConstraintErrorFormat(t *testing.T) {
	r := New()
	r.Get("/items/{id}", ValidateParams(noopHandler, Int("id")))

	rec := doRequest(r, "GET", "/items/abc")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}

	var env errorEnvelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if env.Success {
		t.Error("expected success=false")
	}
	if env.Error.Code != "BAD_REQUEST" {
		t.Errorf("expected code BAD_REQUEST, got %s", env.Error.Code)
	}
	if !strings.Contains(env.Error.Message, "integer") {
		t.Errorf("expected error message to mention 'integer', got %q", env.Error.Message)
	}
}

func TestConstraintPassThrough(t *testing.T) {
	called := false
	r := New()
	r.Get("/items/{id}", ValidateParams(
		func(w http.ResponseWriter, req *http.Request) error {
			called = true
			fmt.Fprint(w, req.PathValue("id"))
			return nil
		},
		Int("id"),
	))

	rec := doRequest(r, "GET", "/items/99")
	if !called {
		t.Error("expected handler to be called")
	}
	if rec.Body.String() != "99" {
		t.Errorf("expected body '99', got %q", rec.Body.String())
	}
}

func TestConstraintIntegration(t *testing.T) {
	r := New()
	api := r.Group("/api")
	api.Get("/users/{id}", ValidateParams(
		func(w http.ResponseWriter, req *http.Request) error {
			w.WriteHeader(http.StatusOK)
			return nil
		},
		Int("id"),
	))

	// Valid.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/users/123", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	// Invalid.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/users/abc", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// ─── Test Helpers ───────────────────────────────────────────────────────────

func noopHandler(_ http.ResponseWriter, _ *http.Request) error { return nil }

func namedTestHandler(_ http.ResponseWriter, _ *http.Request) error { return nil }

func namedStdlibHandler(_ http.ResponseWriter, _ *http.Request) {}
