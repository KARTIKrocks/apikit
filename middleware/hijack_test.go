package middleware

import (
	"bufio"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// fakeHijacker is an http.ResponseWriter that also supports Hijack, which
// httptest.ResponseRecorder does not.
type fakeHijacker struct {
	http.ResponseWriter
	hijacked bool
}

func (f *fakeHijacker) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	f.hijacked = true
	c1, c2 := net.Pipe()
	c2.Close()
	return c1, bufio.NewReadWriter(bufio.NewReader(c1), bufio.NewWriter(c1)), nil
}

// assertHijackable runs a handler behind mw and verifies the handler's
// ResponseWriter still satisfies http.Hijacker and forwards to the underlying
// writer — i.e. WebSocket upgrades survive the middleware.
func assertHijackable(t *testing.T, mw Middleware, name string) {
	t.Helper()
	var sawHijacker bool
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		sawHijacker = ok
		if ok {
			conn, _, err := hj.Hijack()
			if err != nil {
				t.Errorf("%s: hijack failed: %v", name, err)
			} else {
				conn.Close()
			}
		}
	}))

	f := &fakeHijacker{ResponseWriter: httptest.NewRecorder()}
	h.ServeHTTP(f, httptest.NewRequest("GET", "/ws", nil))

	if !sawHijacker {
		t.Fatalf("%s: handler ResponseWriter does not implement http.Hijacker", name)
	}
	if !f.hijacked {
		t.Fatalf("%s: Hijack did not reach the underlying ResponseWriter", name)
	}
}

func TestLoggerPreservesHijacker(t *testing.T) {
	assertHijackable(t, Logger(slog.New(slog.NewTextHandler(io.Discard, nil))), "Logger")
}

func TestTimeoutPreservesHijacker(t *testing.T) {
	assertHijackable(t, Timeout(time.Second), "Timeout")
}
