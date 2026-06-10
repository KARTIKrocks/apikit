package errors

import stderrors "errors"

// Is reports whether any error in err's tree matches target.
//
// It re-exports the standard library's errors.Is so callers that import this
// package as `errors` (shadowing the stdlib) can still use it directly:
//
//	if errors.Is(err, errors.ErrNotFound) { ... }
//
// Because *Error implements Is (matching on Code) and Unwrap, sentinel checks
// and wrapped-cause traversal both work as expected.
func Is(err, target error) bool { return stderrors.Is(err, target) }

// As finds the first error in err's tree that matches target, and if one is
// found, sets target to that error value and returns true.
//
// It re-exports the standard library's errors.As:
//
//	var apiErr *errors.Error
//	if errors.As(err, &apiErr) { _ = apiErr.StatusCode }
func As(err error, target any) bool { return stderrors.As(err, target) }

// Unwrap returns the result of calling the Unwrap method on err, if err's type
// implements it; otherwise it returns nil. It re-exports the stdlib errors.Unwrap.
func Unwrap(err error) error { return stderrors.Unwrap(err) }

// Join returns an error that wraps the given errors, skipping any nil values.
// It re-exports the stdlib errors.Join.
func Join(errs ...error) error { return stderrors.Join(errs...) }
