package request

import (
	stderrors "errors"
	"testing"

	"github.com/KARTIKrocks/apikit/errors"
)

func assertValidationError(t *testing.T, err error, field, contains string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected validation error for field %q, got nil", field)
	}
	var apiErr *errors.Error
	if !stderrors.As(err, &apiErr) {
		t.Fatalf("expected *errors.Error, got %T", err)
	}
	if apiErr.Code != errors.CodeValidation {
		t.Fatalf("expected code %q, got %q", errors.CodeValidation, apiErr.Code)
	}
	msg, ok := apiErr.Fields[field]
	if !ok {
		t.Fatalf("expected field %q in errors, got fields: %v", field, apiErr.Fields)
	}
	if contains != "" && !containsStr(msg, contains) {
		t.Fatalf("field %q message %q does not contain %q", field, msg, contains)
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || findStr(s, sub))
}

func findStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

// --- Individual rules ---

func TestRequired(t *testing.T) {
	type S struct {
		Name string `json:"name" validate:"required"`
	}
	assertValidationError(t, ValidateStruct(S{}), "name", "required")
	assertNoError(t, ValidateStruct(S{Name: "ok"}))
}

func TestRequiredInt(t *testing.T) {
	type S struct {
		Age int `json:"age" validate:"required"`
	}
	assertValidationError(t, ValidateStruct(S{}), "age", "required")
	assertNoError(t, ValidateStruct(S{Age: 25}))
}

func TestEmail(t *testing.T) {
	type S struct {
		Email string `json:"email" validate:"email"`
	}
	assertValidationError(t, ValidateStruct(S{Email: "bad"}), "email", "valid email")
	assertNoError(t, ValidateStruct(S{Email: "user@example.com"}))
	// Empty is OK (not required)
	assertNoError(t, ValidateStruct(S{}))
}

func TestURL(t *testing.T) {
	type S struct {
		URL string `json:"url" validate:"url"`
	}
	assertValidationError(t, ValidateStruct(S{URL: "not-a-url"}), "url", "valid URL")
	assertNoError(t, ValidateStruct(S{URL: "https://example.com/path"}))
	assertNoError(t, ValidateStruct(S{}))
}

func TestMinString(t *testing.T) {
	type S struct {
		Name string `json:"name" validate:"min=3"`
	}
	assertValidationError(t, ValidateStruct(S{Name: "ab"}), "name", "at least 3")
	assertNoError(t, ValidateStruct(S{Name: "abc"}))
}

func TestMinInt(t *testing.T) {
	type S struct {
		Age int `json:"age" validate:"min=18"`
	}
	assertValidationError(t, ValidateStruct(S{Age: 10}), "age", "at least 18")
	assertNoError(t, ValidateStruct(S{Age: 18}))
}

func TestMaxString(t *testing.T) {
	type S struct {
		Name string `json:"name" validate:"max=5"`
	}
	assertValidationError(t, ValidateStruct(S{Name: "toolong"}), "name", "at most 5")
	assertNoError(t, ValidateStruct(S{Name: "ok"}))
}

func TestMaxInt(t *testing.T) {
	type S struct {
		Age int `json:"age" validate:"max=100"`
	}
	assertValidationError(t, ValidateStruct(S{Age: 150}), "age", "at most 100")
	assertNoError(t, ValidateStruct(S{Age: 50}))
}

func TestLen(t *testing.T) {
	type S struct {
		Code string `json:"code" validate:"len=4"`
	}
	assertValidationError(t, ValidateStruct(S{Code: "abc"}), "code", "exactly 4")
	assertValidationError(t, ValidateStruct(S{Code: "abcde"}), "code", "exactly 4")
	assertNoError(t, ValidateStruct(S{Code: "abcd"}))
}

func TestOneOf(t *testing.T) {
	type S struct {
		Role string `json:"role" validate:"oneof=admin user mod"`
	}
	assertValidationError(t, ValidateStruct(S{Role: "guest"}), "role", "one of")
	assertNoError(t, ValidateStruct(S{Role: "admin"}))
}

func TestAlpha(t *testing.T) {
	type S struct {
		Name string `json:"name" validate:"alpha"`
	}
	assertValidationError(t, ValidateStruct(S{Name: "abc123"}), "name", "only letters")
	assertNoError(t, ValidateStruct(S{Name: "abc"}))
	assertNoError(t, ValidateStruct(S{}))
}

func TestAlphanum(t *testing.T) {
	type S struct {
		Code string `json:"code" validate:"alphanum"`
	}
	assertValidationError(t, ValidateStruct(S{Code: "abc-123"}), "code", "letters and digits")
	assertNoError(t, ValidateStruct(S{Code: "abc123"}))
}

func TestNumeric(t *testing.T) {
	type S struct {
		PIN string `json:"pin" validate:"numeric"`
	}
	assertValidationError(t, ValidateStruct(S{PIN: "12ab"}), "pin", "only digits")
	assertNoError(t, ValidateStruct(S{PIN: "1234"}))
}

func TestUUID(t *testing.T) {
	type S struct {
		ID string `json:"id" validate:"uuid"`
	}
	assertValidationError(t, ValidateStruct(S{ID: "not-a-uuid"}), "id", "valid UUID")
	assertNoError(t, ValidateStruct(S{ID: "550e8400-e29b-41d4-a716-446655440000"}))
	assertNoError(t, ValidateStruct(S{}))
}

