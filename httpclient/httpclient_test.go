package httpclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// newTestServer returns an httptest.Server and a Client pointed at it.
func newTestServer(handler http.HandlerFunc, opts ...Option) (*httptest.Server, *Client) {
	ts := httptest.NewServer(handler)
	// Suppress log noise in tests.
	opts = append([]Option{WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil)))}, opts...)
	c := New(ts.URL, opts...)
	return ts, c
}

// --- Basic CRUD ---

func TestGet(t *testing.T) {
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		w.WriteHeader(200)
		fmt.Fprint(w, `{"ok":true}`)
	})
	defer ts.Close()

	resp, err := c.Get(context.Background(), "/ping")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestPost(t *testing.T) {
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		var m map[string]string
		json.Unmarshal(body, &m)
		if m["name"] != "test" {
			t.Fatalf("unexpected body: %s", body)
		}
		w.WriteHeader(201)
		fmt.Fprint(w, `{"id":1}`)
	})
	defer ts.Close()

	resp, err := c.Post(context.Background(), "/items", map[string]string{"name": "test"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 201 {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
}

func TestPut(t *testing.T) {
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}
		w.WriteHeader(200)
	})
	defer ts.Close()

	resp, err := c.Put(context.Background(), "/items/1", map[string]string{"name": "updated"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestPatch(t *testing.T) {
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("expected PATCH, got %s", r.Method)
		}
		w.WriteHeader(200)
	})
	defer ts.Close()

	resp, err := c.Patch(context.Background(), "/items/1", map[string]string{"name": "patched"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestDelete(t *testing.T) {
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(204)
	})
	defer ts.Close()

	resp, err := c.Delete(context.Background(), "/items/1")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 204 {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
}

// --- Retries ---

func TestRetryOn500(t *testing.T) {
	var calls atomic.Int32
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n < 3 {
			w.WriteHeader(500)
			fmt.Fprint(w, `{"error":"fail"}`)
			return
		}
		w.WriteHeader(200)
		fmt.Fprint(w, `{"ok":true}`)
	}, WithMaxRetries(3), WithRetryDelay(time.Millisecond), WithMaxRetryDelay(5*time.Millisecond))
	defer ts.Close()

	resp, err := c.Get(context.Background(), "/flaky")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if calls.Load() != 3 {
		t.Fatalf("expected 3 calls, got %d", calls.Load())
	}
}

func TestNoRetryOn4xx(t *testing.T) {
	var calls atomic.Int32
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(404)
		fmt.Fprint(w, `{"error":"not found"}`)
	}, WithMaxRetries(3), WithRetryDelay(time.Millisecond))
	defer ts.Close()

	resp, err := c.Get(context.Background(), "/missing")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if resp.StatusCode != 404 {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
	if calls.Load() != 1 {
		t.Fatalf("expected 1 call (no retry on 4xx), got %d", calls.Load())
	}
}

func TestContextCancellationStopsRetries(t *testing.T) {
	var calls atomic.Int32
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(500)
		fmt.Fprint(w, `{"error":"fail"}`)
	}, WithMaxRetries(10), WithRetryDelay(50*time.Millisecond))
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := c.Get(ctx, "/slow")
	if err == nil {
		t.Fatal("expected error")
	}
	if calls.Load() > 5 {
		t.Fatalf("expected retries to stop early, got %d calls", calls.Load())
	}
}

func TestExhaustedRetriesReturnsResponse(t *testing.T) {
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		fmt.Fprint(w, `{"error":"boom"}`)
	}, WithMaxRetries(2), WithRetryDelay(time.Millisecond), WithMaxRetryDelay(time.Millisecond))
	defer ts.Close()

	resp, err := c.Get(context.Background(), "/down")
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}

	// The response must survive the retry loop so callers can inspect status
	// and read the error body without unwrapping the error.
	if resp == nil {
		t.Fatal("expected non-nil response on retry exhaustion")
	}
	if resp.StatusCode != 500 {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
	if !resp.IsServerError() {
		t.Fatal("expected IsServerError to be true")
	}

	var body struct {
		Error string `json:"error"`
	}
	if err := resp.JSON(&body); err != nil {
		t.Fatalf("failed to decode error body from response: %v", err)
	}
	if body.Error != "boom" {
		t.Fatalf("expected error body %q, got %q", "boom", body.Error)
	}
}

func TestErrorOnStatusDisabled_ClientError(t *testing.T) {
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		fmt.Fprint(w, `{"error":"not found"}`)
	}, WithErrorOnStatus(false))
	defer ts.Close()

	resp, err := c.Get(context.Background(), "/missing")
	if err != nil {
		t.Fatalf("expected nil error with error-on-status disabled, got %v", err)
	}
	if resp == nil || !resp.IsClientError() {
		t.Fatalf("expected a 4xx response, got %+v", resp)
	}

	var body struct {
		Error string `json:"error"`
	}
	if err := resp.JSON(&body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if body.Error != "not found" {
		t.Fatalf("expected %q, got %q", "not found", body.Error)
	}
}

