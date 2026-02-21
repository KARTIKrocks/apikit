package middleware

import (
	"net/http"
	"sync"
)

// timeoutWriter prevents writes after a timeout has occurred.
type timeoutWriter struct {
	http.ResponseWriter
	mu       sync.Mutex
	written  bool
	timedOut bool
	done     chan struct{}
}

func (tw *timeoutWriter) WriteHeader(code int) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if tw.timedOut {
		return
	}
	tw.written = true
	tw.ResponseWriter.WriteHeader(code)
}

func (tw *timeoutWriter) Write(b []byte) (int, error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if tw.timedOut {
		return 0, http.ErrHandlerTimeout
	}
	tw.written = true
	return tw.ResponseWriter.Write(b)
}