func TestContains(t *testing.T) {
	type S struct {
		Bio string `json:"bio" validate:"contains=go"`
	}
	assertValidationError(t, ValidateStruct(S{Bio: "I love rust"}), "bio", "contain")
	assertNoError(t, ValidateStruct(S{Bio: "I love go"}))
}

func TestStartsWith(t *testing.T) {
	type S struct {
		URL string `json:"url" validate:"startswith=https://"`
	}
	assertValidationError(t, ValidateStruct(S{URL: "http://example.com"}), "url", "start with")
	assertNoError(t, ValidateStruct(S{URL: "https://example.com"}))
}

func TestEndsWith(t *testing.T) {
	type S struct {
		File string `json:"file" validate:"endswith=.go"`
	}
	assertValidationError(t, ValidateStruct(S{File: "main.rs"}), "file", "end with")
	assertNoError(t, ValidateStruct(S{File: "main.go"}))
}

// --- Combined tags ---

func TestCombinedTags(t *testing.T) {
	type S struct {
		Email string `json:"email" validate:"required,email"`
	}
	// Empty fails on required first
	assertValidationError(t, ValidateStruct(S{}), "email", "required")
	// Invalid email
	assertValidationError(t, ValidateStruct(S{Email: "bad"}), "email", "valid email")
	assertNoError(t, ValidateStruct(S{Email: "a@b.com"}))
}

func TestCombinedMinMax(t *testing.T) {
	type S struct {
		Name string `json:"name" validate:"required,min=2,max=10"`
	}
	assertValidationError(t, ValidateStruct(S{}), "name", "required")
	assertValidationError(t, ValidateStruct(S{Name: "a"}), "name", "at least 2")
	assertValidationError(t, ValidateStruct(S{Name: "this is way too long"}), "name", "at most 10")
	assertNoError(t, ValidateStruct(S{Name: "Alice"}))
}

// --- Nested structs ---

func TestNestedStruct(t *testing.T) {
	type Address struct {
		City string `json:"city" validate:"required"`
	}
	type User struct {
		Name    string  `json:"name" validate:"required"`
		Address Address `json:"address"`
	}

	err := ValidateStruct(User{})
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *errors.Error
	if !stderrors.As(err, &apiErr) {
		t.Fatal("expected *errors.Error")
	}
	if _, ok := apiErr.Fields["name"]; !ok {
		t.Error("expected 'name' field error")
	}
	if _, ok := apiErr.Fields["address.city"]; !ok {
		t.Errorf("expected 'address.city' field error, got fields: %v", apiErr.Fields)
	}
}

// --- JSON tag name resolution ---

func TestJSONTagName(t *testing.T) {
	type S struct {
		FirstName string `json:"first_name" validate:"required"`
	}
	assertValidationError(t, ValidateStruct(S{}), "first_name", "required")
}

func TestJSONTagOmitempty(t *testing.T) {
	type S struct {
		Name string `json:"name,omitempty" validate:"required"`
	}
	assertValidationError(t, ValidateStruct(S{}), "name", "required")
}

func TestNoJSONTag(t *testing.T) {
	type S struct {
		Name string `validate:"required"`
	}
	assertValidationError(t, ValidateStruct(S{}), "Name", "required")
}

// --- Pointer input ---

func TestPointerInput(t *testing.T) {
	type S struct {
		Name string `json:"name" validate:"required"`
	}
	s := &S{}
	assertValidationError(t, ValidateStruct(s), "name", "required")
	assertNoError(t, ValidateStruct(&S{Name: "ok"}))
}

func TestNilPointer(t *testing.T) {
	assertNoError(t, ValidateStruct((*struct{})(nil)))
}

func TestNonStruct(t *testing.T) {
	assertNoError(t, ValidateStruct("not a struct"))
}

// --- Skip tag ---

func TestSkipTag(t *testing.T) {
	type S struct {
		Internal string `json:"-" validate:"-"`
	}
	assertNoError(t, ValidateStruct(S{}))
}

// --- Slice min/max ---

func TestSliceMinMax(t *testing.T) {
	type S struct {
		Tags []string `json:"tags" validate:"min=1,max=3"`
	}
	assertValidationError(t, ValidateStruct(S{Tags: []string{}}), "tags", "at least 1")
	assertValidationError(t, ValidateStruct(S{Tags: []string{"a", "b", "c", "d"}}), "tags", "at most 3")
	assertNoError(t, ValidateStruct(S{Tags: []string{"a", "b"}}))
}

// --- Float min/max ---

func TestFloatMinMax(t *testing.T) {
	type S struct {
		Score float64 `json:"score" validate:"min=0,max=100"`
	}
	assertValidationError(t, ValidateStruct(S{Score: -1}), "score", "at least 0")
	assertValidationError(t, ValidateStruct(S{Score: 101}), "score", "at most 100")
	assertNoError(t, ValidateStruct(S{Score: 50.5}))
}

// --- Error format integration ---

func TestErrorFormat(t *testing.T) {
	type S struct {
		Email string `json:"email" validate:"required,email"`
	}
	err := ValidateStruct(S{})
	var apiErr *errors.Error
	if !stderrors.As(err, &apiErr) {
		t.Fatal("expected *errors.Error")
	}
	if apiErr.StatusCode != 422 {
		t.Errorf("expected status 422, got %d", apiErr.StatusCode)
	}
	if apiErr.Code != "VALIDATION_ERROR" {
		t.Errorf("expected code VALIDATION_ERROR, got %s", apiErr.Code)
	}
}
