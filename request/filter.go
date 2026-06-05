package request

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/KARTIKrocks/apikit/errors"
)

// SortDirection represents ascending or descending sort.
type SortDirection string

// Sort direction values.
const (
	SortAsc  SortDirection = "asc"
	SortDesc SortDirection = "desc"
)

// SortField represents a single sort field and direction.
type SortField struct {
	Field     string
	Direction SortDirection
}

// SortConfig configures sort parameter parsing.
type SortConfig struct {
	// Param is the query parameter name. Default: "sort"
	Param string

	// AllowedFields is the allowlist of sortable fields.
	// If empty, all fields are allowed.
	AllowedFields []string

	// Default is the default sort if none specified.
	Default []SortField
}

// ParseSort parses sort parameters from the request.
//
// Format: ?sort=name,-created_at (prefix with - for descending)
//
//	sorts, err := request.ParseSort(r, request.SortConfig{
//	    AllowedFields: []string{"name", "created_at", "email"},
//	    Default: []request.SortField{{Field: "created_at", Direction: request.SortDesc}},
//	})
func ParseSort(r *http.Request, cfg SortConfig) ([]SortField, error) {
	param := cfg.Param
	if param == "" {
		param = "sort"
	}

	raw := r.URL.Query().Get(param)
	if raw == "" {
		return cfg.Default, nil
	}

	allowed := make(map[string]bool, len(cfg.AllowedFields))
	for _, f := range cfg.AllowedFields {
		allowed[f] = true
	}

	parts := strings.Split(raw, ",")
	fields := make([]SortField, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		direction := SortAsc
		field := part

		if strings.HasPrefix(part, "-") {
			direction = SortDesc
			field = part[1:]
		} else if strings.HasPrefix(part, "+") {
			field = part[1:]
		}

		if field == "" {
			continue
		}

		// Validate against allowlist
		if len(cfg.AllowedFields) > 0 && !allowed[field] {
			return nil, errors.BadRequest(
				fmt.Sprintf("Sort field %q is not allowed", field)).
				WithField("sort", fmt.Sprintf("allowed fields: %s", strings.Join(cfg.AllowedFields, ", ")))
		}

		fields = append(fields, SortField{
			Field:     field,
			Direction: direction,
		})
	}

	if len(fields) == 0 {
		return cfg.Default, nil
	}

	return fields, nil
}

// Filter represents a single filter condition.
type Filter struct {
	Field    string
	Operator string
	Value    string
}

// Standard filter operators.
const (
	FilterOpEq  = "eq"
	FilterOpNeq = "neq"
	FilterOpGt  = "gt"
	FilterOpGte = "gte"
	FilterOpLt  = "lt"
	FilterOpLte = "lte"
	FilterOpIn  = "in"
)

// FilterConfig configures filter parsing.
type FilterConfig struct {
	// AllowedFields is the allowlist of filterable fields.
	// If empty, all fields are allowed.
	AllowedFields []string

	// AllowedOperators restricts which operators can be used.
	// If empty, defaults to: eq, neq, gt, gte, lt, lte, in
	AllowedOperators []string
}

// ParseFilters parses filter parameters from the request.
//
// Supports two formats:
//
//	Simple:   ?filter[status]=active          → {Field: "status", Operator: "eq", Value: "active"}
//	Operator: ?filter[age][gte]=18            → {Field: "age", Operator: "gte", Value: "18"}
//
//	filters, err := request.ParseFilters(r, request.FilterConfig{
//	    AllowedFields: []string{"status", "age", "role"},
//	})
func ParseFilters(r *http.Request, cfg FilterConfig) ([]Filter, error) {
	allowed := make(map[string]bool, len(cfg.AllowedFields))
	for _, f := range cfg.AllowedFields {
		allowed[f] = true
	}

	defaultOps := []string{FilterOpEq, FilterOpNeq, FilterOpGt, FilterOpGte, FilterOpLt, FilterOpLte, FilterOpIn}
	ops := cfg.AllowedOperators
	if len(ops) == 0 {
		ops = defaultOps
	}
	allowedOps := make(map[string]bool, len(ops))
	for _, op := range ops {
		allowedOps[op] = true
	}

	var filters []Filter

	for key, values := range r.URL.Query() {
		if len(values) == 0 {
			continue
		}

		// Match filter[field] or filter[field][op]
		if !strings.HasPrefix(key, "filter[") {
			continue
		}

		// Remove "filter[" prefix
		rest := key[7:]

		field, operator, err := parseFilterKey(rest)
		if err != nil {
			return nil, err
		}

		// Validate field
		if len(cfg.AllowedFields) > 0 && !allowed[field] {
			return nil, errors.BadRequest(
				fmt.Sprintf("Filter field %q is not allowed", field)).
				WithField("filter", fmt.Sprintf("allowed fields: %s", strings.Join(cfg.AllowedFields, ", ")))
		}

		// Validate operator
		if !allowedOps[operator] {
			return nil, errors.BadRequest(
				fmt.Sprintf("Filter operator %q is not allowed for field %q", operator, field))
		}

		filters = append(filters, Filter{
			Field:    field,
			Operator: operator,
			Value:    values[0],
		})
	}

	return filters, nil
}

// parseFilterKey parses "field]" or "field][op]" and returns (field, operator).
func parseFilterKey(s string) (string, string, error) {
	// Case 1: filter[field] → "field]"
	if strings.Count(s, "]") == 1 && strings.HasSuffix(s, "]") {
		field := s[:len(s)-1]
		return field, FilterOpEq, nil
	}

	// Case 2: filter[field][op] → "field][op]"
	parts := strings.SplitN(s, "][", 2)
	if len(parts) == 2 && strings.HasSuffix(parts[1], "]") {
		field := parts[0]
		op := parts[1][:len(parts[1])-1]
		return field, op, nil
	}

	return "", "", errors.BadRequest("Malformed filter parameter")
}
