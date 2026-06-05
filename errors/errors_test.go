package errors

import (
	stderrors "errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	err := New(CodeNotFound, "user not found")

	if err.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", err.StatusCode)
	}
	if err.Code != CodeNotFound {
		t.Errorf("expected code %q, got %q", CodeNotFound, err.Code)
	}
	if err.Message != "user not found" {
		t.Errorf("expected message %q, got %q", "user not found", err.Message)
	}
	if err.Stack == "" || err.Stack == "unknown" {
		t.Error("expected stack trace to be captured")
	}
}

func TestNewf(t *testing.T) {
	err := Newf(CodeNotFound, "user %s not found", "abc123")

	if err.Message != "user abc123 not found" {
		t.Errorf("expected formatted message, got %q", err.Message)
	}
}

func TestErrorString(t *testing.T) {
	err := New(CodeNotFound, "user not found")
	expected := "NOT_FOUND: user not found"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}

	// With wrapped error
	err = New(CodeInternal, "failed").Wrap(fmt.Errorf("db connection lost"))
	if !strings.Contains(err.Error(), "db connection lost") {
		t.Errorf("expected wrapped error in string, got %q", err.Error())
	}
}

func TestErrorIs(t *testing.T) {
	err := New(CodeNotFound, "user not found")

	if !stderrors.Is(err, ErrNotFound) {
		t.Error("expected err to match ErrNotFound sentinel")
	}

	if stderrors.Is(err, ErrBadRequest) {
		t.Error("expected err to NOT match ErrBadRequest sentinel")
	}
}

func TestErrorAs(t *testing.T) {
	original := NotFound("user")
	wrapped := fmt.Errorf("handler: %w", original)

	var apiErr *Error
	if !stderrors.As(wrapped, &apiErr) {
		t.Fatal("expected to extract *Error from wrapped error")
	}

	if apiErr.Code != CodeNotFound {
		t.Errorf("expected code %q, got %q", CodeNotFound, apiErr.Code)
	}
}

func TestFrom(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		if From(nil) != nil {
			t.Error("expected nil for nil input")
		}
	})

	t.Run("standard error", func(t *testing.T) {
		stdErr := fmt.Errorf("something went wrong")
		apiErr := From(stdErr)

		if apiErr.StatusCode != 500 {
			t.Errorf("expected status 500, got %d", apiErr.StatusCode)
		}
		if apiErr.Code != CodeInternal {
			t.Errorf("expected code %q, got %q", CodeInternal, apiErr.Code)
		}
		if apiErr.Err != stdErr {
			t.Error("expected original error to be wrapped")
		}
	})

	t.Run("existing API error", func(t *testing.T) {
		original := NotFound("user")
		result := From(original)

		if result != original {
			t.Error("expected same pointer for existing *Error")
		}
	})
}

func TestWithFields(t *testing.T) {
	err := Validation("invalid input", nil).
		WithField("email", "is required").
		WithField("name", "too short")

	if len(err.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(err.Fields))
	}
	if err.Fields["email"] != "is required" {
		t.Errorf("unexpected email error: %q", err.Fields["email"])
	}
}

func TestWithDetails(t *testing.T) {
	err := Internal("something failed").
		WithDetail("trace_id", "abc123").
		WithDetail("retry_after", 30)

	if len(err.Details) != 2 {
		t.Errorf("expected 2 details, got %d", len(err.Details))
	}
}

func TestHTTPStatus(t *testing.T) {
	tests := []struct {
		err    error
		expect int
	}{
		{NotFound("x"), 404},
		{BadRequest("x"), 400},
		{fmt.Errorf("plain error"), 500},
		{fmt.Errorf("wrap: %w", NotFound("x")), 404},
	}

	for _, tt := range tests {
		if got := HTTPStatus(tt.err); got != tt.expect {
			t.Errorf("HTTPStatus(%v) = %d, want %d", tt.err, got, tt.expect)
		}
	}
}

func TestErrorCode(t *testing.T) {
	if code := ErrorCode(NotFound("x")); code != CodeNotFound {
		t.Errorf("expected %q, got %q", CodeNotFound, code)
	}
	if code := ErrorCode(fmt.Errorf("plain")); code != CodeInternal {
		t.Errorf("expected %q, got %q", CodeInternal, code)
	}
}

func TestRegisterCode(t *testing.T) {
	RegisterCode("CUSTOM_ERROR", 418)
	err := New("CUSTOM_ERROR", "I'm a teapot")

	if err.StatusCode != 418 {
		t.Errorf("expected status 418, got %d", err.StatusCode)
	}
}

func TestConstructors(t *testing.T) {
	tests := []struct {
		name       string
		err        *Error
		wantStatus int
		wantCode   string
	}{
		{"BadRequest", BadRequest("bad"), 400, CodeBadRequest},
		{"Unauthorized", Unauthorized("unauth"), 401, CodeUnauthorized},
		{"Forbidden", Forbidden("forbidden"), 403, CodeForbidden},
		{"NotFound", NotFound("user"), 404, CodeNotFound},
		{"Conflict", Conflict("conflict"), 409, CodeConflict},
		{"Validation", Validation("invalid", nil), 422, CodeValidation},
		{"RateLimited", RateLimited("slow down"), 429, CodeRateLimited},
		{"Internal", Internal("oops"), 500, CodeInternal},
		{"ServiceUnavailable", ServiceUnavailable("down"), 503, CodeServiceUnavailable},
		{"Timeout", Timeout("too slow"), 504, CodeTimeout},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.StatusCode != tt.wantStatus {
				t.Errorf("status: got %d, want %d", tt.err.StatusCode, tt.wantStatus)
			}
			if tt.err.Code != tt.wantCode {
				t.Errorf("code: got %q, want %q", tt.err.Code, tt.wantCode)
			}
		})
	}
}

func TestUnwrap(t *testing.T) {
	inner := fmt.Errorf("db error")
	err := Internal("failed").Wrap(inner)

	if stderrors.Unwrap(err) != inner {
		t.Error("expected Unwrap to return inner error")
	}
}
