package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KARTIKrocks/apikit/errors"
	"github.com/KARTIKrocks/apikit/middleware"
)

// --- helpers ---

// headerMiddleware returns middleware that sets a response header.
func headerMiddleware(key, value string) middleware.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add(key, value)
			next.ServeHTTP(w, r)
		})
	}
}

func doRequest(handler http.Handler, method, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	handler.ServeHTTP(rec, req)
	return rec
}

// --- tests ---

func TestMethodHelpers(t *testing.T) {
	methods := []struct {
		register func(r *Router, pattern string, fn HandlerFunc)
		method   string
	}{
		{(*Router).Get, "GET"},
		{(*Router).Post, "POST"},
		{(*Router).Put, "PUT"},
		{(*Router).Patch, "PATCH"},
		{(*Router).Delete, "DELETE"},
	}

	for _, tt := range methods {
		t.Run(tt.method, func(t *testing.T) {
			r := New()
			tt.register(r, "/test", func(w http.ResponseWriter, req *http.Request) error {
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, tt.method)
				return nil
			})

			rec := doRequest(r, tt.method, "/test")
			if rec.Code != http.StatusOK {
				t.Errorf("expected 200, got %d", rec.Code)
			}
			if rec.Body.String() != tt.method {
				t.Errorf("expected body %q, got %q", tt.method, rec.Body.String())
			}
		})
	}
}

func TestHandle(t *testing.T) {
	r := New()
	r.Handle("GET /stdlib", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, "stdlib-handler")
	}))

	rec := doRequest(r, "GET", "/stdlib")
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "stdlib-handler" {
		t.Errorf("expected body %q, got %q", "stdlib-handler", rec.Body.String())
	}
}

func TestHandleFunc(t *testing.T) {
	r := New()
	r.HandleFunc("POST /func", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, "handlefunc")
	})

	rec := doRequest(r, "POST", "/func")
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "handlefunc" {
		t.Errorf("expected body %q, got %q", "handlefunc", rec.Body.String())
	}
}

func TestErrorHandlerDefault(t *testing.T) {
	r := New()
	r.Get("/fail", func(w http.ResponseWriter, req *http.Request) error {
		return errors.NotFound("User")
	})

	rec := doRequest(r, "GET", "/fail")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}

	var env errorEnvelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if env.Success {
		t.Error("expected success=false")
	}
	if env.Error.Code != "NOT_FOUND" {
		t.Errorf("expected code NOT_FOUND, got %s", env.Error.Code)
	}
}

func TestErrorHandlerDefaultInternalError(t *testing.T) {
	r := New()
	r.Get("/err", func(w http.ResponseWriter, req *http.Request) error {
		return fmt.Errorf("something broke")
	})

	rec := doRequest(r, "GET", "/err")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}

	var env errorEnvelope
	json.NewDecoder(rec.Body).Decode(&env)
	if env.Error.Code != "INTERNAL_ERROR" {
		t.Errorf("expected INTERNAL_ERROR, got %s", env.Error.Code)
	}
	if env.Error.Message != "An internal error occurred" {
		t.Errorf("expected generic message, got %s", env.Error.Message)
	}
}

func TestCustomErrorHandler(t *testing.T) {
	var captured error
	r := New(WithErrorHandler(func(w http.ResponseWriter, req *http.Request, err error) {
		captured = err
		w.WriteHeader(http.StatusTeapot)
	}))
	r.Get("/custom", func(w http.ResponseWriter, req *http.Request) error {
		return errors.BadRequest("oops")
	})

	rec := doRequest(r, "GET", "/custom")
	if rec.Code != http.StatusTeapot {
		t.Errorf("expected 418, got %d", rec.Code)
	}
	if captured == nil {
		t.Error("expected error to be captured")
	}
}

func TestGroupPrefix(t *testing.T) {
	r := New()
	api := r.Group("/api")
	api.Get("/users", func(w http.ResponseWriter, req *http.Request) error {
		fmt.Fprint(w, "users")
		return nil
	})

	rec := doRequest(r, "GET", "/api/users")
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "users" {
		t.Errorf("expected body %q, got %q", "users", rec.Body.String())
	}
}

func TestNestedGroups(t *testing.T) {
	r := New()
	api := r.Group("/api")
	v1 := api.Group("/v1")
	v1.Get("/items", func(w http.ResponseWriter, req *http.Request) error {
		fmt.Fprint(w, "items-v1")
		return nil
	})

	rec := doRequest(r, "GET", "/api/v1/items")
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "items-v1" {
		t.Errorf("expected body %q, got %q", "items-v1", rec.Body.String())
	}
}

