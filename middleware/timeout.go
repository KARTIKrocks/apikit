package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/KARTIKrocks/apikit/errors"
	"github.com/KARTIKrocks/apikit/response"
)

// Timeout wraps each request with a context deadline.
// If the handler doesn't complete within the timeout, a 504 response is sent.
//
//	mux.Handle("/slow", middleware.Timeout(5*time.Second)(handler))
func Timeout(duration time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), duration)
			defer cancel()

			// Channel to signal handler completion
			done := make(chan struct{})

			// Wrap the writer to prevent writes after timeout
			tw := &timeoutWriter{
				ResponseWriter: w,
				done:           done,
			}

			go func() {
				defer close(done)
				next.ServeHTTP(tw, r.WithContext(ctx))
			}()

			select {
			case <-done:
				// Handler completed normally
			case <-ctx.Done():
				tw.mu.Lock()
				if !tw.written {
					tw.timedOut = true
					// Write directly to the underlying ResponseWriter while holding the lock.
					// The lock prevents the handler goroutine from writing concurrently.
					response.Err(tw.ResponseWriter, errors.Timeout("Request timed out"))
				}
				tw.mu.Unlock()
			}
		})
	}
}
