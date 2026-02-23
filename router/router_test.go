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
