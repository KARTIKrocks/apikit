// Package response provides structured JSON response helpers
// with a consistent envelope format and generics support.
//
// The standard response envelope:
//
//	{
//	    "success": true,
//	    "message": "User created",
//	    "data": { ... },
//	    "meta": { ... },
//	    "timestamp": 1700000000
//	}
//
// Usage with the builder:
//
//	response.New().
//	    Status(201).
//	    Message("User created").
//	    Data(user).
//	    Send(w)
//
// Or with convenience functions:
//
//	response.OK(w, "Success", user)
//	response.NotFound(w, "User not found")
//	response.ValidationError(w, fields)
package response

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/KARTIKrocks/apikit/errors"
)

// Envelope is the standard API response envelope.
// All API responses are wrapped in this structure.
type Envelope struct {
	Success   bool       `json:"success"`
	Message   string     `json:"message,omitempty"`
	Data      any        `json:"data,omitempty"`
	Error     *ErrorBody `json:"error,omitempty"`
	Meta      any        `json:"meta,omitempty"`
	Timestamp int64      `json:"timestamp"`
}

// ErrorBody is the error portion of the response envelope.
type ErrorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
	Details map[string]any    `json:"details,omitempty"`
}

// TypedEnvelope is a generic version of Envelope for type-safe responses.
// Useful when you want compile-time type checking on the data field.
type TypedEnvelope[T any] struct {
	Success   bool       `json:"success"`
	Message   string     `json:"message,omitempty"`
	Data      T          `json:"data,omitempty"`
	Error     *ErrorBody `json:"error,omitempty"`
	Meta      any        `json:"meta,omitempty"`
	Timestamp int64      `json:"timestamp"`
}

// --- Core write function ---

// write is the internal function that writes the response.
// All public functions ultimately call this.
func write(w http.ResponseWriter, statusCode int, response any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Headers and status are already sent — we cannot write a new status code.
		// Log the error for operators; the client will see a truncated/malformed response.
		slog.Error("failed to encode JSON response", "error", err)
	}
}

// --- Success responses ---

// JSON writes a success response with the given status code and data.
func JSON(w http.ResponseWriter, statusCode int, data any) {
	write(w, statusCode, Envelope{
		Success:   true,
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

// OK writes a 200 response.
func OK(w http.ResponseWriter, message string, data any) {
	write(w, http.StatusOK, Envelope{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

// Created writes a 201 response.
func Created(w http.ResponseWriter, message string, data any) {
	write(w, http.StatusCreated, Envelope{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

// Accepted writes a 202 response.
func Accepted(w http.ResponseWriter, message string, data any) {
	write(w, http.StatusAccepted, Envelope{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

// NoContent writes a 204 response with no body.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// --- Error responses ---

// Err writes an error response from an *errors.Error.
// This is the primary way to send error responses, bridging the errors package.
//
//	if err != nil {
//	    response.Err(w, err)
//	    return
//	}
func Err(w http.ResponseWriter, err error) {
	apiErr := errors.From(err)
	if apiErr == nil {
		// nil error — shouldn't happen, but handle gracefully
		OK(w, "Success", nil)
		return
	}

	write(w, apiErr.StatusCode, Envelope{
		Success: false,
		Error: &ErrorBody{
			Code:    apiErr.Code,
			Message: apiErr.Message,
			Fields:  apiErr.Fields,
			Details: apiErr.Details,
		},
		Timestamp: time.Now().Unix(),
	})
}

// Error writes a custom error response.
func Error(w http.ResponseWriter, statusCode int, code, message string) {
	write(w, statusCode, Envelope{
		Success: false,
		Error: &ErrorBody{
			Code:    code,
			Message: message,
		},
		Timestamp: time.Now().Unix(),
	})
}

// ErrorWithFields writes an error response with field-level errors.
func ErrorWithFields(w http.ResponseWriter, statusCode int, code, message string, fields map[string]string) {
	write(w, statusCode, Envelope{
		Success: false,
		Error: &ErrorBody{
			Code:    code,
			Message: message,
			Fields:  fields,
		},
		Timestamp: time.Now().Unix(),
	})
}

// --- Convenience error helpers ---

// BadRequest writes a 400 response.
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, errors.CodeBadRequest, message)
}

// Unauthorized writes a 401 response.
func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, errors.CodeUnauthorized, message)
}

// Forbidden writes a 403 response.
func Forbidden(w http.ResponseWriter, message string) {
	Error(w, http.StatusForbidden, errors.CodeForbidden, message)
}

// NotFound writes a 404 response.
func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, errors.CodeNotFound, message)
}

// Conflict writes a 409 response.
func Conflict(w http.ResponseWriter, message string) {
	Error(w, http.StatusConflict, errors.CodeConflict, message)
}

// ValidationError writes a 422 response with field errors.
func ValidationError(w http.ResponseWriter, fields map[string]string) {
	ErrorWithFields(w, http.StatusUnprocessableEntity,
		errors.CodeValidation, "Validation failed", fields)
}

// TooManyRequests writes a 429 response.
func TooManyRequests(w http.ResponseWriter, message string) {
	Error(w, http.StatusTooManyRequests, errors.CodeRateLimited, message)
}

// InternalServerError writes a 500 response.
func InternalServerError(w http.ResponseWriter, message string) {
	Error(w, http.StatusInternalServerError, errors.CodeInternal, message)
}

// ServiceUnavailable writes a 503 response.
func ServiceUnavailable(w http.ResponseWriter, message string) {
	Error(w, http.StatusServiceUnavailable, errors.CodeServiceUnavailable, message)
}