func TestGroupMiddleware(t *testing.T) {
	r := New()

	api := r.Group("/api", headerMiddleware("X-Group", "api"))
	api.Get("/data", func(w http.ResponseWriter, req *http.Request) error {
		fmt.Fprint(w, "data")
		return nil
	})

	// Route outside the group should NOT have the middleware header.
	r.Get("/health", func(w http.ResponseWriter, req *http.Request) error {
		fmt.Fprint(w, "ok")
		return nil
	})

	rec := doRequest(r, "GET", "/api/data")
	if rec.Header().Get("X-Group") != "api" {
		t.Errorf("expected X-Group=api, got %q", rec.Header().Get("X-Group"))
	}

	rec2 := doRequest(r, "GET", "/health")
	if rec2.Header().Get("X-Group") != "" {
		t.Error("expected no X-Group header on /health")
	}
}

func TestMiddlewareOrdering(t *testing.T) {
	r := New()
	r.Use(headerMiddleware("X-Order", "root"))

	api := r.Group("/api", headerMiddleware("X-Order", "api"))
	v1 := api.Group("/v1", headerMiddleware("X-Order", "v1"))
	v1.Get("/test", func(w http.ResponseWriter, req *http.Request) error {
		return nil
	})

	rec := doRequest(r, "GET", "/api/v1/test")
	// Middleware runs root → api → v1, each adds to X-Order.
	values := rec.Header().Values("X-Order")
	if len(values) != 3 {
		t.Fatalf("expected 3 X-Order values, got %d: %v", len(values), values)
	}
	expected := []string{"root", "api", "v1"}
	for i, v := range values {
		if v != expected[i] {
			t.Errorf("X-Order[%d]: expected %q, got %q", i, expected[i], v)
		}
	}
}

func TestUseAppliesToSubsequentRoutes(t *testing.T) {
	r := New()

	// Register a route BEFORE Use.
	r.Get("/before", func(w http.ResponseWriter, req *http.Request) error {
		return nil
	})

	r.Use(headerMiddleware("X-After", "yes"))

	// Register a route AFTER Use.
	r.Get("/after", func(w http.ResponseWriter, req *http.Request) error {
		return nil
	})

	rec1 := doRequest(r, "GET", "/before")
	if rec1.Header().Get("X-After") != "" {
		t.Error("expected no X-After on /before")
	}

	rec2 := doRequest(r, "GET", "/after")
	if rec2.Header().Get("X-After") != "yes" {
		t.Error("expected X-After=yes on /after")
	}
}

func TestRouterImplementsHTTPHandler(t *testing.T) {
	var _ http.Handler = New()
}

func TestGroupHandleAndHandleFunc(t *testing.T) {
	r := New()
	api := r.Group("/api", headerMiddleware("X-MW", "applied"))

	api.Handle("GET /raw", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, "raw")
	}))

	api.HandleFunc("POST /func", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, "func")
	})

	rec1 := doRequest(r, "GET", "/api/raw")
	if rec1.Body.String() != "raw" {
		t.Errorf("expected body %q, got %q", "raw", rec1.Body.String())
	}
	if rec1.Header().Get("X-MW") != "applied" {
		t.Error("expected middleware to be applied to Handle")
	}

	rec2 := doRequest(r, "POST", "/api/func")
	if rec2.Body.String() != "func" {
		t.Errorf("expected body %q, got %q", "func", rec2.Body.String())
	}
	if rec2.Header().Get("X-MW") != "applied" {
		t.Error("expected middleware to be applied to HandleFunc")
	}
}

func TestErrorHandlerWithFields(t *testing.T) {
	r := New()
	r.Get("/validate", func(w http.ResponseWriter, req *http.Request) error {
		return errors.Validation("Validation failed", map[string]string{
			"email": "is required",
			"name":  "too short",
		})
	})

	rec := doRequest(r, "GET", "/validate")
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rec.Code)
	}

	var env errorEnvelope
	json.NewDecoder(rec.Body).Decode(&env)
	if len(env.Error.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(env.Error.Fields))
	}
	if env.Error.Fields["email"] != "is required" {
		t.Errorf("expected email field error, got %q", env.Error.Fields["email"])
	}
}

func TestNoErrorReturned(t *testing.T) {
	r := New()
	r.Get("/ok", func(w http.ResponseWriter, req *http.Request) error {
		fmt.Fprint(w, "success")
		return nil
	})

	rec := doRequest(r, "GET", "/ok")
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "success" {
		t.Errorf("expected body %q, got %q", "success", rec.Body.String())
	}
}

