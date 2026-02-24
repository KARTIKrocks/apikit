package response

import (
	"encoding/json"
	"encoding/xml"
	stderrors "errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/KARTIKrocks/apikit/errors"
)

// Stream provides Server-Sent Events (SSE) streaming.
//
//	response.Stream(w, func(send func(event, data string) error) error {
//	    for msg := range messages {
//	        if err := send("message", msg); err != nil {
//	            return err
//	        }
//	    }
//	    return nil
//	})
func Stream(w http.ResponseWriter, fn func(send func(event, data string) error) error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		InternalServerError(w, "Streaming not supported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	send := func(event, data string) error {
		if event != "" {
			if _, err := fmt.Fprintf(w, "event: %s\n", event); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
			return err
		}
		flusher.Flush()
		return nil
	}

	if err := fn(send); err != nil {
		// Can't send a proper error response — headers already sent.
		// Log the full error for operators but only send a safe message to the client.
		slog.Error("SSE stream error", "error", err)

		// Only expose the message if it's an API error (user-facing); hide internal details.
		msg := "An internal error occurred"
		var apiErr *errors.Error
		if stderrors.As(err, &apiErr) {
			msg = apiErr.Message
		}
		_ = send("error", msg)
	}
}

// StreamJSON is like Stream but sends JSON-encoded data.
func StreamJSON(w http.ResponseWriter, fn func(send func(event string, data any) error) error) {
	Stream(w, func(sendRaw func(event, data string) error) error {
		send := func(event string, data any) error {
			b, err := json.Marshal(data)
			if err != nil {
				return err
			}
			return sendRaw(event, string(b))
		}
		return fn(send)
	})
}

// File sends a file as a download.
func File(w http.ResponseWriter, filename string, data []byte, contentType string) {
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	name := filepath.Base(filename)

	// Sanitize filename: remove characters unsafe in Content-Disposition.
	safeName := strings.Map(func(r rune) rune {
		if r == '"' || r == '\\' || r < 0x20 {
			return '_'
		}
		return r
	}, name)

	w.Header().Set("Content-Type", contentType)
	// ASCII-safe filename for broad compatibility, plus RFC 5987 UTF-8 variant.
	w.Header().Set("Content-Disposition",
		fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, safeName, url.PathEscape(name)))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// Redirect sends an HTTP redirect.
func Redirect(w http.ResponseWriter, r *http.Request, url string, code int) {
	http.Redirect(w, r, url, code)
}

// Raw writes raw bytes with a custom content type.
// Use this for non-JSON responses like HTML, XML, plain text, etc.
func Raw(w http.ResponseWriter, statusCode int, contentType string, data []byte) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(statusCode)
	_, _ = w.Write(data)
}

// HTML writes an HTML response.
func HTML(w http.ResponseWriter, statusCode int, html string) {
	Raw(w, statusCode, "text/html; charset=utf-8", []byte(html))
}

// Text writes a plain text response.
func Text(w http.ResponseWriter, statusCode int, text string) {
	Raw(w, statusCode, "text/plain; charset=utf-8", []byte(text))
}

// XML writes an XML response.
func XML(w http.ResponseWriter, statusCode int, data any) {
	b, err := xml.Marshal(data)
	if err != nil {
		slog.Error("failed to encode XML response", "error", err)
		InternalServerError(w, "Failed to encode XML")
		return
	}
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(statusCode)
	_, _ = w.Write([]byte(xml.Header))
	_, _ = w.Write(b)
}

// IndentedJSON writes a pretty-printed JSON response (no envelope).
// Useful for debug endpoints or human-readable API responses.
func IndentedJSON(w http.ResponseWriter, statusCode int, data any) {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		slog.Error("failed to encode JSON response", "error", err)
		InternalServerError(w, "Failed to encode JSON")
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(statusCode)
	_, _ = w.Write(b)
}

// PureJSON writes a JSON response without HTML escaping of <, >, and &.
// Standard json.Encoder escapes these characters for safe embedding in HTML;
// PureJSON preserves them as-is for clients that consume raw JSON.
func PureJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(statusCode)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(data); err != nil {
		slog.Error("failed to encode JSON response", "error", err)
	}
}

// JSONP writes a JSONP response for cross-domain callbacks.
// The callback name is read from the "callback" query parameter.
// If no callback is provided, it falls back to a regular JSON response.
func JSONP(w http.ResponseWriter, r *http.Request, statusCode int, data any) {
	callback := r.URL.Query().Get("callback")
	if callback == "" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(data)
		return
	}

	b, err := json.Marshal(data)
	if err != nil {
		slog.Error("failed to encode JSONP response", "error", err)
		InternalServerError(w, "Failed to encode JSON")
		return
	}

	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(statusCode)
	_, _ = fmt.Fprintf(w, "%s(%s);", callback, b)
}

// Reader streams data from an io.Reader to the response.
// Useful for proxying responses or sending large files without loading them into memory.
func Reader(w http.ResponseWriter, statusCode int, contentType string, contentLength int64, reader io.Reader) {
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", contentType)
	if contentLength >= 0 {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", contentLength))
	}
	w.WriteHeader(statusCode)
	_, _ = io.Copy(w, reader)
}
