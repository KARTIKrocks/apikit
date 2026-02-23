package request

import (
	"slices"
	"fmt"
	"net/mail"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/KARTIKrocks/apikit/errors"
)

// uuidRegex matches standard UUID format (8-4-4-4-12 hex digits).
var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// ValidateStruct validates a struct's fields using `validate` struct tags.
// It returns an *errors.Error (422) with field-level messages if validation fails,
// or nil if all rules pass.
//
// Supported tags (comma-separated in `validate:"..."`):
//
//	required         — value must not be zero
//	email            — valid email address (net/mail)
//	url              — valid URL (net/url)
//	min=N            — minimum length (string/slice/map) or value (int/float)
//	max=N            — maximum length (string/slice/map) or value (int/float)
//	len=N            — exact length (string/slice/map)
//	oneof=a b c      — value must be one of the listed values (space-separated)
//	alpha            — string contains only letters
//	alphanum         — string contains only letters and digits
//	numeric          — string contains only digits
//	uuid             — valid UUID format
//	contains=X       — string contains substring X
//	startswith=X     — string starts with prefix X
//	endswith=X       — string ends with suffix X
//
// Field names in error output use the `json` tag name if present.
func ValidateStruct(v any) error {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil
	}

	fields := make(map[string]string)
	validateStructFields(val, fields, "")

	if len(fields) == 0 {
		return nil
	}
	return errors.Validation("Validation failed", fields)
}

// validateStructFields recursively validates struct fields and collects errors.
func validateStructFields(val reflect.Value, fields map[string]string, prefix string) {
	t := val.Type()
	for i := range t.NumField() {
		field := t.Field(i)
		fieldVal := val.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		tag := field.Tag.Get("validate")

		// Resolve the display name from the json tag, falling back to field name
		name := jsonFieldName(field)
		if prefix != "" {
			name = prefix + "." + name
		}

		// Recurse into nested structs (but not time.Time, etc.)
		if fieldVal.Kind() == reflect.Struct && tag == "" && field.Type.Name() != "Time" {
			validateStructFields(fieldVal, fields, name)
			continue
		}
		if fieldVal.Kind() == reflect.Pointer && fieldVal.Type().Elem().Kind() == reflect.Struct && !fieldVal.IsNil() && tag == "" {
			validateStructFields(fieldVal.Elem(), fields, name)
			continue
		}

		if tag == "" || tag == "-" {
			continue
		}

		if msg := validateField(fieldVal, tag); msg != "" {
			fields[name] = msg
		}
	}
}

// jsonFieldName returns the JSON tag name for a struct field, or the field name if no tag.
func jsonFieldName(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "" || tag == "-" {
		return field.Name
	}
	name, _, _ := strings.Cut(tag, ",")
	if name == "" {
		return field.Name
	}
	return name
}

// validateField runs all comma-separated rules against a value, returning the
// first error message or empty string.
func validateField(val reflect.Value, tag string) string {
	rules := strings.Split(tag, ",")
	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}
		if msg := applyRule(val, rule); msg != "" {
			return msg
		}
	}
	return ""
}

// applyRule applies a single validation rule and returns an error message or "".
func applyRule(val reflect.Value, rule string) string {
	key, param, _ := strings.Cut(rule, "=")

	switch key {
	case "required":
		return ruleRequired(val)
	case "email":
		return ruleEmail(val)
	case "url":
		return ruleURL(val)
	case "min":
		return ruleMin(val, param)
	case "max":
		return ruleMax(val, param)
	case "len":
		return ruleLen(val, param)
	case "oneof":
		return ruleOneOf(val, param)
	case "alpha":
		return ruleAlpha(val)
	case "alphanum":
		return ruleAlphanum(val)
	case "numeric":
		return ruleNumeric(val)
	case "uuid":
		return ruleUUID(val)
	case "contains":
		return ruleContains(val, param)
	case "startswith":
		return ruleStartsWith(val, param)
	case "endswith":
		return ruleEndsWith(val, param)
	default:
		return ""
	}
}

func ruleRequired(val reflect.Value) string {
	if val.IsZero() {
		return "is required"
	}
	return ""
}

func ruleEmail(val reflect.Value) string {
	s := stringVal(val)
	if s == "" {
		return ""
	}
	if !IsValidEmail(s) {
		return "must be a valid email address"
	}
	return ""
}

func ruleURL(val reflect.Value) string {
	s := stringVal(val)
	if s == "" {
		return ""
	}
	if !IsValidURL(s) {
		return "must be a valid URL"
	}
	return ""
}

