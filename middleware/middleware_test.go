package middleware

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/KARTIKrocks/apikit/errors"
)

// --- RequestID tests ---

func TestRequestID_GeneratesID(t *testing.T) {
	handler := RequestID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := GetRequestID(r.Context())
		if id == "" {
			t.Error("expected request ID in context")
		}
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Header().Get("X-Request-ID") == "" {
		t.Error("expected X-Request-ID header in response")
	}
}

func TestRequestID_TrustsProxy(t *testing.T) {
	handler := RequestID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := GetRequestID(r.Context())
		if id != "incoming-id" {
			t.Errorf("expected incoming-id, got %q", id)
		}
	}))

	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-Request-ID", "incoming-id")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Header().Get("X-Request-ID") != "incoming-id" {
		t.Errorf("expected incoming-id in response header, got %q", w.Header().Get("X-Request-ID"))
	}
}

func TestRequestID_NoTrustProxy(t *testing.T) {
	cfg := DefaultRequestIDConfig()
	cfg.TrustProxy = false
	handler := RequestIDWithConfig(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := GetRequestID(r.Context())
		if id == "incoming-id" {
			t.Error("should not trust proxy header")
		}
		if id == "" {
			t.Error("should generate a new ID")
		}
	}))

	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-Request-ID", "incoming-id")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
}

func TestGetRequestID_EmptyContext(t *testing.T) {
	id := GetRequestID(context.Background())
	if id != "" {
		t.Errorf("expected empty string, got %q", id)
	}
}

// --- Logger tests ---

func TestLogger_LogsRequest(t *testing.T) {
	var buf strings.Builder
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	handler := Logger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello"))
	}))

	r := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	log := buf.String()
	if !strings.Contains(log, "GET") {
		t.Error("log should contain method")
	}
	if !strings.Contains(log, "/test") {
		t.Error("log should contain path")
	}
	if !strings.Contains(log, "200") {
		t.Error("log should contain status code")
	}
}

func TestLogger_SkipPaths(t *testing.T) {
	var buf strings.Builder
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	handler := Logger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if buf.Len() > 0 {
		t.Error("health path should be skipped")
	}
}

func TestLogger_ErrorLevel(t *testing.T) {
	var buf strings.Builder
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError}))

	handler := Logger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if !strings.Contains(buf.String(), "ERROR") {
		t.Error("5xx should log at error level")
	}
}

func TestResponseWriter_WriteDefaultsStatus(t *testing.T) {
	rw := &responseWriter{ResponseWriter: httptest.NewRecorder(), statusCode: http.StatusOK}

	// Write without WriteHeader should still track bytes.
	_, _ = rw.Write([]byte("test"))
	if rw.bytesWritten != 4 {
		t.Errorf("expected 4 bytes written, got %d", rw.bytesWritten)
	}
}

func TestResponseWriter_DoubleWriteHeader(t *testing.T) {
	inner := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: inner, statusCode: http.StatusOK}

	rw.WriteHeader(http.StatusCreated)
	rw.WriteHeader(http.StatusBadRequest) // should be ignored

	if rw.statusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", rw.statusCode)
	}
}

// --- Recover tests ---

func TestRecover_CatchesPanic(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	handler := RecoverWithConfig(RecoverConfig{Logger: logger})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	var envelope map[string]any
	if err := json.NewDecoder(w.Body).Decode(&envelope); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if envelope["success"] != false {
		t.Error("expected success=false")
	}
}

