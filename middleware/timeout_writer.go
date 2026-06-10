package middleware

import (
	"bufio"
	"errors"
	"net"
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

// Unwrap returns the underlying ResponseWriter, so http.ResponseController and
// interface probes can reach it through the wrapper.
func (tw *timeoutWriter) Unwrap() http.ResponseWriter {
	return tw.ResponseWriter
}

// Flush implements http.Flusher if the underlying writer supports it.
func (tw *timeoutWriter) Flush() {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if tw.timedOut {
		return
	}
	if f, ok := tw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Hijack implements http.Hijacker by delegating to the underlying writer.
// Note: the Timeout middleware is unsuitable for long-lived hijacked
// connections (e.g. WebSockets) — its timer keeps running and will try to write
// a 503 to the taken-over connection. Mount such routes without Timeout. Once a
// timeout has fired the connection can no longer be hijacked.
func (tw *timeoutWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if tw.timedOut {
		return nil, nil, http.ErrHandlerTimeout
	}
	if hj, ok := tw.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, errors.New("apikit/middleware: underlying ResponseWriter does not implement http.Hijacker")
}
