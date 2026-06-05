package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/KARTIKrocks/apikit/errors"
	"github.com/KARTIKrocks/apikit/response"
)

// RecoverConfig configures the Recover middleware.
type RecoverConfig struct {
	// Logger is used to log panic details.
	// If nil, uses slog.Default().
	Logger *slog.Logger

	// EnableStackTrace includes the stack trace in logs.
	// Default: true
	EnableStackTrace bool

	// OnPanic is called when a panic is recovered.
	// If nil, a default handler is used.
	OnPanic func(w http.ResponseWriter, r *http.Request, recovered any)
}

// DefaultRecoverConfig returns the default configuration.
func DefaultRecoverConfig() RecoverConfig {
	return RecoverConfig{
		Logger:           nil,
		EnableStackTrace: true,
	}
}

// Recover recovers from panics and returns a 500 JSON response.
// Internal error details are logged but NOT exposed to clients.
func Recover() Middleware {
	return RecoverWithConfig(DefaultRecoverConfig())
}

// RecoverWithConfig recovers from panics with custom configuration.
func RecoverWithConfig(cfg RecoverConfig) Middleware {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					// Custom handler
					if cfg.OnPanic != nil {
						cfg.OnPanic(w, r, rec)
						return
					}

					// Log the panic
					attrs := []any{
						slog.String("method", r.Method),
						slog.String("path", r.URL.Path),
						slog.String("panic", fmt.Sprintf("%v", rec)),
					}

					if reqID := GetRequestID(r.Context()); reqID != "" {
						attrs = append(attrs, slog.String("request_id", reqID))
					}

					if cfg.EnableStackTrace {
						attrs = append(attrs, slog.String("stack", string(debug.Stack())))
					}

					logger.Error("Panic recovered", attrs...)

					// Send generic error — never expose panic details to client
					response.Err(w, errors.Internal("An unexpected error occurred"))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