func TestErrorOnStatusDisabled_ServerErrorAfterRetries(t *testing.T) {
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(503)
		fmt.Fprint(w, `{"error":"unavailable"}`)
	}, WithErrorOnStatus(false), WithMaxRetries(2),
		WithRetryDelay(time.Millisecond), WithMaxRetryDelay(time.Millisecond))
	defer ts.Close()

	resp, err := c.Get(context.Background(), "/down")
	if err != nil {
		t.Fatalf("expected nil error after retries with error-on-status disabled, got %v", err)
	}
	if resp == nil || !resp.IsServerError() || resp.StatusCode != 503 {
		t.Fatalf("expected a 503 response, got %+v", resp)
	}
}

func TestErrorOnStatusDisabled_TransportErrorStillErrors(t *testing.T) {
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {},
		WithErrorOnStatus(false), WithMaxRetries(0))
	// Close the server so the connection is refused: a transport failure, not
	// a status error, must still surface even with error-on-status disabled.
	ts.Close()

	resp, err := c.Get(context.Background(), "/anything")
	if err == nil {
		t.Fatalf("expected a transport error, got resp=%+v", resp)
	}
}

func TestRequestBuilder_ErrorOnStatusOverrideDisables(t *testing.T) {
	// Client defaults to error-on-status enabled; the request opts out.
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		fmt.Fprint(w, `{"error":"nope"}`)
	})
	defer ts.Close()

	resp, err := c.Request().Path("/missing").ErrorOnStatus(false).Get(context.Background())
	if err != nil {
		t.Fatalf("expected nil error from per-request override, got %v", err)
	}
	if resp == nil || !resp.IsClientError() {
		t.Fatalf("expected a 4xx response, got %+v", resp)
	}
}

func TestRequestBuilder_ErrorOnStatusOverrideEnables(t *testing.T) {
	// Client defaults to error-on-status disabled; the request opts back in.
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		fmt.Fprint(w, `{"error":"nope"}`)
	}, WithErrorOnStatus(false))
	defer ts.Close()

	resp, err := c.Request().Path("/missing").ErrorOnStatus(true).Get(context.Background())
	if err == nil {
		t.Fatal("expected an error from per-request override")
	}
	var he *HTTPError
	if !errors.As(err, &he) || he.StatusCode != 404 {
		t.Fatalf("expected *HTTPError with 404, got %v", err)
	}
	// Step 1 guarantees the response is still available alongside the error.
	if resp == nil || resp.StatusCode != 404 {
		t.Fatalf("expected non-nil 404 response, got %+v", resp)
	}
}

func TestRequestBuilder_NoOverrideUsesClientDefault(t *testing.T) {
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}, WithErrorOnStatus(false))
	defer ts.Close()

	_, err := c.Request().Path("/missing").Get(context.Background())
	if err != nil {
		t.Fatalf("expected client default (disabled) to apply, got %v", err)
	}
}

// --- Circuit Breaker ---

func TestCircuitBreakerOpensAfterThreshold(t *testing.T) {
	var calls atomic.Int32
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(500)
		fmt.Fprint(w, `{"error":"fail"}`)
	}, WithMaxRetries(0), WithCircuitBreaker(3, time.Second))
	defer ts.Close()

	// Trip the breaker: 3 failures.
	for i := 0; i < 3; i++ {
		c.Get(context.Background(), "/fail")
	}

	if c.cb.State() != StateOpen {
		t.Fatalf("expected circuit to be open, got %d", c.cb.State())
	}

	// Next call should be rejected by the breaker without hitting the server.
	before := calls.Load()
	_, err := c.Get(context.Background(), "/fail")
	if err == nil {
		t.Fatal("expected circuit breaker error")
	}
	if calls.Load() != before {
		t.Fatal("expected no server call when circuit is open")
	}
}

