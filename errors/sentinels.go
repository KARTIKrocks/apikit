package errors

// Sentinel errors for use with errors.Is().
//
// Usage:
//
//	if errors.Is(err, apikit.ErrNotFound) {
//	    // handle not found
//	}
//
// These sentinels only match on Code, not on Message.
// This means New(CodeNotFound, "user not found") will match ErrNotFound.
var (
	ErrBadRequest       = &Error{Code: CodeBadRequest}
	ErrUnauthorized     = &Error{Code: CodeUnauthorized}
	ErrForbidden        = &Error{Code: CodeForbidden}
	ErrNotFound         = &Error{Code: CodeNotFound}
	ErrConflict         = &Error{Code: CodeConflict}
	ErrValidation       = &Error{Code: CodeValidation}
	ErrRateLimited      = &Error{Code: CodeRateLimited}
	ErrInternal         = &Error{Code: CodeInternal}
	ErrServiceUnavail   = &Error{Code: CodeServiceUnavailable}
	ErrTokenExpired     = &Error{Code: CodeTokenExpired}
	ErrTokenInvalid     = &Error{Code: CodeTokenInvalid}
	ErrInvalidCreds     = &Error{Code: CodeInvalidCredentials}
	ErrResourceLocked   = &Error{Code: CodeResourceLocked}
	ErrDuplicateEntry   = &Error{Code: CodeDuplicateEntry}
	ErrQuotaExceeded    = &Error{Code: CodeQuotaExceeded}
	ErrTimeout          = &Error{Code: CodeTimeout}
	ErrExternalService  = &Error{Code: CodeExternalService}
	ErrInsufficientFund = &Error{Code: CodeInsufficientFunds}
)