func TestGroupHandleWithoutMethod(t *testing.T) {
	r := New()
	api := r.Group("/api")
	api.Handle("/catch-all", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, "caught")
	}))

	rec := doRequest(r, "GET", "/api/catch-all")
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "caught" {
		t.Errorf("expected body %q, got %q", "caught", rec.Body.String())
	}
}

func TestMethodFuncHelpers(t *testing.T) {
	r := New()

	// GetFunc — stdlib handler (no error return)
	r.GetFunc("/std", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, "get-func")
	})

	// PostFunc — stdlib handler
	r.PostFunc("/std", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, "post-func")
	})

	// PutFunc
	r.PutFunc("/std", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, "put-func")
	})

	// PatchFunc
	r.PatchFunc("/std", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, "patch-func")
	})

	// DeleteFunc
	r.DeleteFunc("/std", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, "delete-func")
	})

	for _, tt := range []struct {
		method, body string
	}{
		{"GET", "get-func"},
		{"POST", "post-func"},
		{"PUT", "put-func"},
		{"PATCH", "patch-func"},
		{"DELETE", "delete-func"},
	} {
		rec := doRequest(r, tt.method, "/std")
		if rec.Body.String() != tt.body {
			t.Errorf("%s: expected body %q, got %q", tt.method, tt.body, rec.Body.String())
		}
	}
}

func TestGroupFuncHelpers(t *testing.T) {
	r := New()
	api := r.Group("/api", headerMiddleware("X-MW", "yes"))

	api.GetFunc("/data", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, "data")
	})

	rec := doRequest(r, "GET", "/api/data")
	if rec.Body.String() != "data" {
		t.Errorf("expected body %q, got %q", "data", rec.Body.String())
	}
	if rec.Header().Get("X-MW") != "yes" {
		t.Error("expected middleware to be applied to GetFunc")
	}
}

func TestNotFoundReturnsJSON(t *testing.T) {
	r := New()
	r.Get("/exists", func(w http.ResponseWriter, req *http.Request) error {
		fmt.Fprint(w, "ok")
		return nil
	})

	rec := doRequest(r, "GET", "/does-not-exist")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}

	var env errorEnvelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("failed to decode JSON 404 response: %v", err)
	}
	if env.Success {
		t.Error("expected success=false")
	}
	if env.Error.Code != "NOT_FOUND" {
		t.Errorf("expected code NOT_FOUND, got %s", env.Error.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("expected JSON content type, got %q", ct)
	}
}

func TestMethodNotAllowedReturnsJSON(t *testing.T) {
	r := New()
	r.Get("/users", func(w http.ResponseWriter, req *http.Request) error {
		fmt.Fprint(w, "users")
		return nil
	})

	// POST /users is not registered, only GET is — ServeMux returns 405.
	rec := doRequest(r, "POST", "/users")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}

	var env errorEnvelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("failed to decode JSON 405 response: %v", err)
	}
	if env.Success {
		t.Error("expected success=false")
	}
	if env.Error.Code != "METHOD_NOT_ALLOWED" {
		t.Errorf("expected code METHOD_NOT_ALLOWED, got %s", env.Error.Code)
	}
}

func TestHandlerWriting404IsNotIntercepted(t *testing.T) {
	r := New()
	r.Get("/custom-404", func(w http.ResponseWriter, req *http.Request) error {
		// Handler explicitly writes a 404 with its own body — should NOT be intercepted.
		w.Write([]byte("custom body"))
		w.WriteHeader(http.StatusNotFound) // after Write, this is a no-op per http spec, but tests the probe path
		return nil
	})

	rec := doRequest(r, "GET", "/custom-404")
	if rec.Body.String() != "custom body" {
		t.Errorf("expected custom body, got %q", rec.Body.String())
	}
}

func TestDoubleSlashNormalization(t *testing.T) {
	r := New()
	// Trailing slash on prefix + leading slash on pattern.
	api := r.Group("/api/")
	api.Get("/users", func(w http.ResponseWriter, req *http.Request) error {
		fmt.Fprint(w, "users")
		return nil
	})

	rec := doRequest(r, "GET", "/api/users")
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "users" {
		t.Errorf("expected body %q, got %q", "users", rec.Body.String())
	}
}

func TestDoubleSlashNestedGroups(t *testing.T) {
	r := New()
	api := r.Group("/api/")
	v1 := api.Group("/v1/")
	v1.Get("/items", func(w http.ResponseWriter, req *http.Request) error {
		fmt.Fprint(w, "items")
		return nil
	})

	rec := doRequest(r, "GET", "/api/v1/items")
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "items" {
		t.Errorf("expected body %q, got %q", "items", rec.Body.String())
	}
}

