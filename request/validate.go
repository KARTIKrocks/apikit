package request

import (
	"github.com/KARTIKrocks/apikit/errors"
)

// Validator is the interface that request types can implement
// to perform self-validation after binding.
//
//	type CreateUserReq struct {
//	    Name  string `json:"name"`
//	    Email string `json:"email"`
//	}
//
//	func (r CreateUserReq) Validate() error {
//	    v := request.NewValidation()
//	    v.RequireString("name", r.Name)
//	    v.RequireEmail("email", r.Email)
//	    return v.Error()
//	}
type Validator interface {
	Validate() error
}

// Validation provides a fluent API for building field-level validation errors.
type Validation struct {
	fields map[string]string
}

// NewValidation creates a new Validation builder.
func NewValidation() *Validation {
	return &Validation{
		fields: make(map[string]string),
	}
}

// AddError adds a field error.
func (v *Validation) AddError(field, message string) {
	v.fields[field] = message
}

// HasErrors returns true if any validation errors have been added.
func (v *Validation) HasErrors() bool {
	return len(v.fields) > 0
}

// Error returns an *errors.Error if there are validation errors, or nil.
func (v *Validation) Error() error {
	if !v.HasErrors() {
		return nil
	}
	return errors.Validation("Validation failed", v.fields)
}

// Fields returns the validation error fields.
func (v *Validation) Fields() map[string]string {
	return v.fields
}

// --- Common validation rules ---

// RequireString validates that a string field is not empty.
func (v *Validation) RequireString(field, value string) {
	if value == "" {
		v.AddError(field, "is required")
	}
}

// MinLength validates minimum string length.
func (v *Validation) MinLength(field, value string, min int) {
	if len(value) < min {
		v.AddError(field, "is too short")
	}
}

// MaxLength validates maximum string length.
func (v *Validation) MaxLength(field, value string, max int) {
	if len(value) > max {
		v.AddError(field, "is too long")
	}
}

// RequireInt validates that an int field is not zero.
func (v *Validation) RequireInt(field string, value int) {
	if value == 0 {
		v.AddError(field, "is required")
	}
}

// Range validates that a number is within a range.
func (v *Validation) Range(field string, value, min, max int) {
	if value < min || value > max {
		v.AddError(field, "is out of range")
	}
}

// OneOf validates that a string is one of the allowed values.
func (v *Validation) OneOf(field, value string, allowed []string) {
	for _, a := range allowed {
		if value == a {
			return
		}
	}
	v.AddError(field, "is not a valid value")
}

// Custom allows adding a custom validation check.
func (v *Validation) Custom(field string, check func() bool, message string) {
	if !check() {
		v.AddError(field, message)
	}
}
