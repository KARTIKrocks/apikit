package response

import (
	"bytes"
	stderrors "errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/KARTIKrocks/apikit/errors"
)

var content = []byte("0123456789abcdefghij") // 20 bytes

func serve(t *testing.T, req *http.Request, cfg ContentConfig) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	ServeContent(w, req, bytes.NewReader(content), cfg)
	return w
}

func TestServeContent_FullBody(t *testing.T) {
	w := serve(t, httptest.NewRequest(http.MethodGet, "/clip", nil), ContentConfig{})

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if got := w.Body.String(); got != string(content) {
		t.Fatalf("body = %q, want %q", got, content)
	}
	if got := w.Header().Get("Accept-Ranges"); got != "bytes" {
		t.Fatalf("Accept-Ranges = %q, want bytes", got)
	}
}

func TestServeContent_Range(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clip", nil)
	req.Header.Set("Range", "bytes=5-9")
	w := serve(t, req, ContentConfig{})

	if w.Code != http.StatusPartialContent {
		t.Fatalf("status = %d, want 206", w.Code)
	}
	if got := w.Body.String(); got != "56789" {
		t.Fatalf("body = %q, want 56789", got)
	}
	if got := w.Header().Get("Content-Range"); got != "bytes 5-9/20" {
		t.Fatalf("Content-Range = %q, want bytes 5-9/20", got)
	}
	if got := w.Header().Get("Content-Length"); got != "5" {
		t.Fatalf("Content-Length = %q, want 5", got)
	}
}

func TestServeContent_OpenEndedRange(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clip", nil)
	req.Header.Set("Range", "bytes=15-")
	w := serve(t, req, ContentConfig{})

	if w.Code != http.StatusPartialContent {
		t.Fatalf("status = %d, want 206", w.Code)
	}
	if got := w.Body.String(); got != "fghij" {
		t.Fatalf("body = %q, want fghij", got)
	}
}

func TestServeContent_SuffixRange(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clip", nil)
	req.Header.Set("Range", "bytes=-4")
	w := serve(t, req, ContentConfig{})

	if got := w.Body.String(); got != "ghij" {
		t.Fatalf("body = %q, want ghij", got)
	}
}

func TestServeContent_UnsatisfiableRange(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clip", nil)
	req.Header.Set("Range", "bytes=100-200")
	w := serve(t, req, ContentConfig{})

	if w.Code != http.StatusRequestedRangeNotSatisfiable {
		t.Fatalf("status = %d, want 416", w.Code)
	}
	if got := w.Header().Get("Content-Range"); got != "bytes */20" {
		t.Fatalf("Content-Range = %q, want bytes */20", got)
	}
}

func TestServeContent_HeadSendsNoBody(t *testing.T) {
	w := serve(t, httptest.NewRequest(http.MethodHead, "/clip", nil), ContentConfig{})

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if w.Body.Len() != 0 {
		t.Fatalf("body = %q, want empty", w.Body.String())
	}
	if got := w.Header().Get("Content-Length"); got != "20" {
		t.Fatalf("Content-Length = %q, want 20", got)
	}
}

func TestServeContent_NotModifiedByETag(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clip", nil)
	req.Header.Set("If-None-Match", `"v1"`)
	w := serve(t, req, ContentConfig{ETag: "v1"})

	if w.Code != http.StatusNotModified {
		t.Fatalf("status = %d, want 304", w.Code)
	}
	if w.Body.Len() != 0 {
		t.Fatalf("body = %q, want empty", w.Body.String())
	}
}

// An unquoted ETag would never match a client's If-None-Match, silently
// disabling the caching the caller asked for.
func TestServeContent_ETagIsQuoted(t *testing.T) {
	w := serve(t, httptest.NewRequest(http.MethodGet, "/clip", nil), ContentConfig{ETag: "v1"})

	if got := w.Header().Get("Etag"); got != `"v1"` {
		t.Fatalf("Etag = %q, want %q", got, `"v1"`)
	}
}

