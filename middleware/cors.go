package middleware

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// CORSConfig configures the CORS middleware.
type CORSConfig struct {
	// AllowOrigins is a list of allowed origins.
	// Use "*" to allow all origins (not recommended for production with credentials).
	// Default: []
	AllowOrigins []string

	// AllowMethods is a list of allowed HTTP methods.
	// Default: GET, POST, PUT, PATCH, DELETE, OPTIONS
	AllowMethods []string

	// AllowHeaders is a list of allowed request headers.
	// Default: Content-Type, Authorization, X-Request-ID
	AllowHeaders []string

	// ExposeHeaders is a list of headers the browser is allowed to access.
	// Default: X-Request-ID
	ExposeHeaders []string

	// AllowCredentials indicates whether credentials (cookies, auth headers) are allowed.
	// Default: false
	AllowCredentials bool

	// MaxAge is how long preflight results can be cached.
	// Default: 12 hours
	MaxAge time.Duration
}

// DefaultCORSConfig returns a permissive CORS configuration for development.
// For production, you should specify AllowOrigins explicitly.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization", "X-Request-ID", "X-API-Key"},
		ExposeHeaders: []string{"X-Request-ID"},
		MaxAge:        12 * time.Hour,
	}
}

// CORS adds Cross-Origin Resource Sharing headers.
func CORS(cfg CORSConfig) Middleware {
	allowAll := len(cfg.AllowOrigins) == 1 && cfg.AllowOrigins[0] == "*"

	// Warn about invalid CORS configuration per spec:
	// credentials cannot be used with wildcard origin.
	if allowAll && cfg.AllowCredentials {
		slog.Warn("apikit/middleware: CORS misconfiguration — AllowCredentials is true with wildcard origin. " +
			"The CORS spec forbids Access-Control-Allow-Credentials with Access-Control-Allow-Origin: *. " +
			"Requests will use the specific Origin header instead of *, but you should list explicit origins.")
	}
	originsSet := make(map[string]bool, len(cfg.AllowOrigins))
	for _, o := range cfg.AllowOrigins {
		originsSet[strings.ToLower(o)] = true
	}

	methods := strings.Join(cfg.AllowMethods, ", ")
	headers := strings.Join(cfg.AllowHeaders, ", ")
	exposed := strings.Join(cfg.ExposeHeaders, ", ")
	maxAge := strconv.Itoa(int(cfg.MaxAge.Seconds()))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// No Origin header — not a CORS request
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Check if origin is allowed
			allowed := allowAll || originsSet[strings.ToLower(origin)]
			if !allowed {
				next.ServeHTTP(w, r)
				return
			}

			// Set the actual origin (not "*") when credentials are enabled
			if cfg.AllowCredentials || !allowAll {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}

			if cfg.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if exposed != "" {
				w.Header().Set("Access-Control-Expose-Headers", exposed)
			}

			// Handle preflight
			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", methods)
				w.Header().Set("Access-Control-Allow-Headers", headers)
				w.Header().Set("Access-Control-Max-Age", maxAge)
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
