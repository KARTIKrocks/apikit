// Package request provides helpers for parsing, binding, and validating
// HTTP request data.
//
// The core function is Bind[T], which decodes the request body into a
// typed struct, enforces size limits, and validates content types:
//
//	type CreateUserReq struct {
//	    Name  string `json:"name"`
//	    Email string `json:"email"`
//	}
//
//	func handler(w http.ResponseWriter, r *http.Request) error {
//	    req, err := request.Bind[CreateUserReq](r)
//	    if err != nil { return err }
//	    // req is CreateUserReq, fully decoded and typed
//	}
package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/KARTIKrocks/apikit/errors"
)

// Default limits
const (
	DefaultMaxBodySize = 1 << 20 // 1 MB
)

// Config holds request binding configuration.
type Config struct {
	// MaxBodySize is the maximum allowed request body size in bytes.
	// Defaults to 1MB if zero.
	MaxBodySize int64

	// DisallowUnknownFields rejects JSON bodies with unknown fields.
	DisallowUnknownFields bool
}

// DefaultConfig returns the default binding configuration.
func DefaultConfig() Config {
	return Config{
		MaxBodySize:           DefaultMaxBodySize,
		DisallowUnknownFields: false,
	}
}

// global config, can be overridden via SetConfig.
var globalConfig = DefaultConfig()

// SetConfig sets the global binding configuration.
func SetConfig(cfg Config) {
	if cfg.MaxBodySize <= 0 {
		cfg.MaxBodySize = DefaultMaxBodySize
	}
	globalConfig = cfg
}

// Bind decodes the JSON request body into T using the global config.
// Returns the decoded value and an *errors.Error on failure.
//
//	user, err := request.Bind[CreateUserReq](r)
func Bind[T any](r *http.Request) (T, error) {
	return BindWithConfig[T](r, globalConfig)
}

// BindWithConfig decodes the JSON request body into T using the provided config.
func BindWithConfig[T any](r *http.Request, cfg Config) (T, error) {
	var v T

	// Check content type
	if err := validateContentType(r, "application/json"); err != nil {
		return v, err
	}

	// Check body exists
	if r.Body == nil || r.Body == http.NoBody {
		return v, errors.BadRequest("Request body is required")
	}

	// Enforce body size limit
	maxSize := cfg.MaxBodySize
	if maxSize <= 0 {
		maxSize = DefaultMaxBodySize
	}

	// Read up to maxSize+1 bytes. If we get maxSize+1, the body is too large.
	body, err := io.ReadAll(io.LimitReader(r.Body, maxSize+1))
	if err != nil {
		return v, errors.BadRequest("Failed to read request body")
	}
	if int64(len(body)) > maxSize {
		return v, errors.New(errors.CodeRequestTooLarge, "Request body too large").
			WithStatus(http.StatusRequestEntityTooLarge)
	}

	// Drain and close the body to enable connection reuse
	_, _ = io.Copy(io.Discard, r.Body)
	_ = r.Body.Close()

	// Decode JSON
	decoder := json.NewDecoder(bytes.NewReader(body))
	if cfg.DisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}

	if err := decoder.Decode(&v); err != nil {
		return v, handleDecodeError(err)
	}

	// Run struct tag validation (simple rules)
	if err := ValidateStruct(&v); err != nil {
		return v, err
	}

	// Auto-validate if T implements Validator (cross-field logic)
	if validator, ok := any(&v).(Validator); ok {
		if err := validator.Validate(); err != nil {
			return v, err
		}
	}

	return v, nil
}

// BindJSON is an alias for Bind that makes the JSON intent explicit.
func BindJSON[T any](r *http.Request) (T, error) {
	return Bind[T](r)
}

// DecodeJSON decodes the request body into the provided pointer.
// Use Bind[T] for a generic alternative.
func DecodeJSON(r *http.Request, v any) error {
	if err := validateContentType(r, "application/json"); err != nil {
		return err
	}

	if r.Body == nil || r.Body == http.NoBody {
		return errors.BadRequest("Request body is required")
	}

	maxSize := globalConfig.MaxBodySize
	if maxSize <= 0 {
		maxSize = DefaultMaxBodySize
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxSize+1))
	if err != nil {
		return errors.BadRequest("Failed to read request body")
	}
	if int64(len(body)) > maxSize {
		return errors.New(errors.CodeRequestTooLarge, "Request body too large").
			WithStatus(http.StatusRequestEntityTooLarge)
	}

	// Drain and close the body to enable connection reuse
	_, _ = io.Copy(io.Discard, r.Body)
	_ = r.Body.Close()

	decoder := json.NewDecoder(bytes.NewReader(body))
	if globalConfig.DisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}

	if err := decoder.Decode(v); err != nil {
		return handleDecodeError(err)
	}

	// Run struct tag validation
	if err := ValidateStruct(v); err != nil {
		return err
	}

	return nil
}

// validateContentType checks that the request has the expected content type.
func validateContentType(r *http.Request, expected string) error {
	ct := r.Header.Get("Content-Type")
	if ct == "" {
		// Be lenient: if no content type is set, allow it
		return nil
	}

	mediaType, _, err := mime.ParseMediaType(ct)
	if err != nil {
		return errors.BadRequest("Malformed Content-Type header")
	}

	if !strings.EqualFold(mediaType, expected) {
		return errors.New(errors.CodeUnsupportedMedia,
			fmt.Sprintf("Content-Type must be %s, got %s", expected, mediaType)).
			WithStatus(http.StatusUnsupportedMediaType)
	}

	return nil
}

// handleDecodeError converts JSON decode errors to structured API errors.
func handleDecodeError(err error) *errors.Error {
	switch err {
	case io.EOF:
		return errors.BadRequest("Request body is empty")

	case io.ErrUnexpectedEOF:
		return errors.BadRequest("Malformed JSON: unexpected end of input")

	default:
		switch e := err.(type) {
		case *json.SyntaxError:
			return errors.BadRequest(
				fmt.Sprintf("Malformed JSON at position %d", e.Offset))

		case *json.UnmarshalTypeError:
			return errors.BadRequest(
				fmt.Sprintf("Invalid type for field %q: expected %s", e.Field, e.Type.String())).
				WithField(e.Field, fmt.Sprintf("expected type %s", e.Type.String()))

		case *json.InvalidUnmarshalError:
			// This is a programming error (passing non-pointer)
			return errors.Internal("Internal error processing request")

		default:
			msg := err.Error()
			if strings.Contains(msg, "unknown field") {
				field := strings.TrimPrefix(msg, "json: unknown field ")
				field = strings.Trim(field, "\"")
				return errors.BadRequest(
					fmt.Sprintf("Unknown field %q in request body", field)).
					WithField(field, "unknown field")
			}

			return errors.BadRequest("Invalid request body")
		}
	}
}
