package errors

import (
	"net/http"
	"sync"
)

// Standard error codes.
// Use these as the Code field in Error for consistency across your API.
const (
	// Client errors (4xx)
	CodeBadRequest          = "BAD_REQUEST"
	CodeUnauthorized        = "UNAUTHORIZED"
	CodeForbidden           = "FORBIDDEN"
	CodeNotFound            = "NOT_FOUND"
	CodeMethodNotAllowed    = "METHOD_NOT_ALLOWED"
	CodeConflict            = "CONFLICT"
	CodeGone                = "GONE"
	CodeValidation          = "VALIDATION_ERROR"
	CodeRateLimited         = "RATE_LIMITED"
	CodeRequestTooLarge     = "REQUEST_TOO_LARGE"
	CodeUnsupportedMedia    = "UNSUPPORTED_MEDIA_TYPE"
	CodeUnprocessable       = "UNPROCESSABLE_ENTITY"
	CodeTooManyRequests     = "TOO_MANY_REQUESTS"
	CodeInvalidCredentials  = "INVALID_CREDENTIALS"
	CodeTokenExpired        = "TOKEN_EXPIRED"
	CodeTokenInvalid        = "TOKEN_INVALID"
	CodeEmailNotVerified    = "EMAIL_NOT_VERIFIED"
	CodeInsufficientScope   = "INSUFFICIENT_SCOPE"
	CodeResourceLocked      = "RESOURCE_LOCKED"
	CodePreconditionFailed  = "PRECONDITION_FAILED"
	CodeIdempotencyConflict = "IDEMPOTENCY_CONFLICT"

	// Server errors (5xx)
	CodeInternal           = "INTERNAL_ERROR"
	CodeNotImplemented     = "NOT_IMPLEMENTED"
	CodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	CodeTimeout            = "TIMEOUT"
	CodeDatabaseError      = "DATABASE_ERROR"
	CodeExternalService    = "EXTERNAL_SERVICE_ERROR"

	// Business logic errors
	CodeInsufficientFunds   = "INSUFFICIENT_FUNDS"
	CodeQuotaExceeded       = "QUOTA_EXCEEDED"
	CodeOperationNotAllowed = "OPERATION_NOT_ALLOWED"
	CodeDuplicateEntry      = "DUPLICATE_ENTRY"
)

// codeToStatus maps error codes to default HTTP status codes.
// You can override the status code using WithStatus().
var codeStatusMap = map[string]int{
	CodeBadRequest:          http.StatusBadRequest,
	CodeUnauthorized:        http.StatusUnauthorized,
	CodeForbidden:           http.StatusForbidden,
	CodeNotFound:            http.StatusNotFound,
	CodeMethodNotAllowed:    http.StatusMethodNotAllowed,
	CodeConflict:            http.StatusConflict,
	CodeGone:                http.StatusGone,
	CodeValidation:          http.StatusUnprocessableEntity,
	CodeRateLimited:         http.StatusTooManyRequests,
	CodeRequestTooLarge:     http.StatusRequestEntityTooLarge,
	CodeUnsupportedMedia:    http.StatusUnsupportedMediaType,
	CodeUnprocessable:       http.StatusUnprocessableEntity,
	CodeTooManyRequests:     http.StatusTooManyRequests,
	CodeInvalidCredentials:  http.StatusUnauthorized,
	CodeTokenExpired:        http.StatusUnauthorized,
	CodeTokenInvalid:        http.StatusUnauthorized,
	CodeEmailNotVerified:    http.StatusForbidden,
	CodeInsufficientScope:   http.StatusForbidden,
	CodeResourceLocked:      http.StatusLocked,
	CodePreconditionFailed:  http.StatusPreconditionFailed,
	CodeIdempotencyConflict: http.StatusConflict,

	CodeInternal:           http.StatusInternalServerError,
	CodeNotImplemented:     http.StatusNotImplemented,
	CodeServiceUnavailable: http.StatusServiceUnavailable,
	CodeTimeout:            http.StatusGatewayTimeout,
	CodeDatabaseError:      http.StatusInternalServerError,
	CodeExternalService:    http.StatusBadGateway,

	CodeInsufficientFunds:   http.StatusPaymentRequired,
	CodeQuotaExceeded:       http.StatusTooManyRequests,
	CodeOperationNotAllowed: http.StatusForbidden,
	CodeDuplicateEntry:      http.StatusConflict,
}

// codeStatusMu protects codeStatusMap for concurrent access.
var codeStatusMu sync.RWMutex

// codeToStatus returns the HTTP status code for the given error code.
// Returns 500 if the code is not recognized.
func codeToStatus(code string) int {
	codeStatusMu.RLock()
	status, ok := codeStatusMap[code]
	codeStatusMu.RUnlock()
	if ok {
		return status
	}
	return http.StatusInternalServerError
}

// RegisterCode registers a custom error code with its default HTTP status.
// It is safe for concurrent use.
//
//	func init() {
//	    errors.RegisterCode("SUBSCRIPTION_EXPIRED", 402)
//	}
func RegisterCode(code string, status int) {
	codeStatusMu.Lock()
	codeStatusMap[code] = status
	codeStatusMu.Unlock()
}
