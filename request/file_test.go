package request

import (
	"bytes"
	"encoding/json"
	stderrors "errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/KARTIKrocks/apikit/errors"
)

// pngHeader is a minimal PNG signature; http.DetectContentType keys off it.
var pngHeader = []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}

// pngFile returns a fresh buffer holding a PNG signature padded to n bytes.
func pngFile(n int) []byte {
	b := make([]byte, 0, n)
	b = append(b, pngHeader...)
	return append(b, bytes.Repeat([]byte{0xAB}, n-len(b))...)
}

// newFileRequest builds a multipart request carrying a single file field.
func newFileRequest(field, filename string, content []byte) *http.Request {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, _ := w.CreateFormFile(field, filename)
	_, _ = fw.Write(content)
	_ = w.Close()

	r := httptest.NewRequest("POST", "/", &buf)
	r.Header.Set("Content-Type", w.FormDataContentType())
	return r
}

// apiError extracts the structured error, failing the test if err is not one.
func apiError(t *testing.T, err error) *errors.Error {
	t.Helper()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	var apiErr *errors.Error
	if !stderrors.As(err, &apiErr) {
		t.Fatalf("error is not *errors.Error: %v", err)
	}
	return apiErr
}

func TestFormFileWithConfig_HappyPath(t *testing.T) {
	content := pngFile(40)
	r := newFileRequest("avatar", "avatar.png", content)

	f, fh, err := FormFileWithConfig(r, "avatar", FileConfig{
		MaxBytes:     1 << 20,
		AllowedTypes: []string{"image/png"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()

	if fh.Filename != "avatar.png" {
		t.Errorf("Filename = %q, want %q", fh.Filename, "avatar.png")
	}
	if fh.Size != int64(len(content)) {
		t.Errorf("Size = %d, want %d", fh.Size, len(content))
	}

	// The sniff must have rewound the file: the caller reads from byte zero.
	got, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("read %d bytes, want the original %d", len(got), len(content))
	}
}

func TestFormFileWithConfig_RewindsShortFile(t *testing.T) {
	// Shorter than sniffLen, so the sniff hits ErrUnexpectedEOF.
	content := []byte("hello")
	r := newFileRequest("doc", "doc.txt", content)

	f, _, err := FormFileWithConfig(r, "doc", FileConfig{AllowedTypes: []string{"text/plain"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()

	got, _ := io.ReadAll(f)
	if !bytes.Equal(got, content) {
		t.Errorf("read %q, want %q", got, content)
	}
}

// Files larger than MaxMemory are spooled to disk and backed by an *os.File,
// a different Seek implementation than the in-memory path above.
func TestFormFileWithConfig_RewindsDiskBackedFile(t *testing.T) {
	content := pngFile(8200)
	r := newFileRequest("avatar", "avatar.png", content)

	f, fh, err := FormFileWithConfig(r, "avatar", FileConfig{
		MaxBytes:     1 << 20,
		MaxMemory:    64, // below the file size, forcing the spool-to-disk path
		AllowedTypes: []string{"image/png"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()

	if _, ok := f.(*os.File); !ok {
		t.Fatalf("got %T, want *os.File — the disk path is not being exercised", f)
	}
	got, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("read %d bytes, want the original %d", len(got), len(content))
	}
	if fh.Size != int64(len(content)) {
		t.Errorf("Size = %d, want %d", fh.Size, len(content))
	}
}

func TestFormFileWithConfig_EmptyFile(t *testing.T) {
	r := newFileRequest("avatar", "avatar.png", nil)

	// An empty file sniffs as text/plain, so an image allowlist rejects it.
	_, _, err := FormFileWithConfig(r, "avatar", FileConfig{AllowedTypes: []string{"image/*"}})
	apiErr := apiError(t, err)
	if apiErr.StatusCode != http.StatusUnsupportedMediaType {
		t.Errorf("status = %d, want %d", apiErr.StatusCode, http.StatusUnsupportedMediaType)
	}

	// With no allowlist it is accepted; enforcing non-empty is the caller's call.
	r2 := newFileRequest("avatar", "avatar.png", nil)
	f, fh, err := FormFileWithConfig(r2, "avatar", FileConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()
	if fh.Size != 0 {
		t.Errorf("Size = %d, want 0", fh.Size)
	}
}

func TestFormFileWithConfig_ExactlyAtLimit(t *testing.T) {
	// The limit applies to the whole multipart body, not the file alone, so a
	// file at exactly MaxBytes must still be rejected once framing is added.
	content := bytes.Repeat([]byte{'a'}, 512)
	r := newFileRequest("doc", "doc.txt", content)

	_, _, err := FormFileWithConfig(r, "doc", FileConfig{MaxBytes: 512})
	apiErr := apiError(t, err)
	if apiErr.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want %d", apiErr.StatusCode, http.StatusRequestEntityTooLarge)
	}
}

func TestFormFileWithConfig_TooLarge(t *testing.T) {
	r := newFileRequest("avatar", "avatar.png", bytes.Repeat([]byte{'a'}, 4096))

	_, _, err := FormFileWithConfig(r, "avatar", FileConfig{MaxBytes: 512})
	apiErr := apiError(t, err)
	if apiErr.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want %d", apiErr.StatusCode, http.StatusRequestEntityTooLarge)
	}
	if apiErr.Code != errors.CodeRequestTooLarge {
		t.Errorf("code = %q, want %q", apiErr.Code, errors.CodeRequestTooLarge)
	}
}

func TestFormFileWithConfig_TooLargeWhenPreParsed(t *testing.T) {
	r := newFileRequest("avatar", "avatar.png", bytes.Repeat([]byte{'a'}, 4096))
	// Simulate middleware that already consumed the body; the MaxBytesReader
	// path is skipped and only the fh.Size backstop can catch this.
	if err := r.ParseMultipartForm(DefaultMaxMultipartMemory); err != nil {
		t.Fatalf("pre-parse failed: %v", err)
	}

	_, _, err := FormFileWithConfig(r, "avatar", FileConfig{MaxBytes: 512})
	apiErr := apiError(t, err)
	if apiErr.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want %d", apiErr.StatusCode, http.StatusRequestEntityTooLarge)
	}
}

func TestFormFileWithConfig_DisallowedType(t *testing.T) {
	r := newFileRequest("avatar", "avatar.png", []byte("plain text, not an image"))

	_, _, err := FormFileWithConfig(r, "avatar", FileConfig{
		AllowedTypes: []string{"image/png", "image/jpeg"},
	})
	apiErr := apiError(t, err)
	if apiErr.StatusCode != http.StatusUnsupportedMediaType {
		t.Errorf("status = %d, want %d", apiErr.StatusCode, http.StatusUnsupportedMediaType)
	}
}

// A forged part header must not get a file past the allowlist.
func TestFormFileWithConfig_IgnoresClaimedContentType(t *testing.T) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	h := make(map[string][]string)
	h["Content-Disposition"] = []string{`form-data; name="avatar"; filename="evil.png"`}
	h["Content-Type"] = []string{"image/png"}
	fw, _ := w.CreatePart(h)
	_, _ = fw.Write([]byte("#!/bin/sh\nrm -rf /\n"))
	_ = w.Close()

	r := httptest.NewRequest("POST", "/", &buf)
	r.Header.Set("Content-Type", w.FormDataContentType())

	_, _, err := FormFileWithConfig(r, "avatar", FileConfig{AllowedTypes: []string{"image/png"}})
	apiErr := apiError(t, err)
	if apiErr.StatusCode != http.StatusUnsupportedMediaType {
		t.Errorf("status = %d, want %d", apiErr.StatusCode, http.StatusUnsupportedMediaType)
	}
}

func TestFormFileWithConfig_WildcardType(t *testing.T) {
	content := pngFile(40)
	r := newFileRequest("avatar", "avatar.png", content)

	f, _, err := FormFileWithConfig(r, "avatar", FileConfig{AllowedTypes: []string{"image/*"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	f.Close()
}

func TestFormFileWithConfig_EmptyAllowlistAllowsAny(t *testing.T) {
	r := newFileRequest("doc", "doc.bin", []byte{0x00, 0x01, 0x02})

	f, _, err := FormFileWithConfig(r, "doc", FileConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	f.Close()
}

func TestFormFileWithConfig_MissingField(t *testing.T) {
	r := newFileRequest("avatar", "avatar.png", pngHeader)

	_, _, err := FormFileWithConfig(r, "banner", FileConfig{})
	apiErr := apiError(t, err)
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", apiErr.StatusCode, http.StatusBadRequest)
	}
}

func TestFormFileWithConfig_NotMultipart(t *testing.T) {
	r := httptest.NewRequest("POST", "/", bytes.NewReader([]byte("a=1")))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	_, _, err := FormFileWithConfig(r, "avatar", FileConfig{})
	apiErr := apiError(t, err)
	if apiErr.StatusCode != http.StatusUnsupportedMediaType {
		t.Errorf("status = %d, want %d", apiErr.StatusCode, http.StatusUnsupportedMediaType)
	}
}

func TestFormFileWithConfig_NoBody(t *testing.T) {
	r := httptest.NewRequest("POST", "/", nil)

	_, _, err := FormFileWithConfig(r, "avatar", FileConfig{})
	apiError(t, err)
}

// FormFile previously flattened every failure into 400, including an
// oversized body that the rest of the package reports as 413.
func TestFormFile_TooLargeIsNot400(t *testing.T) {
	r := newFileRequest("avatar", "avatar.png", bytes.Repeat([]byte{'a'}, 4096))
	r.Body = http.MaxBytesReader(nil, r.Body, 512)

	_, err := FormFile(r, "avatar")
	apiErr := apiError(t, err)
	if apiErr.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want %d", apiErr.StatusCode, http.StatusRequestEntityTooLarge)
	}
}

// An unclassified failure must keep the underlying cause reachable via
// errors.Is/As for logs, without exposing it in the client-facing message.
func TestFileError_PreservesCause(t *testing.T) {
	cause := stderrors.New("some multipart failure")
	err := fileError("avatar", cause)

	if !stderrors.Is(err, cause) {
		t.Errorf("errors.Is could not reach the cause through %v", err)
	}
	apiErr := apiError(t, err)
	if strings.Contains(apiErr.Message, cause.Error()) {
		t.Errorf("Message %q leaks the internal cause to the client", apiErr.Message)
	}

	// Error.Err is json:"-", so the cause must not reach the response body.
	b, marshalErr := json.Marshal(apiErr)
	if marshalErr != nil {
		t.Fatalf("marshal: %v", marshalErr)
	}
	if bytes.Contains(b, []byte("some multipart failure")) {
		t.Errorf("serialized error leaks the cause: %s", b)
	}
}

func TestMediaTypeAllowed(t *testing.T) {
	tests := []struct {
		name     string
		detected string
		allowed  []string
		want     bool
	}{
		{"exact match", "image/png", []string{"image/png"}, true},
		{"case insensitive", "image/png", []string{"IMAGE/PNG"}, true},
		{"second entry", "image/jpeg", []string{"image/png", "image/jpeg"}, true},
		{"no match", "text/plain", []string{"image/png"}, false},
		{"wildcard match", "image/gif", []string{"image/*"}, true},
		{"wildcard no match", "video/mp4", []string{"image/*"}, false},
		{"wildcard is not a prefix match", "imagexml/foo", []string{"image/*"}, false},
		{"empty allowlist", "image/png", nil, false},
		{"match anything", "application/octet-stream", []string{"*/*"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mediaTypeAllowed(tt.detected, tt.allowed); got != tt.want {
				t.Errorf("mediaTypeAllowed(%q, %v) = %v, want %v",
					tt.detected, tt.allowed, got, tt.want)
			}
		})
	}
}
