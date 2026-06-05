package errors

import "fmt"

// --- Client error constructors ---

// BadRequest creates a 400 Bad Request error.
func BadRequest(message string) *Error {
	return &Error{
		StatusCode: 400,
		Code:       CodeBadRequest,
		Message:    message,
		Stack:      caller(2),
	}
}

// Unauthorized creates a 401 Unauthorized error.
func Unauthorized(message string) *Error {
	return &Error{
		StatusCode: 401,
		Code:       CodeUnauthorized,
		Message:    message,
		Stack:      caller(2),
	}
}

// Forbidden creates a 403 Forbidden error.
func Forbidden(message string) *Error {
	return &Error{
		StatusCode: 403,
		Code:       CodeForbidden,
		Message:    message,
		Stack:      caller(2),
	}
}

// NotFound creates a 404 Not Found error.
// If resource is provided, the message will be "<resource> not found".
func NotFound(resource string) *Error {
	msg := "Resource not found"
	if resource != "" {
		msg = fmt.Sprintf("%s not found", resource)
	}
	return &Error{
		StatusCode: 404,
		Code:       CodeNotFound,
		Message:    msg,
		Stack:      caller(2),
	}
}

// Conflict creates a 409 Conflict error.
func Conflict(message string) *Error {
	return &Error{
		StatusCode: 409,
		Code:       CodeConflict,
		Message:    message,
		Stack:      caller(2),
	}
}

// Validation creates a 422 Validation error with field errors.
func Validation(message string, fields map[string]string) *Error {
	return &Error{
		StatusCode: 422,
		Code:       CodeValidation,
		Message:    message,
		Fields:     fields,
		Stack:      caller(2),
	}
}

// RateLimited creates a 429 Too Many Requests error.
func RateLimited(message string) *Error {
	return &Error{
		StatusCode: 429,
		Code:       CodeRateLimited,
		Message:    message,
		Stack:      caller(2),
	}
}

// --- Server error constructors ---

// Internal creates a 500 Internal Server Error.
// The message provided should be safe to expose to clients.
// Wrap the original error using .Wrap(err) for logging.
func Internal(message string) *Error {
	return &Error{
		StatusCode: 500,
		Code:       CodeInternal,
		Message:    message,
		Stack:      caller(2),
	}
}

// Internalf creates a 500 error wrapping the original error.
// The original error is NOT exposed to clients; only the message is.
func Internalf(err error, message string) *Error {
	return &Error{
		StatusCode: 500,
		Code:       CodeInternal,
		Message:    message,
		Err:        err,
		Stack:      caller(2),
	}
}

// ServiceUnavailable creates a 503 Service Unavailable error.
func ServiceUnavailable(message string) *Error {
	return &Error{
		StatusCode: 503,
		Code:       CodeServiceUnavailable,
		Message:    message,
		Stack:      caller(2),
	}
}

// Timeout creates a 504 Gateway Timeout error.
func Timeout(message string) *Error {
	return &Error{
		StatusCode: 504,
		Code:       CodeTimeout,
		Message:    message,
		Stack:      caller(2),
	}
}
