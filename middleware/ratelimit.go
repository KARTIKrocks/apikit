package middleware

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/KARTIKrocks/apikit/errors"
	"github.com/KARTIKrocks/apikit/response"
)

// RateLimiter defines the interface for rate limiting backends.
// Implement this interface to use Redis, memcached, etc.
type RateLimiter interface {
	// Allow checks if a request from the given key is allowed.
	// Returns true if allowed, false if rate limited.
	Allow(key string) bool
}

// RateLimitConfig configures the rate limiting middleware.
type RateLimitConfig struct {
	// Limiter is the rate limiting backend.
	// If nil, a default in-memory token bucket is used.
	Limiter RateLimiter

	// KeyFunc extracts the rate limit key from the request.
	// Default: uses client IP from RemoteAddr (safe).
	// Set TrustProxy to true if behind a reverse proxy.
	KeyFunc func(r *http.Request) string

	// TrustProxy enables reading client IP from X-Forwarded-For and X-Real-IP headers.
	// Only enable this if your server is behind a trusted reverse proxy.
	// Default: false (uses RemoteAddr only)
	TrustProxy bool

	// Rate is the number of requests allowed per window (for default limiter).
	// Default: 100
	Rate int

	// Window is the time window for the rate limit (for default limiter).
	// Default: 1 minute
	Window time.Duration

	// Message is the error message sent when rate limited.
	// Default: "Rate limit exceeded"
	Message string
}

// DefaultRateLimitConfig returns sensible rate limit defaults.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Rate:    100,
		Window:  time.Minute,
		Message: "Rate limit exceeded. Please try again later.",
	}
}

// RateLimit applies rate limiting per client.
func RateLimit(cfg RateLimitConfig) Middleware {
	if cfg.Rate <= 0 {
		cfg.Rate = 100
	}
	if cfg.Window <= 0 {
		cfg.Window = time.Minute
	}
	if cfg.Message == "" {
		cfg.Message = "Rate limit exceeded. Please try again later."
	}
	if cfg.KeyFunc == nil {
		cfg.KeyFunc = makeKeyFunc(cfg.TrustProxy)
	}
	if cfg.Limiter == nil {
		cfg.Limiter = NewTokenBucket(cfg.Rate, cfg.Window)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := cfg.KeyFunc(r)

			if !cfg.Limiter.Allow(key) {
				w.Header().Set("Retry-After", strconv.Itoa(int(cfg.Window.Seconds())))
				response.Err(w, errors.RateLimited(cfg.Message))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// makeKeyFunc returns a key extraction function based on proxy trust setting.
func makeKeyFunc(trustProxy bool) func(r *http.Request) string {
	return func(r *http.Request) string {
		if trustProxy {
			if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
				// Take only the first IP (the original client), not the full chain
				if first, _, ok := strings.Cut(xff, ","); ok {
					return strings.TrimSpace(first)
				}
				return strings.TrimSpace(xff)
			}
			if xri := r.Header.Get("X-Real-IP"); xri != "" {
				return strings.TrimSpace(xri)
			}
		}
		// Use net.SplitHostPort to handle IPv6 correctly
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			return r.RemoteAddr
		}
		return host
	}
}

// --- In-memory token bucket implementation ---

// TokenBucket implements a simple in-memory token bucket rate limiter.
type TokenBucket struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	rate    int
	window  time.Duration
	stop    chan struct{}
}

type bucket struct {
	tokens    int
	lastReset time.Time
}

// NewTokenBucket creates a new in-memory token bucket rate limiter.
// Call Stop() when the limiter is no longer needed to release the cleanup goroutine.
func NewTokenBucket(rate int, window time.Duration) *TokenBucket {
	tb := &TokenBucket{
		buckets: make(map[string]*bucket),
		rate:    rate,
		window:  window,
		stop:    make(chan struct{}),
	}

	// Start cleanup goroutine
	go tb.cleanup()

	return tb
}

// Stop terminates the background cleanup goroutine.
// The TokenBucket should not be used after calling Stop.
func (tb *TokenBucket) Stop() {
	close(tb.stop)
}

// Allow checks if a request is allowed for the given key.
func (tb *TokenBucket) Allow(key string) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	b, exists := tb.buckets[key]

	if !exists || now.Sub(b.lastReset) >= tb.window {
		tb.buckets[key] = &bucket{
			tokens:    tb.rate - 1,
			lastReset: now,
		}
		return true
	}

	if b.tokens <= 0 {
		return false
	}

	b.tokens--
	return true
}

// cleanup periodically removes expired buckets.
func (tb *TokenBucket) cleanup() {
	ticker := time.NewTicker(tb.window * 2)
	defer ticker.Stop()

	for {
		select {
		case <-tb.stop:
			return
		case <-ticker.C:
			tb.mu.Lock()
			now := time.Now()
			for key, b := range tb.buckets {
				if now.Sub(b.lastReset) >= tb.window*2 {
					delete(tb.buckets, key)
				}
			}
			tb.mu.Unlock()
		}
	}
}
