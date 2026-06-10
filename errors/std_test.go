package errors

import (
	"fmt"
	"testing"
)

// These cover the package-level re-exports so callers that import this package
// as `errors` (shadowing the stdlib) can use Is/As/Unwrap/Join directly,
// matching the package doc examples.

func TestIs(t *testing.T) {
	err := New(CodeNotFound, "user not found")

	if !Is(err, ErrNotFound) {
		t.Error("expected Is to match ErrNotFound sentinel")
	}
	if Is(err, ErrBadRequest) {
		t.Error("expected Is to NOT match ErrBadRequest sentinel")
	}

	// Matching should traverse the wrapped cause.
	wrapped := New(CodeInternal, "failed").Wrap(NotFound("user"))
	if !Is(wrapped, ErrNotFound) {
		t.Error("expected Is to find ErrNotFound in the wrapped chain")
	}
}

func TestAs(t *testing.T) {
	wrapped := fmt.Errorf("handler: %w", NotFound("user"))

	var apiErr *Error
	if !As(wrapped, &apiErr) {
		t.Fatal("expected As to extract *Error from wrapped error")
	}
	if apiErr.Code != CodeNotFound {
		t.Errorf("expected code %q, got %q", CodeNotFound, apiErr.Code)
	}
}

func TestUnwrapExport(t *testing.T) {
	cause := fmt.Errorf("db connection lost")
	err := New(CodeInternal, "failed").Wrap(cause)

	if Unwrap(err) != cause {
		t.Error("expected Unwrap to return the wrapped cause")
	}
}

func TestJoin(t *testing.T) {
	a := NotFound("user")
	b := New(CodeConflict, "duplicate")

	joined := Join(a, b)
	if !Is(joined, ErrNotFound) || !Is(joined, ErrConflict) {
		t.Error("expected Join result to match both joined errors")
	}
	if Join(nil, nil) != nil {
		t.Error("expected Join of all-nil to be nil")
	}
}