func TestServeContent_NotModifiedByModTime(t *testing.T) {
	modTime := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	req := httptest.NewRequest(http.MethodGet, "/clip", nil)
	req.Header.Set("If-Modified-Since", modTime.Format(http.TimeFormat))
	w := serve(t, req, ContentConfig{ModTime: modTime})

	if w.Code != http.StatusNotModified {
		t.Fatalf("status = %d, want 304", w.Code)
	}
}

// If-Range on a stale validator must yield the whole body, not the range.
func TestServeContent_IfRangeMismatchSendsFullBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clip", nil)
	req.Header.Set("Range", "bytes=5-9")
	req.Header.Set("If-Range", `"stale"`)
	w := serve(t, req, ContentConfig{ETag: "v1"})

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if got := w.Body.String(); got != string(content) {
		t.Fatalf("body = %q, want full content", got)
	}
}

func TestServeContent_ExplicitContentType(t *testing.T) {
	w := serve(t, httptest.NewRequest(http.MethodGet, "/clip", nil), ContentConfig{
		Filename:    "clip.mp4",
		ContentType: "video/webm",
	})

	if got := w.Header().Get("Content-Type"); got != "video/webm" {
		t.Fatalf("Content-Type = %q, want video/webm", got)
	}
}

func TestServeContent_ContentTypeFromFilename(t *testing.T) {
	w := serve(t, httptest.NewRequest(http.MethodGet, "/clip", nil), ContentConfig{
		Filename: "clip.mp4",
	})

	if got := w.Header().Get("Content-Type"); !strings.HasPrefix(got, "video/mp4") {
		t.Fatalf("Content-Type = %q, want video/mp4", got)
	}
}

func TestServeContent_Disposition(t *testing.T) {
	tests := []struct {
		name   string
		cfg    ContentConfig
		want   string
		absent bool
	}{
		{name: "attachment", cfg: ContentConfig{Filename: "a.mp4"}, want: `attachment; filename="a.mp4"; filename*=UTF-8''a.mp4`},
		{name: "inline", cfg: ContentConfig{Filename: "a.mp4", Inline: true}, want: `inline; filename="a.mp4"; filename*=UTF-8''a.mp4`},
		{name: "none", cfg: ContentConfig{}, absent: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := serve(t, httptest.NewRequest(http.MethodGet, "/clip", nil), tt.cfg)
			got := w.Header().Get("Content-Disposition")
			if tt.absent {
				if got != "" {
					t.Fatalf("Content-Disposition = %q, want unset", got)
				}
				return
			}
			if got != tt.want {
				t.Fatalf("Content-Disposition = %q, want %q", got, tt.want)
			}
		})
	}
}

// A quote or backslash in the name must not break out of the quoted string.
func TestServeContent_DispositionSanitizesFilename(t *testing.T) {
	w := serve(t, httptest.NewRequest(http.MethodGet, "/clip", nil), ContentConfig{
		Filename: "ev\"il\\.mp4",
	})

	got := w.Header().Get("Content-Disposition")
	if !strings.HasPrefix(got, `attachment; filename="ev_il_.mp4"`) {
		t.Fatalf("Content-Disposition = %q, want sanitized filename", got)
	}
}

// A directory component in the name must not leak the server's paths.
func TestServeContent_DispositionStripsPath(t *testing.T) {
	w := serve(t, httptest.NewRequest(http.MethodGet, "/clip", nil), ContentConfig{
		Filename: "/srv/secret/media/a.mp4",
	})

	got := w.Header().Get("Content-Disposition")
	if strings.Contains(got, "secret") {
		t.Fatalf("Content-Disposition = %q, must not contain the directory", got)
	}
}