func TestRecover_CustomOnPanic(t *testing.T) {
	var called bool
	handler := RecoverWithConfig(RecoverConfig{
		OnPanic: func(w http.ResponseWriter, r *http.Request, recovered any) {
			called = true
			w.WriteHeader(http.StatusServiceUnavailable)
		},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if !called {
		t.Error("OnPanic should be called")
	}
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
}

func TestRecover_NoPanic(t *testing.T) {
	handler := Recover()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// --- Timeout tests ---

func TestTimeout_CompletesInTime(t *testing.T) {
	handler := Timeout(time.Second)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestTimeout_Expires(t *testing.T) {
	handler := Timeout(50 * time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(5 * time.Second):
		}
	}))

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusGatewayTimeout {
		t.Errorf("expected 504, got %d", w.Code)
	}
}

func TestTimeoutWriter_Flush(t *testing.T) {
	inner := httptest.NewRecorder()
	tw := &timeoutWriter{ResponseWriter: inner, done: make(chan struct{})}

	// Flush should not panic.
	tw.Flush()
}

func TestTimeoutWriter_Unwrap(t *testing.T) {
	inner := httptest.NewRecorder()
	tw := &timeoutWriter{ResponseWriter: inner, done: make(chan struct{})}

	if tw.Unwrap() != inner {
		t.Error("Unwrap should return inner writer")
	}
}

func TestTimeoutWriter_BlocksAfterTimeout(t *testing.T) {
	inner := httptest.NewRecorder()
	tw := &timeoutWriter{ResponseWriter: inner, done: make(chan struct{})}
	tw.timedOut = true

	tw.WriteHeader(http.StatusOK) // should be suppressed

	n, err := tw.Write([]byte("test"))
	if n != 0 {
		t.Errorf("expected 0 bytes written, got %d", n)
	}
	if err != http.ErrHandlerTimeout {
		t.Errorf("expected ErrHandlerTimeout, got %v", err)
	}

	tw.Flush() // should not panic
}

// --- Auth tests ---

func TestAuth_BearerToken(t *testing.T) {
	handler := Auth(AuthConfig{
		Authenticate: func(ctx context.Context, token string) (any, error) {
			if token == "valid-token" {
				return "user-1", nil
			}
			return nil, errors.Unauthorized("Invalid token")
		},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := GetAuthUserAs[string](r.Context())
		if !ok || user != "user-1" {
			t.Errorf("expected user-1, got %v", user)
		}
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAuth_MissingToken(t *testing.T) {
	handler := Auth(AuthConfig{
		Authenticate: func(ctx context.Context, token string) (any, error) {
			return nil, nil
		},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuth_SkipPaths(t *testing.T) {
	handler := Auth(AuthConfig{
		Authenticate: func(ctx context.Context, token string) (any, error) {
			return nil, errors.Unauthorized("no")
		},
		SkipPaths: map[string]bool{"/health": true},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for skipped path, got %d", w.Code)
	}
}

func TestAuth_APIKey(t *testing.T) {
	handler := Auth(AuthConfig{
		Scheme: "api-key",
		Authenticate: func(ctx context.Context, key string) (any, error) {
			if key == "secret" {
				return "api-user", nil
			}
			return nil, errors.Unauthorized("bad key")
		},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-API-Key", "secret")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAuth_InvalidBearerFormat(t *testing.T) {
	handler := Auth(AuthConfig{
		Authenticate: func(ctx context.Context, token string) (any, error) {
			return nil, nil
		},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid bearer format, got %d", w.Code)
	}
}

func TestAuth_PanicsWithoutAuthenticate(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when Authenticate is nil")
		}
	}()
	Auth(AuthConfig{})
}

func TestGetAuthUser_Nil(t *testing.T) {
	user := GetAuthUser(context.Background())
	if user != nil {
		t.Errorf("expected nil, got %v", user)
	}
}

func TestGetAuthUserAs_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), authUserKey{}, 42)
	_, ok := GetAuthUserAs[string](ctx)
	if ok {
		t.Error("expected ok=false for wrong type")
	}
}

func TestRequireRole_Allowed(t *testing.T) {
	mw := RequireRole("admin", func(ctx context.Context) []string {
		return []string{"user", "admin"}
	})

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRequireRole_Denied(t *testing.T) {
	mw := RequireRole("admin", func(ctx context.Context) []string {
		return []string{"user"}
	})

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

// --- CORS tests ---

func TestCORS_AllowAll(t *testing.T) {
	handler := CORS(DefaultCORSConfig())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected *, got %q", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORS_SpecificOrigin(t *testing.T) {
	cfg := CORSConfig{
		AllowOrigins: []string{"https://example.com"},
		AllowMethods: []string{"GET"},
		AllowHeaders: []string{"Content-Type"},
	}
	handler := CORS(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Errorf("expected specific origin, got %q", w.Header().Get("Access-Control-Allow-Origin"))
	}
	if w.Header().Get("Vary") != "Origin" {
		t.Error("expected Vary: Origin header")
	}
}

func TestCORS_DisallowedOrigin(t *testing.T) {
	cfg := CORSConfig{
		AllowOrigins: []string{"https://allowed.com"},
		AllowMethods: []string{"GET"},
	}
	handler := CORS(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Origin", "https://evil.com")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("disallowed origin should not get CORS headers")
	}
}

func TestCORS_Preflight(t *testing.T) {
	handler := CORS(DefaultCORSConfig())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for preflight")
	}))

	r := httptest.NewRequest("OPTIONS", "/", nil)
	r.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204 for preflight, got %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("preflight should include Allow-Methods")
	}
}

func TestCORS_NoOrigin(t *testing.T) {
	var handlerCalled bool
	handler := CORS(DefaultCORSConfig())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if !handlerCalled {
		t.Error("handler should be called for non-CORS request")
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("non-CORS request should not get CORS headers")
	}
}

func TestCORS_Credentials(t *testing.T) {
	cfg := CORSConfig{
		AllowOrigins:     []string{"https://example.com"},
		AllowMethods:     []string{"GET"},
		AllowCredentials: true,
	}
	handler := CORS(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Error("expected credentials header")
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Error("credentials should use specific origin, not *")
	}
}

// --- RateLimit tests ---

func TestRateLimit_AllowsUnderLimit(t *testing.T) {
	cfg := RateLimitConfig{Rate: 5, Window: time.Minute}
	handler := RateLimit(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 5; i++ {
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = "1.2.3.4:1234"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i, w.Code)
		}
	}
}

func TestRateLimit_BlocksOverLimit(t *testing.T) {
	cfg := RateLimitConfig{Rate: 2, Window: time.Minute}
	handler := RateLimit(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 3; i++ {
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = "1.2.3.4:1234"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		if i < 2 && w.Code != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i, w.Code)
		}
		if i == 2 {
			if w.Code != http.StatusTooManyRequests {
				t.Errorf("request %d: expected 429, got %d", i, w.Code)
			}
			if w.Header().Get("Retry-After") == "" {
				t.Error("expected Retry-After header")
			}
		}
	}
}

func TestRateLimit_PerClient(t *testing.T) {
	cfg := RateLimitConfig{Rate: 1, Window: time.Minute}
	handler := RateLimit(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First client uses their quota.
	r1 := httptest.NewRequest("GET", "/", nil)
	r1.RemoteAddr = "1.1.1.1:1234"
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, r1)
	if w1.Code != http.StatusOK {
		t.Errorf("client 1: expected 200, got %d", w1.Code)
	}

	// Second client should have their own quota.
	r2 := httptest.NewRequest("GET", "/", nil)
	r2.RemoteAddr = "2.2.2.2:1234"
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, r2)
	if w2.Code != http.StatusOK {
		t.Errorf("client 2: expected 200, got %d", w2.Code)
	}
}

func TestFixedWindow_Stop(t *testing.T) {
	fw := NewFixedWindow(10, time.Minute)
	fw.Stop() // should not panic or block
}

func TestRateLimit_TrustProxy(t *testing.T) {
	cfg := RateLimitConfig{Rate: 1, Window: time.Minute, TrustProxy: true}
	handler := RateLimit(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First request from proxied IP.
	r1 := httptest.NewRequest("GET", "/", nil)
	r1.RemoteAddr = "10.0.0.1:1234"
	r1.Header.Set("X-Forwarded-For", "203.0.113.1, 10.0.0.1")
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, r1)
	if w1.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w1.Code)
	}

	// Second request, same client IP, should be limited.
	r2 := httptest.NewRequest("GET", "/", nil)
	r2.RemoteAddr = "10.0.0.1:1234"
	r2.Header.Set("X-Forwarded-For", "203.0.113.1, 10.0.0.1")
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, r2)
	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", w2.Code)
	}
}

