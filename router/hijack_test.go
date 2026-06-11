package router

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

// fakeHijacker is an http.ResponseWriter that also supports Hijack — something
// httptest.ResponseRecorder does not, so it lets us assert that the router's
// probeWriter forwards http.Hijacker to a capable underlying writer.
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

// TestProbeWriterHijack verifies a WebSocket-style handler can assert
// http.Hijacker directly through the probeWriter. gorilla/websocket asserts
// http.Hijacker rather than consulting Unwrap, so the method must be present on
// the wrapper itself.
func TestProbeWriterHijack(t *testing.T) {
	t.Parallel()
	r := New()
	var sawHijacker bool
	r.GetFunc("/ws", func(w http.ResponseWriter, req *http.Request) {
		hj, ok := w.(http.Hijacker)
		sawHijacker = ok
		if ok {
			if _, _, err := hj.Hijack(); err != nil {
				t.Errorf("hijack failed: %v", err)
			}
		}
	})

	f := &fakeHijacker{ResponseWriter: httptest.NewRecorder()}
	r.ServeHTTP(f, httptest.NewRequest("GET", "/ws", nil))

	if !sawHijacker {
		t.Fatal("handler's ResponseWriter does not implement http.Hijacker through probeWriter")
	}
	if !f.hijacked {
		t.Fatal("Hijack did not reach the underlying ResponseWriter")
	}
}

// TestProbeWriterHijackUnsupported confirms a clear error (not a panic) when the
// underlying writer cannot be hijacked.
func TestProbeWriterHijackUnsupported(t *testing.T) {
	r := New()
	var hijackErr error
	r.GetFunc("/ws", func(w http.ResponseWriter, req *http.Request) {
		if hj, ok := w.(http.Hijacker); ok {
			_, _, hijackErr = hj.Hijack()
		}
	})

	// httptest.ResponseRecorder is not an http.Hijacker.
	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/ws", nil))

	if hijackErr == nil {
		t.Fatal("expected an error hijacking a non-hijackable writer")
	}
}