func TestDoubleSlashGroupHandle(t *testing.T) {
	r := New()
	api := r.Group("/api/")
	api.Handle("GET /raw", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, "raw")
	}))

	rec := doRequest(r, "GET", "/api/raw")
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "raw" {
		t.Errorf("expected body %q, got %q", "raw", rec.Body.String())
	}
}

// --- Pattern precedence with similar routes ---

func TestOverlappingPatterns(t *testing.T) {
	r := New()

	// Wildcard route — catches /users/<anything>
	r.Get("/users/{id}", func(w http.ResponseWriter, req *http.Request) error {
		fmt.Fprintf(w, "wildcard:%s", req.PathValue("id"))
		return nil
	})

	// Exact literal routes — should take priority over wildcard
	r.Get("/users/me", func(w http.ResponseWriter, req *http.Request) error {
		fmt.Fprint(w, "exact:me")
		return nil
	})
	r.Get("/users/21", func(w http.ResponseWriter, req *http.Request) error {
		fmt.Fprint(w, "exact:21")
		return nil
	})

	tests := []struct {
		path     string
		wantBody string
	}{
		{"/users/me", "exact:me"},    // exact match wins over {id}
		{"/users/21", "exact:21"},    // exact match wins over {id}
		{"/users/42", "wildcard:42"}, // no exact match — falls through to wildcard
		{"/users/hello", "wildcard:hello"},
		{"/users/me-too", "wildcard:me-too"}, // not "me" — wildcard
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			rec := doRequest(r, "GET", tt.path)
			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", rec.Code)
			}
			if rec.Body.String() != tt.wantBody {
				t.Errorf("expected body %q, got %q", tt.wantBody, rec.Body.String())
			}
		})
	}
}

func TestOverlappingPatternsInGroup(t *testing.T) {
	r := New()
	api := r.Group("/api/v1")

	api.Get("/users/{id}", func(w http.ResponseWriter, req *http.Request) error {
		fmt.Fprintf(w, "wildcard:%s", req.PathValue("id"))
		return nil
	})
	api.Get("/users/me", func(w http.ResponseWriter, req *http.Request) error {
		fmt.Fprint(w, "exact:me")
		return nil
	})
	api.Get("/users/search", func(w http.ResponseWriter, req *http.Request) error {
		fmt.Fprint(w, "exact:search")
		return nil
	})
	api.Post("/users/{id}", func(w http.ResponseWriter, req *http.Request) error {
		fmt.Fprintf(w, "post:%s", req.PathValue("id"))
		return nil
	})

	tests := []struct {
		method   string
		path     string
		wantBody string
	}{
		{"GET", "/api/v1/users/me", "exact:me"},
		{"GET", "/api/v1/users/search", "exact:search"},
		{"GET", "/api/v1/users/99", "wildcard:99"},
		{"GET", "/api/v1/users/abc", "wildcard:abc"},
		{"POST", "/api/v1/users/99", "post:99"},
		{"POST", "/api/v1/users/me", "post:me"}, // POST has no /me exact — wildcard handles it
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			rec := doRequest(r, tt.method, tt.path)
			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", rec.Code)
			}
			if rec.Body.String() != tt.wantBody {
				t.Errorf("expected body %q, got %q", tt.wantBody, rec.Body.String())
			}
		})
	}
}

func TestSimilarPrefixPatterns(t *testing.T) {
	r := New()

	// Routes with very similar prefixes
	r.Get("/users", func(w http.ResponseWriter, req *http.Request) error {
		fmt.Fprint(w, "list-users")
		return nil
	})
	r.Get("/users/{id}", func(w http.ResponseWriter, req *http.Request) error {
		fmt.Fprintf(w, "get-user:%s", req.PathValue("id"))
		return nil
	})
	r.Get("/users/{id}/posts", func(w http.ResponseWriter, req *http.Request) error {
		fmt.Fprintf(w, "user-posts:%s", req.PathValue("id"))
		return nil
	})
	r.Get("/users/{id}/posts/{postID}", func(w http.ResponseWriter, req *http.Request) error {
		fmt.Fprintf(w, "user-post:%s:%s", req.PathValue("id"), req.PathValue("postID"))
		return nil
	})
	r.Get("/users/{id}/comments", func(w http.ResponseWriter, req *http.Request) error {
		fmt.Fprintf(w, "user-comments:%s", req.PathValue("id"))
		return nil
	})

	tests := []struct {
		path     string
		wantBody string
	}{
		{"/users", "list-users"},
		{"/users/5", "get-user:5"},
		{"/users/5/posts", "user-posts:5"},
		{"/users/5/posts/10", "user-post:5:10"},
		{"/users/5/comments", "user-comments:5"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			rec := doRequest(r, "GET", tt.path)
			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", rec.Code)
			}
			if rec.Body.String() != tt.wantBody {
				t.Errorf("expected body %q, got %q", tt.wantBody, rec.Body.String())
			}
		})
	}
}

