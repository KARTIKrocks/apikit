package request

import (
	stderrors "errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/KARTIKrocks/apikit/errors"
)

// DefaultMaxUploadSize is the default body limit applied by FormFileWithConfig
// when FileConfig.MaxBytes is zero.
const DefaultMaxUploadSize = 10 << 20 // 10 MB

// sniffLen is the number of leading bytes inspected to detect a file's real
// content type. http.DetectContentType never looks at more than this.
const sniffLen = 512

// FileConfig constrains an uploaded file.
type FileConfig struct {
	// MaxBytes is the maximum allowed request body size in bytes. It is
	// enforced before the body is read, so an oversized upload is rejected
	// without being buffered to memory or disk.
	// Defaults to DefaultMaxUploadSize if zero.
	MaxBytes int64

	// MaxMemory is the maximum memory (in bytes) used for parsing the
	// multipart form before spilling to disk.
	// Defaults to DefaultMaxMultipartMemory if zero.
	MaxMemory int64

	// AllowedTypes is an allowlist of media types, matched against the type
	// detected from the file's own bytes rather than the client-supplied
	// Content-Type header. Entries may be exact ("image/png"), a wildcard
	// subtype ("image/*"), or "*/*"; matching is case-insensitive.
	// An empty slice allows any content type.
	//
	// Note that http.DetectContentType recognises a fixed set of signatures
	// and falls back to "application/octet-stream" for formats it does not
	// know, so an allowlist for an exotic type may need that entry (or a
	// caller-side check) to accept legitimate files.
	AllowedTypes []string
}

// FormFileWithConfig returns the first file for the given form field, enforcing
// a size limit and an optional content-type allowlist.
//
// Unlike FormFile it also returns the opened file, positioned at the start:
//
//	f, fh, err := request.FormFileWithConfig(r, "avatar", request.FileConfig{
//	    MaxBytes:     5 << 20,
//	    AllowedTypes: []string{"image/png", "image/jpeg"},
//	})
//	if err != nil {
//	    return err
//	}
//	defer f.Close()
//
// The caller is responsible for closing the returned file.
//
// Content types are detected with http.DetectContentType over the first 512
// bytes of the file. The multipart part's own Content-Type header is not
// consulted, since it is supplied by the client and trivially forged.
func FormFileWithConfig(r *http.Request, field string, cfg FileConfig) (multipart.File, *multipart.FileHeader, error) {
	maxBytes := cfg.MaxBytes
	if maxBytes <= 0 {
		maxBytes = DefaultMaxUploadSize
	}
	maxMemory := cfg.MaxMemory
	if maxMemory <= 0 {
		maxMemory = DefaultMaxMultipartMemory
	}

	// Limit the body before anything reads it. Once the form has been parsed
	// the bytes are already spooled, so a limit applied afterwards would be
	// no protection at all.
	if r.MultipartForm == nil {
		if r.Body == nil || r.Body == http.NoBody {
			return nil, nil, errors.BadRequest("Request body is required")
		}
		r.Body = http.MaxBytesReader(nil, r.Body, maxBytes)

		if err := r.ParseMultipartForm(maxMemory); err != nil {
			return nil, nil, fileError(field, err)
		}
	}

	f, fh, err := r.FormFile(field)
	if err != nil {
		return nil, nil, fileError(field, err)
	}

	// Backstop for a form parsed before this call (by middleware, or by an
	// earlier FormFile), where the body limit above was never applied.
	if fh.Size > maxBytes {
		_ = f.Close()
		return nil, nil, errors.New(errors.CodeRequestTooLarge,
			fmt.Sprintf("File %q exceeds the maximum size of %d bytes", field, maxBytes)).
			WithStatus(http.StatusRequestEntityTooLarge)
	}

	if len(cfg.AllowedTypes) > 0 {
		detected, err := sniffContentType(f)
		if err != nil {
			_ = f.Close()
			return nil, nil, errors.BadRequest(fmt.Sprintf("Failed to read file field %q", field)).
				Wrap(err)
		}
		if !mediaTypeAllowed(detected, cfg.AllowedTypes) {
			_ = f.Close()
			return nil, nil, errors.New(errors.CodeUnsupportedMedia,
				fmt.Sprintf("File type %s is not allowed, expected one of: %s",
					detected, strings.Join(cfg.AllowedTypes, ", "))).
				WithStatus(http.StatusUnsupportedMediaType).
				WithField(field, "unsupported file type")
		}
	}

	return f, fh, nil
}

// sniffContentType detects the media type of f from its leading bytes and
// rewinds it so the caller reads from the start.
func sniffContentType(f multipart.File) (string, error) {
	buf := make([]byte, sniffLen)
	n, err := io.ReadFull(f, buf)
	if err != nil && !stderrors.Is(err, io.EOF) && !stderrors.Is(err, io.ErrUnexpectedEOF) {
		return "", err
	}

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	// DetectContentType appends parameters (e.g. "text/plain; charset=utf-8");
	// strip them so the allowlist can be written in plain media types.
	mediaType, _, err := mime.ParseMediaType(http.DetectContentType(buf[:n]))
	if err != nil {
		return "", err
	}
	return mediaType, nil
}

// mediaTypeAllowed reports whether detected matches any allowlist entry.
// Entries may be exact ("image/png"), a wildcard subtype ("image/*"), or the
// match-anything "*/*".
func mediaTypeAllowed(detected string, allowed []string) bool {
	for _, entry := range allowed {
		if strings.EqualFold(entry, detected) {
			return true
		}
		if entry == "*/*" {
			return true
		}
		if prefix, ok := strings.CutSuffix(entry, "/*"); ok {
			if detectedPrefix, _, found := strings.Cut(detected, "/"); found &&
				strings.EqualFold(prefix, detectedPrefix) {
				return true
			}
		}
	}
	return false
}

// fileError converts a multipart parsing or lookup failure into a structured
// API error, preserving the distinction between a missing field (400) and an
// oversized upload (413).
func fileError(field string, err error) error {
	if isTooLarge(err) {
		return errors.New(errors.CodeRequestTooLarge, "Request body too large").
			WithStatus(http.StatusRequestEntityTooLarge)
	}
	if stderrors.Is(err, http.ErrMissingFile) {
		return errors.BadRequest(fmt.Sprintf("Missing file field %q", field))
	}
	if stderrors.Is(err, http.ErrNotMultipart) {
		return errors.New(errors.CodeUnsupportedMedia,
			"Content-Type must be multipart/form-data").
			WithStatus(http.StatusUnsupportedMediaType)
	}
	// Cause unknown — keep the original for logs and errors.Is/As. The
	// client-facing Message is unchanged; Error.Err is json:"-", so nothing
	// extra reaches the response body.
	return errors.BadRequest(fmt.Sprintf("Missing or invalid file field %q", field)).Wrap(err)
}

// isTooLarge reports whether err came from exceeding a body or memory limit.
func isTooLarge(err error) bool {
	var maxBytesErr *http.MaxBytesError
	if stderrors.As(err, &maxBytesErr) {
		return true
	}
	if stderrors.Is(err, multipart.ErrMessageTooLarge) {
		return true
	}
	// http.MaxBytesReader wraps its error before Go 1.19 and some paths still
	// surface it as a plain string.
	return err != nil && err.Error() == "http: request body too large"
}