func TestCircuitBreakerTransitions(t *testing.T) {
	var shouldFail atomic.Bool
	shouldFail.Store(true)

	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if shouldFail.Load() {
			w.WriteHeader(500)
			fmt.Fprint(w, `{"error":"fail"}`)
			return
		}
		w.WriteHeader(200)
		fmt.Fprint(w, `{"ok":true}`)
	}, WithMaxRetries(0), WithCircuitBreaker(2, 50*time.Millisecond))
	defer ts.Close()

	// Trip breaker: closed -> open.
	for i := 0; i < 2; i++ {
		c.Get(context.Background(), "/x")
	}
	if c.cb.State() != StateOpen {
		t.Fatalf("expected open, got %d", c.cb.State())
	}

	// Wait for timeout -> half-open.
	time.Sleep(60 * time.Millisecond)
	shouldFail.Store(false)

	// This call should be allowed (half-open) and succeed.
	resp, err := c.Get(context.Background(), "/x")
	if err != nil {
		t.Fatalf("expected success in half-open, got %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	// Need threshold successes to close. We already have 1, need 1 more.
	_, err = c.Get(context.Background(), "/x")
	if err != nil {
		t.Fatal(err)
	}
	if c.cb.State() != StateClosed {
		t.Fatalf("expected closed after successes, got %d", c.cb.State())
	}
}

// --- MockClient ---

func TestMockClientImplementsInterface(t *testing.T) {
	var _ HTTPClient = NewMockClient()
}

func TestMockClientResponses(t *testing.T) {
	mc := NewMockClient()
	mc.OnGet("/users", 200, []byte(`[{"id":1}]`))
	mc.OnPost("/users", 201, []byte(`{"id":2}`))
	mc.OnPut("/users/1", 200, []byte(`{"id":1}`))
	mc.OnPatch("/users/1", 200, []byte(`{"id":1}`))
	mc.OnDelete("/users/1", 204, nil)

	ctx := context.Background()

	resp, err := mc.Get(ctx, "/users")
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("get: err=%v status=%d", err, resp.StatusCode)
	}

	resp, err = mc.Post(ctx, "/users", nil)
	if err != nil || resp.StatusCode != 201 {
		t.Fatalf("post: err=%v status=%d", err, resp.StatusCode)
	}

	resp, err = mc.Put(ctx, "/users/1", nil)
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("put: err=%v status=%d", err, resp.StatusCode)
	}

	resp, err = mc.Patch(ctx, "/users/1", nil)
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("patch: err=%v status=%d", err, resp.StatusCode)
	}

	resp, err = mc.Delete(ctx, "/users/1")
	if err != nil || resp.StatusCode != 204 {
		t.Fatalf("delete: err=%v status=%d", err, resp.StatusCode)
	}

	if mc.GetCallCount() != 5 {
		t.Fatalf("expected 5 calls, got %d", mc.GetCallCount())
	}
}

func TestMockClientError(t *testing.T) {
	mc := NewMockClient()
	mc.OnError("GET", "/fail", fmt.Errorf("boom"))

	_, err := mc.Get(context.Background(), "/fail")
	if err == nil || err.Error() != "boom" {
		t.Fatalf("expected boom error, got %v", err)
	}
}

func TestMockClientDefault404(t *testing.T) {
	mc := NewMockClient()
	resp, err := mc.Get(context.Background(), "/unknown")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 404 {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestMockClientReset(t *testing.T) {
	mc := NewMockClient()
	mc.OnGet("/x", 200, nil)
	mc.Get(context.Background(), "/x")
	mc.Reset()

	if mc.GetCallCount() != 0 {
		t.Fatal("expected 0 calls after reset")
	}
	resp, _ := mc.Get(context.Background(), "/x")
	if resp.StatusCode != 404 {
		t.Fatal("expected 404 after reset")
	}
}

// --- Functional Options ---

func TestFunctionalOptions(t *testing.T) {
	c := New("http://example.com",
		WithTimeout(5*time.Second),
		WithMaxRetries(1),
		WithRetryDelay(2*time.Second),
		WithMaxRetryDelay(8*time.Second),
		WithLogger(slog.Default()),
	)

	if c.timeout != 5*time.Second {
		t.Fatalf("expected timeout 5s, got %v", c.timeout)
	}
	if c.maxRetries != 1 {
		t.Fatalf("expected maxRetries 1, got %d", c.maxRetries)
	}
	if c.retryDelay != 2*time.Second {
		t.Fatalf("expected retryDelay 2s, got %v", c.retryDelay)
	}
	if c.maxRetryDelay != 8*time.Second {
		t.Fatalf("expected maxRetryDelay 8s, got %v", c.maxRetryDelay)
	}
	if c.httpClient.Timeout != 5*time.Second {
		t.Fatalf("expected http client timeout 5s, got %v", c.httpClient.Timeout)
	}
}

func TestWithTransport(t *testing.T) {
	custom := &http.Transport{MaxIdleConns: 42}
	c := New("http://example.com", WithTransport(custom))
	if c.httpClient.Transport != custom {
		t.Fatal("expected custom transport")
	}
}

// --- Concurrency ---

func TestConcurrentSetHeader(t *testing.T) {
	c := New("http://example.com",
		WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))),
	)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			c.SetHeader(fmt.Sprintf("X-Key-%d", n), "val")
		}(i)
	}
	wg.Wait()

	c.mu.RLock()
	count := len(c.headers)
	c.mu.RUnlock()
	if count != 100 {
		t.Fatalf("expected 100 headers, got %d", count)
	}
}