// --- Security Headers tests ---

func TestSecureHeaders_Default(t *testing.T) {
	handler := SecureHeaders()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	checks := map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"X-XSS-Protection":          "1; mode=block",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
	}
	for header, expected := range checks {
		got := w.Header().Get(header)
		if got != expected {
			t.Errorf("%s: expected %q, got %q", header, expected, got)
		}
	}
}

func TestSecureHeaders_Custom(t *testing.T) {
	cfg := SecurityHeadersConfig{
		ContentTypeNosniff:    true,
		XFrameOptions:         "SAMEORIGIN",
		ContentSecurityPolicy: "default-src 'self'",
	}
	handler := SecureHeadersWithConfig(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Header().Get("X-Frame-Options") != "SAMEORIGIN" {
		t.Errorf("expected SAMEORIGIN, got %q", w.Header().Get("X-Frame-Options"))
	}
	if w.Header().Get("Content-Security-Policy") != "default-src 'self'" {
		t.Error("expected CSP header")
	}
	if w.Header().Get("Strict-Transport-Security") != "" {
		t.Error("HSTS should not be set when maxAge=0")
	}
}

// --- BodyLimit tests ---

func TestBodyLimit_UnderLimit(t *testing.T) {
	handler := BodyLimit(100)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest("POST", "/", strings.NewReader("small body"))
	r.ContentLength = 10
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestBodyLimit_OverLimit(t *testing.T) {
	handler := BodyLimit(10)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	r := httptest.NewRequest("POST", "/", strings.NewReader("this body is way too large for the limit"))
	r.ContentLength = 40
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413, got %d", w.Code)
	}
}

// --- Chain tests ---

func TestChain_Order(t *testing.T) {
	var order []string

	mw := func(name string) Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name+"-before")
				next.ServeHTTP(w, r)
				order = append(order, name+"-after")
			})
		}
	}

	handler := Chain(mw("first"), mw("second"), mw("third"))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	}))

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	expected := []string{"first-before", "second-before", "third-before", "handler", "third-after", "second-after", "first-after"}
	if len(order) != len(expected) {
		t.Fatalf("expected %d entries, got %d: %v", len(expected), len(order), order)
	}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("order[%d]: expected %q, got %q", i, v, order[i])
		}
	}
}

func TestThen(t *testing.T) {
	var called bool
	mw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			next.ServeHTTP(w, r)
		})
	}

	handler := Then(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}), mw)

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if !called {
		t.Error("middleware should have been called")
	}
}

// --- FixedWindow concurrent safety test ---

func TestFixedWindow_ConcurrentAccess(t *testing.T) {
	fw := NewFixedWindow(100, time.Minute)
	defer fw.Stop()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := "client"
			for j := 0; j < 10; j++ {
				fw.Allow(key)
				_ = n
			}
		}(i)
	}
	wg.Wait()
}
