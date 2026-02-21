package response

import (
	"encoding/json"
	stderrors "errors"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"

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
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition",
		fmt.Sprintf(`attachment; filename="%s"`, filepath.Base(filename)))
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
