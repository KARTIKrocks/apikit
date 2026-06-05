package apitest

import (
	"strings"
	"testing"
)

// AssertStatus asserts that the response status code matches the expected value.
func (r *Response) AssertStatus(t *testing.T, expected int) {
	t.Helper()
	if r.StatusCode != expected {
		t.Errorf("expected status %d, got %d", expected, r.StatusCode)
	}
}

// AssertSuccess asserts that the response has a 2xx status code and success=true in the envelope.
func (r *Response) AssertSuccess(t *testing.T) {
	t.Helper()
	if r.StatusCode < 200 || r.StatusCode >= 300 {
		t.Errorf("expected 2xx status, got %d", r.StatusCode)
	}
	env, err := r.Envelope()
	if err != nil {
		t.Errorf("failed to decode envelope: %v", err)
		return
	}
	if !env.Success {
		t.Errorf("expected success=true, got false")
	}
}

// AssertError asserts that the response has success=false and the error code matches.
func (r *Response) AssertError(t *testing.T, code string) {
	t.Helper()
	env, err := r.Envelope()
	if err != nil {
		t.Errorf("failed to decode envelope: %v", err)
		return
	}
	if env.Success {
		t.Errorf("expected success=false, got true")
	}
	if env.Error == nil {
		t.Errorf("expected error body, got nil")
		return
	}
	if env.Error.Code != code {
		t.Errorf("expected error code %q, got %q", code, env.Error.Code)
	}
}

// AssertHeader asserts that the response header matches the expected value.
func (r *Response) AssertHeader(t *testing.T, key, expected string) {
	t.Helper()
	got := r.Header(key)
	if got != expected {
		t.Errorf("expected header %q = %q, got %q", key, expected, got)
	}
}

// AssertBodyContains asserts that the response body contains the given substring.
func (r *Response) AssertBodyContains(t *testing.T, substr string) {
	t.Helper()
	if !strings.Contains(string(r.Body), substr) {
		t.Errorf("expected body to contain %q, body: %s", substr, r.Body)
	}
}

// AssertValidationError asserts that the response contains a validation error
// with the given field name present in the error fields.
func (r *Response) AssertValidationError(t *testing.T, field string) {
	t.Helper()
	env, err := r.Envelope()
	if err != nil {
		t.Errorf("failed to decode envelope: %v", err)
		return
	}
	if env.Success {
		t.Errorf("expected success=false, got true")
	}
	if env.Error == nil {
		t.Errorf("expected error body, got nil")
		return
	}
	if env.Error.Fields == nil {
		t.Errorf("expected error fields, got nil")
		return
	}
	if _, ok := env.Error.Fields[field]; !ok {
		t.Errorf("expected field %q in error fields, got fields: %v", field, env.Error.Fields)
	}
}
