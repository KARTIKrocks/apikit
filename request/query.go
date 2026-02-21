package request

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/KARTIKrocks/apikit/errors"
)

// Query provides typed access to URL query parameters.
// Query values are parsed once and cached for efficient repeated access.
type Query struct {
	values url.Values
}

// QueryFrom creates a Query helper from an HTTP request.
func QueryFrom(r *http.Request) *Query {
	return &Query{values: r.URL.Query()}
}

// String returns a query parameter as a string, or the default value.
func (q *Query) String(key, defaultVal string) string {
	val := q.values.Get(key)
	if val == "" {
		return defaultVal
	}
	return val
}

// StringRequired returns a query parameter or an error if missing.
func (q *Query) StringRequired(key string) (string, error) {
	val := q.values.Get(key)
	if val == "" {
		return "", errors.BadRequest(fmt.Sprintf("Query parameter %q is required", key)).
			WithField(key, "required")
	}
	return val, nil
}

// Int returns a query parameter as an int, or the default value.
func (q *Query) Int(key string, defaultVal int) (int, error) {
	val := q.values.Get(key)
	if val == "" {
		return defaultVal, nil
	}

	n, err := strconv.Atoi(val)
	if err != nil {
		return 0, errors.BadRequest(fmt.Sprintf("Query parameter %q must be an integer", key)).
			WithField(key, "must be an integer")
	}
	return n, nil
}

// IntRange returns an int query parameter clamped to [min, max].
func (q *Query) IntRange(key string, defaultVal, min, max int) (int, error) {
	n, err := q.Int(key, defaultVal)
	if err != nil {
		return 0, err
	}

	if n < min {
		n = min
	}
	if n > max {
		n = max
	}
	return n, nil
}

// Int64 returns a query parameter as an int64.
func (q *Query) Int64(key string, defaultVal int64) (int64, error) {
	val := q.values.Get(key)
	if val == "" {
		return defaultVal, nil
	}

	n, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, errors.BadRequest(fmt.Sprintf("Query parameter %q must be an integer", key)).
			WithField(key, "must be an integer")
	}
	return n, nil
}

// Float64 returns a query parameter as a float64.
func (q *Query) Float64(key string, defaultVal float64) (float64, error) {
	val := q.values.Get(key)
	if val == "" {
		return defaultVal, nil
	}

	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0, errors.BadRequest(fmt.Sprintf("Query parameter %q must be a number", key)).
			WithField(key, "must be a number")
	}
	return f, nil
}

// Bool returns a query parameter as a bool.
// Accepts: true, false, 1, 0, yes, no.
func (q *Query) Bool(key string, defaultVal bool) (bool, error) {
	val := q.values.Get(key)
	if val == "" {
		return defaultVal, nil
	}

	switch strings.ToLower(val) {
	case "true", "1", "yes":
		return true, nil
	case "false", "0", "no":
		return false, nil
	default:
		return false, errors.BadRequest(
			fmt.Sprintf("Query parameter %q must be a boolean (true/false)", key)).
			WithField(key, "must be true or false")
	}
}

// Time parses a query parameter as a time.Time using the given layout.
func (q *Query) Time(key, layout string) (time.Time, error) {
	val := q.values.Get(key)
	if val == "" {
		return time.Time{}, nil
	}

	t, err := time.Parse(layout, val)
	if err != nil {
		return time.Time{}, errors.BadRequest(
			fmt.Sprintf("Query parameter %q must be a valid date/time", key)).
			WithField(key, fmt.Sprintf("expected format: %s", layout))
	}
	return t, nil
}

// StringSlice returns a query parameter as a string slice.
// Splits by comma by default: ?tags=a,b,c → ["a", "b", "c"]
// Also supports repeated params: ?tag=a&tag=b → ["a", "b"]
func (q *Query) StringSlice(key string) []string {
	// First try repeated params
	values := q.values[key]
	if len(values) > 1 {
		return values
	}

	// Then try comma-separated
	val := q.values.Get(key)
	if val == "" {
		return nil
	}

	parts := strings.Split(val, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// Has checks if a query parameter exists (even if empty).
func (q *Query) Has(key string) bool {
	return q.values.Has(key)
}

// --- Standalone helpers (for when you don't want to create a Query object) ---

// QueryString returns a query parameter as a string.
func QueryString(r *http.Request, key, defaultVal string) string {
	return QueryFrom(r).String(key, defaultVal)
}

// QueryInt returns a query parameter as an int.
func QueryInt(r *http.Request, key string, defaultVal int) (int, error) {
	return QueryFrom(r).Int(key, defaultVal)
}

// QueryInt64 returns a query parameter as an int64.
func QueryInt64(r *http.Request, key string, defaultVal int64) (int64, error) {
	return QueryFrom(r).Int64(key, defaultVal)
}

// QueryBool returns a query parameter as a bool.
func QueryBool(r *http.Request, key string, defaultVal bool) (bool, error) {
	return QueryFrom(r).Bool(key, defaultVal)
}
