package request

import (
	"fmt"
	"net/mail"
	"net/url"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/KARTIKrocks/apikit/errors"
)

// regexpCache caches compiled regular expressions to avoid recompilation.
var regexpCache sync.Map // string → *regexp.Regexp

// uuidRegex matches standard UUID format (8-4-4-4-12 hex digits).
var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// e164Regex matches E.164 phone numbers: a leading '+', a nonzero first digit,
// and up to 15 digits total.
var e164Regex = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)

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
//	gte=N            — >= N (alias of min: length for string/slice/map, value for numbers)
//	lte=N            — <= N (alias of max)
//	gt=N             — strictly greater than N (length for string/slice/map, value for numbers)
//	lt=N             — strictly less than N
//	eq=X             — equals X (string/number value, or item count for slice/map)
//	ne=X             — does not equal X
//	len=N            — exact length (string/slice/map)
//	oneof=a b c      — value must be one of the listed values (space-separated)
//	alpha            — string contains only letters
//	alphanum         — string contains only letters and digits
//	numeric          — string contains only digits
//	uuid             — valid UUID format
//	e164             — valid E.164 phone number (e.g. +14155552671)
//	contains=X       — string contains substring X
//	startswith=X     — string starts with prefix X
//	endswith=X       — string ends with suffix X
//
// Field names in error output use the `json` tag name if present.
//
// Nested struct fields without a validate tag are recursed automatically; struct
// fields with a validate tag have only that tag checked. Struct elements of
// slices, arrays, and maps are always recursed (reported as e.g. items[0].name).
// No separate "dive" tag is needed. Cross-field rules (eqfield, required_with, …)
// are intentionally out of scope for the tag engine; use request.NewValidation()
// for cross-field logic.
//
// An unrecognized rule is a programmer error: it panics rather than silently
// passing, so typos like `e_164` surface immediately instead of letting
// invalid input through unvalidated.
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

		// An explicit "-" skips the field (and its elements) entirely.
		if tag == "-" {
			continue
		}

		// Apply this field's own rules (e.g. required/min/max on the field itself).
		if tag != "" {
			if msg := validateField(fieldVal, tag); msg != "" {
				fields[name] = msg
			}
		}

		// Recurse into struct elements of slices, arrays, and maps so their own
		// tags are validated (reported as e.g. items[0].name), without requiring
		// a separate "dive" tag.
		diveElements(fieldVal, fields, name)
	}
}

// diveElements validates the elements of slices, arrays, and maps whose
// elements are structs (or pointers to structs). Other kinds are ignored.
func diveElements(val reflect.Value, fields map[string]string, name string) {
	switch val.Kind() {
	case reflect.Slice, reflect.Array:
		for i := range val.Len() {
			diveValue(val.Index(i), fields, fmt.Sprintf("%s[%d]", name, i))
		}
	case reflect.Map:
		iter := val.MapRange()
		for iter.Next() {
			diveValue(iter.Value(), fields, fmt.Sprintf("%s[%v]", name, iter.Key().Interface()))
		}
	}
}

// diveValue recurses into a single element when it is a struct or a non-nil
// pointer to a struct, skipping time.Time.
func diveValue(elem reflect.Value, fields map[string]string, name string) {
	if elem.Kind() == reflect.Interface {
		elem = elem.Elem()
	}
	if elem.Kind() == reflect.Pointer {
		if elem.IsNil() {
			return
		}
		elem = elem.Elem()
	}
	if elem.Kind() == reflect.Struct && elem.Type().Name() != "Time" {
		validateStructFields(elem, fields, name)
	}
}

