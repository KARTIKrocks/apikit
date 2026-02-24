package request

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/KARTIKrocks/apikit/errors"
)

// BindForm decodes an application/x-www-form-urlencoded request body into T.
func BindForm[T any](r *http.Request) (T, error) {
	return BindFormWithConfig[T](r, getConfig())
}

// BindFormWithConfig decodes an application/x-www-form-urlencoded request body
// into T using the provided config.
func BindFormWithConfig[T any](r *http.Request, cfg Config) (T, error) {
	var v T
	if err := validateContentType(r, "application/x-www-form-urlencoded"); err != nil {
		return v, err
	}
	return bindFormWithConfig[T](r, cfg)
}

// BindMultipart decodes a multipart/form-data request body into T.
func BindMultipart[T any](r *http.Request) (T, error) {
	return BindMultipartWithConfig[T](r, getConfig())
}

// BindMultipartWithConfig decodes a multipart/form-data request body into T
// using the provided config.
func BindMultipartWithConfig[T any](r *http.Request, cfg Config) (T, error) {
	var v T
	if err := validateContentType(r, "multipart/form-data"); err != nil {
		return v, err
	}
	return bindMultipartWithConfig[T](r, cfg)
}

// FormFile returns the first file for the given form field name.
func FormFile(r *http.Request, field string) (*multipart.FileHeader, error) {
	_, fh, err := r.FormFile(field)
	if err != nil {
		return nil, errors.BadRequest(fmt.Sprintf("Missing or invalid file field %q", field))
	}
	return fh, nil
}

// FormFiles returns all uploaded files from a multipart request, keyed by field name.
func FormFiles(r *http.Request) map[string][]*multipart.FileHeader {
	if r.MultipartForm == nil {
		return nil
	}
	return r.MultipartForm.File
}

// bindFormWithConfig is the internal form binding implementation.
func bindFormWithConfig[T any](r *http.Request, cfg Config) (T, error) {
	var v T

	if r.Body == nil || r.Body == http.NoBody {
		return v, errors.BadRequest("Request body is required")
	}

	maxSize := cfg.MaxBodySize
	if maxSize <= 0 {
		maxSize = DefaultMaxBodySize
	}

	// Wrap the body with a size limiter
	r.Body = http.MaxBytesReader(nil, r.Body, maxSize)

	if err := r.ParseForm(); err != nil {
		if err.Error() == "http: request body too large" {
			return v, errors.New(errors.CodeRequestTooLarge, "Request body too large").
				WithStatus(http.StatusRequestEntityTooLarge)
		}
		return v, errors.BadRequest("Failed to parse form data")
	}

	if err := decodeFormValues(r.PostForm, &v); err != nil {
		return v, err
	}

	return runValidation(v)
}

// bindMultipartWithConfig is the internal multipart binding implementation.
func bindMultipartWithConfig[T any](r *http.Request, cfg Config) (T, error) {
	var v T

	if r.Body == nil || r.Body == http.NoBody {
		return v, errors.BadRequest("Request body is required")
	}

	maxMemory := cfg.MaxMultipartMemory
	if maxMemory <= 0 {
		maxMemory = DefaultMaxMultipartMemory
	}

	maxSize := cfg.MaxBodySize
	if maxSize <= 0 {
		maxSize = DefaultMaxBodySize
	}

	r.Body = http.MaxBytesReader(nil, r.Body, maxSize)

	if err := r.ParseMultipartForm(maxMemory); err != nil {
		if err.Error() == "http: request body too large" {
			return v, errors.New(errors.CodeRequestTooLarge, "Request body too large").
				WithStatus(http.StatusRequestEntityTooLarge)
		}
		return v, errors.BadRequest("Failed to parse multipart form data")
	}

	if err := decodeFormValues(r.PostForm, &v); err != nil {
		return v, err
	}

	return runValidation(v)
}

// runValidation runs struct tag validation and the Validator interface.
func runValidation[T any](v T) (T, error) {
	if err := ValidateStruct(&v); err != nil {
		return v, err
	}
	if validator, ok := any(&v).(Validator); ok {
		if err := validator.Validate(); err != nil {
			return v, err
		}
	}
	return v, nil
}

// decodeFormValues maps url.Values into a struct using `form` (then `json`) tags.
func decodeFormValues(values url.Values, dst any) error {
	val := reflect.ValueOf(dst)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return errors.Internal("decodeFormValues: dst must be a pointer to a struct")
	}

	t := val.Type()
	for i := range t.NumField() {
		field := t.Field(i)
		fieldVal := val.Field(i)

		if !field.IsExported() {
			continue
		}

		name := formFieldName(field)
		if name == "-" {
			continue
		}

		rawValues, ok := values[name]
		if !ok || len(rawValues) == 0 {
			continue
		}

		if err := setFieldValue(fieldVal, field, name, rawValues); err != nil {
			return err
		}
	}
	return nil
}

// formFieldName resolves the form field name: form tag → json tag → Go field name.
func formFieldName(field reflect.StructField) string {
	if tag := field.Tag.Get("form"); tag != "" {
		name, _, _ := strings.Cut(tag, ",")
		if name != "" {
			return name
		}
	}
	if tag := field.Tag.Get("json"); tag != "" {
		name, _, _ := strings.Cut(tag, ",")
		if name != "" {
			return name
		}
	}
	return field.Name
}

// setFieldValue converts raw string values and sets them on the struct field.
func setFieldValue(fieldVal reflect.Value, field reflect.StructField, key string, rawValues []string) error {
	raw := rawValues[0]

	switch fieldVal.Kind() {
	case reflect.String:
		fieldVal.SetString(raw)

	case reflect.Bool:
		b, err := parseBool(raw)
		if err != nil {
			return errors.BadRequest(fmt.Sprintf("Field %q: invalid boolean value", key)).
				WithField(key, "must be a boolean")
		}
		fieldVal.SetBool(b)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(raw, 10, field.Type.Bits())
		if err != nil {
			return errors.BadRequest(fmt.Sprintf("Field %q: invalid integer value", key)).
				WithField(key, "must be an integer")
		}
		fieldVal.SetInt(n)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(raw, 10, field.Type.Bits())
		if err != nil {
			return errors.BadRequest(fmt.Sprintf("Field %q: invalid unsigned integer value", key)).
				WithField(key, "must be a positive integer")
		}
		fieldVal.SetUint(n)

	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(raw, field.Type.Bits())
		if err != nil {
			return errors.BadRequest(fmt.Sprintf("Field %q: invalid number value", key)).
				WithField(key, "must be a number")
		}
		fieldVal.SetFloat(f)

	case reflect.Slice:
		if field.Type.Elem().Kind() == reflect.String {
			fieldVal.Set(reflect.ValueOf(rawValues))
		}
	}

	return nil
}

// parseBool extends strconv.ParseBool with HTML-specific values.
// Accepts everything strconv.ParseBool does plus "on"/"off" and "yes"/"no".
func parseBool(s string) (bool, error) {
	switch strings.ToLower(s) {
	case "on", "yes":
		return true, nil
	case "off", "no":
		return false, nil
	default:
		return strconv.ParseBool(s)
	}
}

