package response

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/KARTIKrocks/apikit/errors"
)

// ContentConfig configures ServeContent and ServeFile.
//
// Every field is optional; the zero value serves the content with a sniffed
// content type, no caching headers, and no Content-Disposition.
type ContentConfig struct {
	// Filename is used for content-type detection (by extension) and, when
	// non-empty, sets a Content-Disposition header naming the download.
	Filename string

	// Inline serves the content with an "inline" Content-Disposition rather
	// than "attachment", so browsers render it in place instead of
	// downloading it. Only applies when Filename is set.
	Inline bool

	// ContentType overrides detection. When empty, the type is derived from
	// Filename's extension, falling back to sniffing the first 512 bytes.
	ContentType string

	// ModTime is the content's last modification time. When non-zero it is
	// sent as Last-Modified and used to answer If-Modified-Since and
	// If-Range requests.
	ModTime time.Time

	// ETag identifies this version of the content and is used to answer
	// If-None-Match and If-Range requests. Quotes are added if missing, so
	// both `abc123` and `"abc123"` are accepted; a weak validator must be
	// written in full as `W/"abc123"`.
	ETag string

	// CacheControl sets the Cache-Control header, e.g. "public, max-age=3600".
	// When empty the header is not set.
	CacheControl string
}

// ServeContent serves an io.ReadSeeker with full HTTP Range support, so it can
// back media that the client needs to seek within (a <video> or <audio>
// element) as well as resumable downloads.
//
//	f, err := os.Open(path)
//	if err != nil {
//	    return errors.NotFound("File")
//	}
//	defer f.Close()
//
//	response.ServeContent(w, r, f, response.ContentConfig{
//	    Filename: "clip.mp4",
//	    Inline:   true,
//	    ModTime:  info.ModTime(),
//	})
//
// Range requests, conditional GETs (If-None-Match, If-Modified-Since,
// If-Range), HEAD, and 416 Range Not Satisfiable are all handled. Unlike
// Reader, which streams a response start to finish, this seeks within the
// content to serve the requested byte range.
//
// The caller is responsible for closing the content if it needs closing.
func ServeContent(w http.ResponseWriter, r *http.Request, content io.ReadSeeker, cfg ContentConfig) {
	serveContent(w, r, cfg.Filename, content, cfg)
}

// ServeFile opens the file at path and serves it with ServeContent, defaulting
// ModTime to the file's modification time and detecting the content type from
// its extension.
//
//	response.Handle(func(w http.ResponseWriter, r *http.Request) error {
//	    return response.ServeFile(w, r, "/srv/media/"+id+".mp4", response.ContentConfig{
//	        Inline:       true,
//	        CacheControl: "public, max-age=86400",
//	    })
//	})
//
// A missing file becomes a 404 and a directory a 404; any other failure is a
// 500 wrapping the underlying error. Nothing is written to w when an error is
// returned, so the caller (or response.Handle) can render it.
//
// path is trusted. Never build it directly from user input — a request-supplied
// segment must be validated, or joined with filepath.Join and confirmed to
// still live under the intended root, or an attacker can traverse out of it.
func ServeFile(w http.ResponseWriter, r *http.Request, path string, cfg ContentConfig) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.NotFound("File")
		}
		return errors.Internalf(err, "Failed to open file")
	}
	defer func() { _ = f.Close() }()

	info, err := f.Stat()
	if err != nil {
		return errors.Internalf(err, "Failed to read file")
	}
	// Directories have no meaningful content and http.ServeContent would
	// serve their raw bytes; report them as absent rather than listing them.
	if info.IsDir() {
		return errors.NotFound("File")
	}

	if cfg.ModTime.IsZero() {
		cfg.ModTime = info.ModTime()
	}
	// The on-disk name drives content-type detection even when the caller
	// does not want a Content-Disposition header.
	serveContent(w, r, filepath.Base(path), f, cfg)
	return nil
}

// serveContent applies the config headers and delegates the transfer to
// http.ServeContent, which owns Range and conditional-request handling.
// name is used for extension-based content-type detection only.
func serveContent(w http.ResponseWriter, r *http.Request, name string, content io.ReadSeeker, cfg ContentConfig) {
	if cfg.ContentType != "" {
		// http.ServeContent skips detection when the header is already set.
		w.Header().Set("Content-Type", cfg.ContentType)
		name = ""
	}
	if cfg.ETag != "" {
		w.Header().Set("Etag", quoteETag(cfg.ETag))
	}
	if cfg.CacheControl != "" {
		w.Header().Set("Cache-Control", cfg.CacheControl)
	}
	if cfg.Filename != "" {
		w.Header().Set("Content-Disposition", contentDisposition(cfg.Filename, cfg.Inline))
	}

	http.ServeContent(w, r, name, cfg.ModTime, content)
}

// quoteETag returns tag in the quoted form RFC 9110 requires. An unquoted tag
// would never match an If-None-Match or If-Range header, silently disabling
// the conditional handling it was supplied to enable.
func quoteETag(tag string) string {
	if strings.HasPrefix(tag, `W/"`) && strings.HasSuffix(tag, `"`) {
		return tag
	}
	if strings.HasPrefix(tag, `"`) && strings.HasSuffix(tag, `"`) && len(tag) >= 2 {
		return tag
	}
	return `"` + strings.ReplaceAll(tag, `"`, "") + `"`
}

// contentDisposition builds a Content-Disposition header value for filename,
// pairing an ASCII-safe form for broad compatibility with the RFC 5987 UTF-8
// variant.
func contentDisposition(filename string, inline bool) string {
	disposition := "attachment"
	if inline {
		disposition = "inline"
	}

	name := filepath.Base(filename)
	// Strip characters that would break out of the quoted-string form.
	safeName := strings.Map(func(rn rune) rune {
		if rn == '"' || rn == '\\' || rn < 0x20 {
			return '_'
		}
		return rn
	}, name)

	return fmt.Sprintf(`%s; filename="%s"; filename*=UTF-8''%s`,
		disposition, safeName, url.PathEscape(name))
}