// jsonFieldName returns the display name for a struct field in validation errors.
// It checks the `form` tag first, then `json` tag, falling back to the Go field name.
func jsonFieldName(field reflect.StructField) string {
	if tag := field.Tag.Get("form"); tag != "" && tag != "-" {
		name, _, _ := strings.Cut(tag, ",")
		if name != "" {
			return name
		}
	}
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

// ruleFunc validates a value against a single rule, with param being the text
// after '=' (empty for param-less rules). It returns an error message or "".
type ruleFunc func(val reflect.Value, param string) string

// paramless adapts a rule that ignores its parameter to a ruleFunc.
func paramless(fn func(reflect.Value) string) ruleFunc {
	return func(val reflect.Value, _ string) string { return fn(val) }
}

// ruleFuncs is the dispatch table of supported rules. Adding a rule here is the
// only change needed to support a new tag; unknown tags are rejected by panic.
var ruleFuncs = map[string]ruleFunc{
	"required":   paramless(ruleRequired),
	"email":      paramless(ruleEmail),
	"url":        paramless(ruleURL),
	"min":        ruleMin,
	"max":        ruleMax,
	"gte":        ruleMin, // gte is min: >= N (length for string/slice/map)
	"lte":        ruleMax, // lte is max: <= N
	"gt":         ruleGt,
	"lt":         ruleLt,
	"eq":         ruleEq,
	"ne":         ruleNe,
	"len":        ruleLen,
	"oneof":      ruleOneOf,
	"alpha":      paramless(ruleAlpha),
	"alphanum":   paramless(ruleAlphanum),
	"numeric":    paramless(ruleNumeric),
	"uuid":       paramless(ruleUUID),
	"e164":       paramless(ruleE164),
	"contains":   ruleContains,
	"startswith": ruleStartsWith,
	"endswith":   ruleEndsWith,
}

// applyRule applies a single validation rule and returns an error message or "".
// An unrecognized rule is a programmer error and panics (see ValidateStruct).
func applyRule(val reflect.Value, rule string) string {
	key, param, _ := strings.Cut(rule, "=")
	fn, ok := ruleFuncs[key]
	if !ok {
		panic(fmt.Sprintf("apikit: unknown validate rule %q", key))
	}
	return fn(val, param)
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

// numericOrLen returns a comparable magnitude for a value: the numeric value
// for numbers, or the element/character count for strings, slices, arrays, and
// maps. ok is false for kinds with no meaningful magnitude (bool, struct, …).
func numericOrLen(val reflect.Value) (float64, bool) {
	switch val.Kind() {
	case reflect.String:
		return float64(len(val.String())), true
	case reflect.Slice, reflect.Map, reflect.Array:
		return float64(val.Len()), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(val.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(val.Uint()), true
	case reflect.Float32, reflect.Float64:
		return val.Float(), true
	}
	return 0, false
}

// lenAwareMsg picks a message variant based on whether the value is measured by
// length (string/collection) or by numeric value.
func lenAwareMsg(kind reflect.Kind, numMsg, strMsg, itemsMsg string) string {
	switch kind {
	case reflect.String:
		return strMsg
	case reflect.Slice, reflect.Map, reflect.Array:
		return itemsMsg
	default:
		return numMsg
	}
}

func ruleGt(val reflect.Value, param string) string {
	n, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return ""
	}
	mag, ok := numericOrLen(val)
	if !ok || mag > n {
		return ""
	}
	return lenAwareMsg(val.Kind(),
		fmt.Sprintf("must be greater than %s", param),
		fmt.Sprintf("must be more than %s characters", param),
		fmt.Sprintf("must have more than %s items", param))
}

func ruleLt(val reflect.Value, param string) string {
	n, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return ""
	}
	mag, ok := numericOrLen(val)
	if !ok || mag < n {
		return ""
	}
	return lenAwareMsg(val.Kind(),
		fmt.Sprintf("must be less than %s", param),
		fmt.Sprintf("must be fewer than %s characters", param),
		fmt.Sprintf("must have fewer than %s items", param))
}

// valueEquals reports whether val equals param. For strings it compares the
// string; for numbers and bools the value; for slices/arrays/maps the item
// count. ok is false when param cannot be parsed for the value's kind (treated
// as a no-op, mirroring min/max).
func valueEquals(val reflect.Value, param string) (equal bool, ok bool) {
	switch val.Kind() {
	case reflect.String:
		return val.String() == param, true
	case reflect.Bool:
		b, err := strconv.ParseBool(param)
		return err == nil && val.Bool() == b, err == nil
	case reflect.Slice, reflect.Map, reflect.Array:
		n, err := strconv.Atoi(param)
		return err == nil && val.Len() == n, err == nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(param, 10, 64)
		return err == nil && val.Int() == n, err == nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(param, 10, 64)
		return err == nil && val.Uint() == n, err == nil
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(param, 64)
		return err == nil && val.Float() == f, err == nil
	}
	return false, false
}

func ruleEq(val reflect.Value, param string) string {
	equal, ok := valueEquals(val, param)
	if !ok || equal {
		return ""
	}
	return lenAwareMsg(val.Kind(),
		fmt.Sprintf("must equal %s", param),
		fmt.Sprintf("must equal %q", param),
		fmt.Sprintf("must have exactly %s items", param))
}

func ruleNe(val reflect.Value, param string) string {
	equal, ok := valueEquals(val, param)
	if !ok || !equal {
		return ""
	}
	return lenAwareMsg(val.Kind(),
		fmt.Sprintf("must not equal %s", param),
		fmt.Sprintf("must not equal %q", param),
		fmt.Sprintf("must not have %s items", param))
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

func ruleE164(val reflect.Value) string {
	s := stringVal(val)
	if s == "" {
		return ""
	}
	if !IsValidE164(s) {
		return "must be a valid E.164 phone number"
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

// IsValidE164 checks whether s is a valid E.164 phone number
// (a leading '+' followed by up to 15 digits, first digit nonzero).
func IsValidE164(s string) bool {
	return e164Regex.MatchString(s)
}

// MatchesRegexp checks whether s matches the given regexp pattern.
// Compiled patterns are cached to avoid recompilation on repeated calls.
func MatchesRegexp(s, pattern string) bool {
	re, ok := regexpCache.Load(pattern)
	if !ok {
		compiled, err := regexp.Compile(pattern)
		if err != nil {
			return false
		}
		re, _ = regexpCache.LoadOrStore(pattern, compiled)
	}
	return re.(*regexp.Regexp).MatchString(s)
}
