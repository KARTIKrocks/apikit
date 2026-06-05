package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

type requestIDKey struct{}

// RequestIDConfig configures the RequestID middleware.
type RequestIDConfig struct {
	// Header is the header to read/write the request ID.
	// Default: "X-Request-ID"
	Header string

	// Generator is a function that generates a new request ID.
	// Default: generates a random 16-byte hex string.
	Generator func() string

	// TrustProxy trusts the incoming X-Request-ID header if present.
	// If false, always generates a new ID.
	// Default: true
	TrustProxy bool
}

// DefaultRequestIDConfig returns the default configuration.
func DefaultRequestIDConfig() RequestIDConfig {
	return RequestIDConfig{
		Header:     "X-Request-ID",
		Generator:  generateID,
		TrustProxy: true,
	}
}

// RequestID adds a unique request ID to each request.
// The ID is stored in the request context and set as a response header.
func RequestID() Middleware {
	return RequestIDWithConfig(DefaultRequestIDConfig())
}

// RequestIDWithConfig adds request IDs with custom configuration.
func RequestIDWithConfig(cfg RequestIDConfig) Middleware {
	if cfg.Header == "" {
		cfg.Header = "X-Request-ID"
	}
	if cfg.Generator == nil {
		cfg.Generator = generateID
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := ""
			if cfg.TrustProxy {
				id = r.Header.Get(cfg.Header)
			}
			if id == "" {
				id = cfg.Generator()
			}

			// Set on response
			w.Header().Set(cfg.Header, id)

			// Set in context
			ctx := context.WithValue(r.Context(), requestIDKey{}, id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetRequestID retrieves the request ID from the context.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		return id
	}
	return ""
}

// generateID generates a random 16-byte hex string.
func generateID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
