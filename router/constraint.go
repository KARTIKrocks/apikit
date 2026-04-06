package router

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/KARTIKrocks/apikit/errors"
)

// ParamConstraint defines a validation rule for a single path parameter.
type ParamConstraint struct {
	Name       string            // path parameter name (must match {name} in the pattern)
	Validate   func(string) bool // returns true if the value is valid
	ErrMessage string            // message for the BadRequest error on failure
}

// ValidateParams wraps a HandlerFunc with path parameter validation.
// Constraints are checked in order before the handler is called.
// On the first failure, it returns errors.BadRequest with the constraint's ErrMessage.
func ValidateParams(fn HandlerFunc, constraints ...ParamConstraint) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		for _, c := range constraints {
			val := r.PathValue(c.Name)
			if !c.Validate(val) {
				return errors.BadRequest(c.ErrMessage)
			}
		}
		return fn(w, r)
	}
}

// Int returns a constraint that requires the parameter to be a valid integer.
func Int(name string) ParamConstraint {
	return ParamConstraint{
		Name: name,
		Validate: func(s string) bool {
			_, err := strconv.Atoi(s)
			return err == nil
		},
		ErrMessage: fmt.Sprintf("parameter %q must be an integer", name),
	}
}

// UUID returns a constraint that requires the parameter to be a valid UUID.
func UUID(name string) ParamConstraint {
	re := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	return ParamConstraint{
		Name:       name,
		Validate:   re.MatchString,
		ErrMessage: fmt.Sprintf("parameter %q must be a valid UUID", name),
	}
}

// Regex returns a constraint that requires the parameter to match the given regular expression.
// It panics if the pattern is not a valid regular expression.
func Regex(name string, pattern string) ParamConstraint {
	re := regexp.MustCompile(pattern)
	return ParamConstraint{
		Name:       name,
		Validate:   re.MatchString,
		ErrMessage: fmt.Sprintf("parameter %q has invalid format", name),
	}
}

// OneOf returns a constraint that requires the parameter to be one of the allowed values.
func OneOf(name string, values ...string) ParamConstraint {
	allowed := make(map[string]struct{}, len(values))
	for _, v := range values {
		allowed[v] = struct{}{}
	}
	return ParamConstraint{
		Name: name,
		Validate: func(s string) bool {
			_, ok := allowed[s]
			return ok
		},
		ErrMessage: fmt.Sprintf("parameter %q must be one of: %s", name, strings.Join(values, ", ")),
	}
}