func TestSplitPatternExtraSpaces(t *testing.T) {
	method, path := splitPattern("GET  /users")
	if method != "GET" {
		t.Errorf("expected method GET, got %q", method)
	}
	if path != "/users" {
		t.Errorf("expected path /users, got %q", path)
	}
}

func TestSplitPatternNoMethod(t *testing.T) {
	method, path := splitPattern("/users")
	if method != "" {
		t.Errorf("expected empty method, got %q", method)
	}
	if path != "/users" {
		t.Errorf("expected path /users, got %q", path)
	}
}

func TestSplitPatternNormalCase(t *testing.T) {
	method, path := splitPattern("POST /users")
	if method != "POST" {
		t.Errorf("expected method POST, got %q", method)
	}
	if path != "/users" {
		t.Errorf("expected path /users, got %q", path)
	}
}

func TestHeadAndOptionsHelpers(t *testing.T) {
	r := New()
	r.Head("/ping", func(w http.ResponseWriter, req *http.Request) error {
		w.Header().Set("X-Ping", "pong")
		return nil
	})
	r.Options("/cors", func(w http.ResponseWriter, req *http.Request) error {
		w.Header().Set("Allow", "GET, POST, OPTIONS")
		return nil
	})

	rec := doRequest(r, "HEAD", "/ping")
	if rec.Code != http.StatusOK {
		t.Errorf("HEAD: expected 200, got %d", rec.Code)
	}
	if rec.Header().Get("X-Ping") != "pong" {
		t.Errorf("HEAD: expected X-Ping=pong, got %q", rec.Header().Get("X-Ping"))
	}

	rec = doRequest(r, "OPTIONS", "/cors")
	if rec.Code != http.StatusOK {
		t.Errorf("OPTIONS: expected 200, got %d", rec.Code)
	}
	if rec.Header().Get("Allow") != "GET, POST, OPTIONS" {
		t.Errorf("OPTIONS: expected Allow header, got %q", rec.Header().Get("Allow"))
	}
}

func TestHeadFuncAndOptionsFuncHelpers(t *testing.T) {
	r := New()
	r.HeadFunc("/ping", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("X-Ping", "pong")
	})
	r.OptionsFunc("/cors", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Allow", "GET, OPTIONS")
	})

	rec := doRequest(r, "HEAD", "/ping")
	if rec.Code != http.StatusOK {
		t.Errorf("HeadFunc: expected 200, got %d", rec.Code)
	}
	if rec.Header().Get("X-Ping") != "pong" {
		t.Errorf("HeadFunc: expected X-Ping=pong")
	}

	rec = doRequest(r, "OPTIONS", "/cors")
	if rec.Code != http.StatusOK {
		t.Errorf("OptionsFunc: expected 200, got %d", rec.Code)
	}
	if rec.Header().Get("Allow") != "GET, OPTIONS" {
		t.Errorf("OptionsFunc: expected Allow header")
	}
}

func TestGroupHeadAndOptions(t *testing.T) {
	r := New()
	api := r.Group("/api", headerMiddleware("X-MW", "yes"))

	api.Head("/ping", func(w http.ResponseWriter, req *http.Request) error {
		w.Header().Set("X-Ping", "pong")
		return nil
	})
	api.Options("/cors", func(w http.ResponseWriter, req *http.Request) error {
		w.Header().Set("Allow", "GET")
		return nil
	})

	rec := doRequest(r, "HEAD", "/api/ping")
	if rec.Code != http.StatusOK {
		t.Errorf("Group HEAD: expected 200, got %d", rec.Code)
	}
	if rec.Header().Get("X-MW") != "yes" {
		t.Error("Group HEAD: expected middleware applied")
	}

	rec = doRequest(r, "OPTIONS", "/api/cors")
	if rec.Code != http.StatusOK {
		t.Errorf("Group OPTIONS: expected 200, got %d", rec.Code)
	}
	if rec.Header().Get("X-MW") != "yes" {
		t.Error("Group OPTIONS: expected middleware applied")
	}
}