func ruleMin(val reflect.Value, param string) string {
	n, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return ""
	}
	switch val.Kind() {
	case reflect.String:
		if len(val.String()) < int(n) {
			return fmt.Sprintf("must be at least %s characters", param)
		}
	case reflect.Slice, reflect.Map, reflect.Array:
		if val.Len() < int(n) {
			return fmt.Sprintf("must have at least %s items", param)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(val.Int()) < n {
			return fmt.Sprintf("must be at least %s", param)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if float64(val.Uint()) < n {
			return fmt.Sprintf("must be at least %s", param)
		}
	case reflect.Float32, reflect.Float64:
		if val.Float() < n {
			return fmt.Sprintf("must be at least %s", param)
		}
	}
	return ""
}

func ruleMax(val reflect.Value, param string) string {
	n, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return ""
	}
	switch val.Kind() {
	case reflect.String:
		if len(val.String()) > int(n) {
			return fmt.Sprintf("must be at most %s characters", param)
		}
	case reflect.Slice, reflect.Map, reflect.Array:
		if val.Len() > int(n) {
			return fmt.Sprintf("must have at most %s items", param)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(val.Int()) > n {
			return fmt.Sprintf("must be at most %s", param)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if float64(val.Uint()) > n {
			return fmt.Sprintf("must be at most %s", param)
		}
	case reflect.Float32, reflect.Float64:
		if val.Float() > n {
			return fmt.Sprintf("must be at most %s", param)
		}
	}
	return ""
}

func ruleLen(val reflect.Value, param string) string {
	n, err := strconv.Atoi(param)
	if err != nil {
		return ""
	}
	switch val.Kind() {
	case reflect.String:
		if len(val.String()) != n {
			return fmt.Sprintf("must be exactly %s characters", param)
		}
	case reflect.Slice, reflect.Map, reflect.Array:
		if val.Len() != n {
			return fmt.Sprintf("must have exactly %s items", param)
		}
	}
	return ""
}

func ruleOneOf(val reflect.Value, param string) string {
	s := fmt.Sprintf("%v", val.Interface())
	allowed := strings.Fields(param)
	if slices.Contains(allowed, s) {
		return ""
	}
	return fmt.Sprintf("must be one of: %s", strings.Join(allowed, ", "))
}

func ruleAlpha(val reflect.Value) string {
	s := stringVal(val)
	if s == "" {
		return ""
	}
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return "must contain only letters"
		}
	}
	return ""
}

func ruleAlphanum(val reflect.Value) string {
	s := stringVal(val)
	if s == "" {
		return ""
	}
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return "must contain only letters and digits"
		}
	}
	return ""
}

func ruleNumeric(val reflect.Value) string {
	s := stringVal(val)
	if s == "" {
		return ""
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return "must contain only digits"
		}
	}
	return ""
}

func ruleUUID(val reflect.Value) string {
	s := stringVal(val)
	if s == "" {
		return ""
	}
	if !IsValidUUID(s) {
		return "must be a valid UUID"
	}
	return ""
}

func ruleContains(val reflect.Value, param string) string {
	s := stringVal(val)
	if s == "" {
		return ""
	}
	if !strings.Contains(s, param) {
		return fmt.Sprintf("must contain %q", param)
	}
	return ""
}

func ruleStartsWith(val reflect.Value, param string) string {
	s := stringVal(val)
	if s == "" {
		return ""
	}
	if !strings.HasPrefix(s, param) {
		return fmt.Sprintf("must start with %q", param)
	}
	return ""
}

func ruleEndsWith(val reflect.Value, param string) string {
	s := stringVal(val)
	if s == "" {
		return ""
	}
	if !strings.HasSuffix(s, param) {
		return fmt.Sprintf("must end with %q", param)
	}
	return ""
}

// stringVal extracts a string from a reflect.Value, returning "" for non-strings.
func stringVal(val reflect.Value) string {
	if val.Kind() == reflect.String {
		return val.String()
	}
	return ""
}

// --- Shared validation helpers (used by both tag engine and Validation builder) ---

// IsValidEmail checks whether s is a valid bare email address.
// It rejects display-name formats like "Alice <alice@example.com>".
func IsValidEmail(s string) bool {
	addr, err := mail.ParseAddress(s)
	if err != nil {
		return false
	}
	// Reject display-name format — only accept bare addresses.
	return addr.Address == s
}

// IsValidURL checks whether s is a valid absolute URL.
func IsValidURL(s string) bool {
	u, err := url.ParseRequestURI(s)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// IsValidUUID checks whether s matches the standard UUID format.
func IsValidUUID(s string) bool {
	return uuidRegex.MatchString(s)
}

// MatchesRegexp checks whether s matches the given regexp pattern.
func MatchesRegexp(s, pattern string) bool {
	matched, err := regexp.MatchString(pattern, s)
	return err == nil && matched
}
