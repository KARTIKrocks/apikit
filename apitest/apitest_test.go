package apitest_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/KARTIKrocks/apikit/apitest"
	"github.com/KARTIKrocks/apikit/errors"
	"github.com/KARTIKrocks/apikit/response"
)

// --- RequestBuilder tests ---

func TestNewRequest_GET(t *testing.T) {
	req := apitest.NewRequest("GET", "/api/v1/users").Build()

	if req.Method != "GET" {
		t.Errorf("expected method GET, got %s", req.Method)
	}
	if req.URL.Path != "/api/v1/users" {
		t.Errorf("expected path /api/v1/users, got %s", req.URL.Path)
	}
}

func TestNewRequest_WithBody(t *testing.T) {
	type body struct {
		Name string `json:"name"`
	}

	req := apitest.NewRequest("POST", "/users").
		WithBody(body{Name: "Alice"}).
		Build()

	if req.Header.Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", req.Header.Get("Content-Type"))
	}
	if req.Body == nil {
		t.Fatal("expected non-nil body")
	}
}

func TestNewRequest_WithQuery(t *testing.T) {
	req := apitest.NewRequest("GET", "/users").
		WithQuery("page", "2").
		WithQuery("limit", "10").
		Build()

	if req.URL.Query().Get("page") != "2" {
		t.Errorf("expected page=2, got %s", req.URL.Query().Get("page"))
	}
	if req.URL.Query().Get("limit") != "10" {
		t.Errorf("expected limit=10, got %s", req.URL.Query().Get("limit"))
	}
}

func TestNewRequest_WithHeader(t *testing.T) {
	req := apitest.NewRequest("GET", "/users").
		WithHeader("X-Custom", "value").
		Build()

	if req.Header.Get("X-Custom") != "value" {
		t.Errorf("expected X-Custom=value, got %s", req.Header.Get("X-Custom"))
	}
}

func TestNewRequest_WithBearerToken(t *testing.T) {
	req := apitest.NewRequest("GET", "/users").
		WithBearerToken("test-token").
		Build()

	expected := "Bearer test-token"
	if req.Header.Get("Authorization") != expected {
		t.Errorf("expected Authorization=%q, got %q", expected, req.Header.Get("Authorization"))
	}
}

func TestNewRequest_WithPathValue(t *testing.T) {
	req := apitest.NewRequest("GET", "/users/{id}").
		WithPathValue("id", "123").
		Build()

	if req.PathValue("id") != "123" {
		t.Errorf("expected path value id=123, got %s", req.PathValue("id"))
	}
}

func TestNewRequest_WithContext(t *testing.T) {
	type ctxKey string
	ctx := context.WithValue(context.Background(), ctxKey("key"), "val")

	req := apitest.NewRequest("GET", "/users").
		WithContext(ctx).
		Build()

	if req.Context().Value(ctxKey("key")) != "val" {
		t.Error("expected context value to be preserved")
	}
}

// --- Record tests ---

func TestRecord(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response.OK(w, "hello", map[string]string{"key": "value"})
	})

	req := apitest.NewRequest("GET", "/test").Build()
	resp := apitest.Record(handler, req)

	resp.AssertStatus(t, 200)
	resp.AssertSuccess(t)
	resp.AssertBodyContains(t, "hello")
	resp.AssertBodyContains(t, "value")
}

func TestRecordHandler(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) error {
		response.Created(w, "created", map[string]string{"id": "1"})
		return nil
	}

	req := apitest.NewRequest("POST", "/test").Build()
	resp := apitest.RecordHandler(handler, req)

	resp.AssertStatus(t, 201)
	resp.AssertSuccess(t)
}

func TestRecordHandler_Error(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) error {
		return errors.NotFound("User")
	}

	req := apitest.NewRequest("GET", "/users/999").Build()
	resp := apitest.RecordHandler(handler, req)

	resp.AssertStatus(t, 404)
	resp.AssertError(t, "NOT_FOUND")
}

// --- Response method tests ---

func TestResponse_Decode(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response.OK(w, "ok", map[string]string{"name": "Alice"})
	})

	req := apitest.NewRequest("GET", "/test").Build()
	resp := apitest.Record(handler, req)

	var env response.Envelope
	if err := resp.Decode(&env); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if !env.Success {
		t.Error("expected success=true")
	}
}

func TestResponse_Envelope(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response.OK(w, "retrieved", nil)
	})

	req := apitest.NewRequest("GET", "/test").Build()
	resp := apitest.Record(handler, req)

	env, err := resp.Envelope()
	if err != nil {
		t.Fatalf("failed to decode envelope: %v", err)
	}
	if env.Message != "retrieved" {
		t.Errorf("expected message 'retrieved', got %q", env.Message)
	}
}

func TestResponse_Header(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-ID", "abc")
		response.OK(w, "ok", nil)
	})

	req := apitest.NewRequest("GET", "/test").Build()
	resp := apitest.Record(handler, req)

	resp.AssertHeader(t, "X-Request-ID", "abc")
}

// --- Assertion tests ---

func TestAssertError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) error {
		return errors.BadRequest("invalid input")
	}

	req := apitest.NewRequest("POST", "/test").Build()
	resp := apitest.RecordHandler(handler, req)

	resp.AssertStatus(t, 400)
	resp.AssertError(t, "BAD_REQUEST")
}

func TestAssertValidationError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) error {
		return errors.Validation("Validation failed", map[string]string{
			"email": "is required",
			"name":  "too short",
		})
	}

	req := apitest.NewRequest("POST", "/test").Build()
	resp := apitest.RecordHandler(handler, req)

	resp.AssertStatus(t, 422)
	resp.AssertValidationError(t, "email")
	resp.AssertValidationError(t, "name")
}

func TestAssertBodyContains(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response.OK(w, "hello alice", nil)
	})

	req := apitest.NewRequest("GET", "/test").Build()
	resp := apitest.Record(handler, req)

	resp.AssertBodyContains(t, "alice")
}

// --- Integration: full request-response cycle ---

func TestFullCycle(t *testing.T) {
	type CreateReq struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	type User struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	handler := func(w http.ResponseWriter, r *http.Request) error {
		response.Created(w, "User created", User{
			ID:    "new-1",
			Name:  "Alice",
			Email: "alice@example.com",
		})
		return nil
	}

	req := apitest.NewRequest("POST", "/api/v1/users").
		WithBody(CreateReq{Name: "Alice", Email: "alice@example.com"}).
		WithBearerToken("test-token").
		WithHeader("X-Request-ID", "req-1").
		Build()

	resp := apitest.RecordHandler(handler, req)

	resp.AssertStatus(t, 201)
	resp.AssertSuccess(t)
	resp.AssertBodyContains(t, "Alice")
	resp.AssertBodyContains(t, "alice@example.com")
}
