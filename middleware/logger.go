package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// LoggerConfig configures the Logger middleware.
type LoggerConfig struct {
	// Logger is the slog.Logger to use. Default: slog.Default()
	Logger *slog.Logger

	// Level is the default log level for successful requests. Default: slog.LevelInfo
	Level slog.Level

	// SkipPaths is a set of paths to skip logging for (e.g., health checks).
	SkipPaths map[string]bool

	// LogRequestBody logs the request body (use with caution — PII concerns).
	// Default: false
	LogRequestBody bool

	// LogResponseBody logs the response body.
	// Default: false
	LogResponseBody bool
}

// DefaultLoggerConfig returns the default logger configuration.
func DefaultLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Level:     slog.LevelInfo,
		SkipPaths: map[string]bool{"/health": true, "/ready": true, "/healthz": true},
	}
}

// Logger logs each HTTP request with structured fields.
// Logs: method, path, status, duration, request_id, client_ip, user_agent.
func Logger(logger *slog.Logger) Middleware {
	cfg := DefaultLoggerConfig()
	cfg.Logger = logger
	return LoggerWithConfig(cfg)
}

// LoggerWithConfig creates a logging middleware with custom configuration.
func LoggerWithConfig(cfg LoggerConfig) Middleware {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip configured paths
			if cfg.SkipPaths[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()

			// Wrap response writer to capture status code
			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(rw, r)

			duration := time.Since(start)

			attrs := []any{
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rw.statusCode),
				slog.Duration("duration", duration),
				slog.Int("bytes", rw.bytesWritten),
			}

			// Add request ID if present
			if reqID := GetRequestID(r.Context()); reqID != "" {
				attrs = append(attrs, slog.String("request_id", reqID))
			}

			// Add query string if present
			if r.URL.RawQuery != "" {
				attrs = append(attrs, slog.String("query", r.URL.RawQuery))
			}

			attrs = append(attrs,
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
			)

			// Log at appropriate level based on status code
			switch {
			case rw.statusCode >= 500:
				logger.Error("HTTP request", attrs...)
			case rw.statusCode >= 400:
				logger.Warn("HTTP request", attrs...)
			default:
				logger.Log(r.Context(), cfg.Level, "HTTP request", attrs...)
			}
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture the status code and bytes written.
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
	written      bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
	}
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.written = true
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}

// Unwrap returns the underlying ResponseWriter.
// This is needed for http.Flusher, http.Hijacker, etc.
func (rw *responseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

// Flush implements http.Flusher if the underlying writer supports it.
func (rw *responseWriter) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
