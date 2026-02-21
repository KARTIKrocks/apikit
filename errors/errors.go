// Package errors provides structured API error types that integrate with
// Go's standard errors package (errors.Is, errors.As, error wrapping).
//
// Usage:
//
//	err := errors.New(errors.CodeNotFound, "user not found")
//	err = err.WithStatus(404).WithDetail("user_id", "abc123")
//
//	// In handlers:
//	if errors.Is(err, errors.ErrNotFound) { ... }
//
//	// Automatic HTTP response:
//	var apiErr *errors.Error
//	if errors.As(err, &apiErr) {
//	    apiErr.StatusCode // 404
//	}
package errors

import (
	stderrors "errors"
	"fmt"
	"net/http"
	"runtime"
)

// Error represents a structured API error.
// It implements the error interface and supports Go's error wrapping.
type Error struct {
	// StatusCode is the HTTP status code to respond with.
	StatusCode int `json:"status_code"`

	// Code is a machine-readable error code (e.g., "NOT_FOUND", "VALIDATION_ERROR").
	Code string `json:"code"`

	// Message is a human-readable error description.
	Message string `json:"message"`

	// Fields contains field-level validation errors.
	// Key is the field name, value is the error description.
	Fields map[string]string `json:"fields,omitempty"`

	// Details contains arbitrary additional error metadata.
	Details map[string]any `json:"details,omitempty"`

	// Err is the underlying wrapped error.
	Err error `json:"-"`

	// Stack holds the caller information for debugging.
	Stack string `json:"-"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error for errors.Is/As support.
func (e *Error) Unwrap() error {
	return e.Err
}

// Is reports whether target matches this error's Code.
// This allows errors.Is(err, ErrNotFound) to work even when messages differ.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// clone creates a shallow copy of the error so that With* methods
// are safe to use on sentinel errors without mutating the original.
func (e *Error) clone() *Error {
	cp := *e
	return &cp
}

// WithStatus sets the HTTP status code on a copy of the error.
func (e *Error) WithStatus(status int) *Error {
	cp := e.clone()
	cp.StatusCode = status
	return cp
}

// WithMessage replaces the error message on a copy of the error.
func (e *Error) WithMessage(msg string) *Error {
	cp := e.clone()
	cp.Message = msg
	return cp
}

// WithField adds a single field error on a copy of the error.
func (e *Error) WithField(field, message string) *Error {
	cp := e.clone()
	cp.Fields = make(map[string]string, len(e.Fields)+1)
	for k, v := range e.Fields {
		cp.Fields[k] = v
	}
	cp.Fields[field] = message
	return cp
}

// WithFields sets multiple field errors on a copy of the error.
func (e *Error) WithFields(fields map[string]string) *Error {
	cp := e.clone()
	cp.Fields = fields
	return cp
}

// WithDetail adds a single detail entry on a copy of the error.
func (e *Error) WithDetail(key string, value any) *Error {
	cp := e.clone()
	cp.Details = make(map[string]any, len(e.Details)+1)
	for k, v := range e.Details {
		cp.Details[k] = v
	}
	cp.Details[key] = value
	return cp
}

// WithDetails sets the details map on a copy of the error.
func (e *Error) WithDetails(details map[string]any) *Error {
	cp := e.clone()
	cp.Details = details
	return cp
}

// Wrap wraps an underlying error on a copy of the error.
func (e *Error) Wrap(err error) *Error {
	cp := e.clone()
	cp.Err = err
	return cp
}

// New creates a new Error with the given code and message.
// It captures the caller's location for debugging.
func New(code string, message string) *Error {
	return &Error{
		StatusCode: codeToStatus(code),
		Code:       code,
		Message:    message,
		Stack:      caller(2),
	}
}

// Newf creates a new Error with a formatted message.
func Newf(code string, format string, args ...any) *Error {
	return &Error{
		StatusCode: codeToStatus(code),
		Code:       code,
		Message:    fmt.Sprintf(format, args...),
		Stack:      caller(2),
	}
}

// From wraps a standard error into an API Error.
// If the error is already an *Error, it is returned as-is.
// Otherwise, it wraps the error as an internal server error.
func From(err error) *Error {
	if err == nil {
		return nil
	}

	var apiErr *Error
	if stderrors.As(err, &apiErr) {
		return apiErr
	}

	return &Error{
		StatusCode: http.StatusInternalServerError,
		Code:       CodeInternal,
		Message:    "An internal error occurred",
		Err:        err,
		Stack:      caller(2),
	}
}

// Fromf wraps a standard error with a custom message.
func Fromf(err error, code string, format string, args ...any) *Error {
	return &Error{
		StatusCode: codeToStatus(code),
		Code:       code,
		Message:    fmt.Sprintf(format, args...),
		Err:        err,
		Stack:      caller(2),
	}
}

// HTTPStatus returns the status code, or 500 if the error is not an *Error.
func HTTPStatus(err error) int {
	var apiErr *Error
	if stderrors.As(err, &apiErr) {
		return apiErr.StatusCode
	}
	return http.StatusInternalServerError
}

// ErrorCode extracts the error code, or CodeInternal if the error is not an *Error.
func ErrorCode(err error) string {
	var apiErr *Error
	if stderrors.As(err, &apiErr) {
		return apiErr.Code
	}
	return CodeInternal
}

// caller returns a string identifying the caller at the given skip depth.
func caller(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}
	return fmt.Sprintf("%s:%d", file, line)
}