func TestServeContent_CacheControl(t *testing.T) {
	w := serve(t, httptest.NewRequest(http.MethodGet, "/clip", nil), ContentConfig{
		CacheControl: "public, max-age=3600",
	})

	if got := w.Header().Get("Cache-Control"); got != "public, max-age=3600" {
		t.Fatalf("Cache-Control = %q", got)
	}
}

func TestQuoteETag(t *testing.T) {
	tests := []struct{ in, want string }{
		{"v1", `"v1"`},
		{`"v1"`, `"v1"`},
		{`W/"v1"`, `W/"v1"`},
		{`ev"il`, `"evil"`},
	}

	for _, tt := range tests {
		if got := quoteETag(tt.in); got != tt.want {
			t.Errorf("quoteETag(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func writeTempFile(t *testing.T, name string, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	return path
}

func assertStatus(t *testing.T, err error, want int) {
	t.Helper()
	var apiErr *errors.Error
	if !stderrors.As(err, &apiErr) {
		t.Fatalf("err = %v, want *errors.Error", err)
	}
	if apiErr.StatusCode != want {
		t.Fatalf("status = %d, want %d", apiErr.StatusCode, want)
	}
}

func TestServeFile_Range(t *testing.T) {
	path := writeTempFile(t, "clip.mp4", content)

	req := httptest.NewRequest(http.MethodGet, "/clip", nil)
	req.Header.Set("Range", "bytes=2-4")
	w := httptest.NewRecorder()

	if err := ServeFile(w, req, path, ContentConfig{}); err != nil {
		t.Fatalf("ServeFile: %v", err)
	}
	if w.Code != http.StatusPartialContent {
		t.Fatalf("status = %d, want 206", w.Code)
	}
	if got := w.Body.String(); got != "234" {
		t.Fatalf("body = %q, want 234", got)
	}
}

// The extension must drive detection even though no Content-Disposition was
// requested, so serving media does not fall back to octet-stream.
func TestServeFile_DetectsTypeWithoutDisposition(t *testing.T) {
	path := writeTempFile(t, "clip.mp4", content)

	w := httptest.NewRecorder()
	if err := ServeFile(w, httptest.NewRequest(http.MethodGet, "/clip", nil), path, ContentConfig{}); err != nil {
		t.Fatalf("ServeFile: %v", err)
	}
	if got := w.Header().Get("Content-Type"); !strings.HasPrefix(got, "video/mp4") {
		t.Fatalf("Content-Type = %q, want video/mp4", got)
	}
	if got := w.Header().Get("Content-Disposition"); got != "" {
		t.Fatalf("Content-Disposition = %q, want unset", got)
	}
}

func TestServeFile_DefaultsModTime(t *testing.T) {
	path := writeTempFile(t, "clip.mp4", content)

	w := httptest.NewRecorder()
	if err := ServeFile(w, httptest.NewRequest(http.MethodGet, "/clip", nil), path, ContentConfig{}); err != nil {
		t.Fatalf("ServeFile: %v", err)
	}
	if w.Header().Get("Last-Modified") == "" {
		t.Fatal("Last-Modified is unset, want the file's mtime")
	}
}

func TestServeFile_Missing(t *testing.T) {
	w := httptest.NewRecorder()
	err := ServeFile(w, httptest.NewRequest(http.MethodGet, "/clip", nil),
		filepath.Join(t.TempDir(), "nope.mp4"), ContentConfig{})

	assertStatus(t, err, http.StatusNotFound)
	if w.Body.Len() != 0 {
		t.Fatalf("body = %q, want nothing written on error", w.Body.String())
	}
}

func TestServeFile_Directory(t *testing.T) {
	w := httptest.NewRecorder()
	err := ServeFile(w, httptest.NewRequest(http.MethodGet, "/clip", nil), t.TempDir(), ContentConfig{})

	assertStatus(t, err, http.StatusNotFound)
	if w.Body.Len() != 0 {
		t.Fatalf("body = %q, want nothing written on error", w.Body.String())
	}
}
