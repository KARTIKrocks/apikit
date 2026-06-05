package config

import (
	stderrors "errors"
	"fmt"

	"github.com/KARTIKrocks/apikit/errors"
	"github.com/KARTIKrocks/apikit/request"
)

// Option configures the behavior of Load.
type Option func(*options)

type options struct {
	prefix   string
	envFile  string
	jsonFile string
	required bool
}

// WithPrefix sets a prefix for all environment variable lookups.
// For example, WithPrefix("APP") causes field with env:"PORT" to read APP_PORT.
func WithPrefix(prefix string) Option {
	return func(o *options) {
		o.prefix = prefix
	}
}

// WithEnvFile loads environment variables from a file (e.g., ".env").
// Values from the file do not override existing environment variables.
func WithEnvFile(path string) Option {
	return func(o *options) {
		o.envFile = path
	}
}

// WithJSONFile loads configuration from a JSON file as the base layer.
// Environment variables and .env values take precedence over JSON values.
func WithJSONFile(path string) Option {
	return func(o *options) {
		o.jsonFile = path
	}
}

// WithRequired causes Load to return an error if any specified files
// (env file or JSON file) do not exist.
func WithRequired() Option {
	return func(o *options) {
		o.required = true
	}
}

// Load populates dst from environment variables, optional .env files, and
// optional JSON config files, then validates the result using struct tags.
//
// dst must be a non-nil pointer to a struct.
//
// Sources are applied in priority order (high to low):
//  1. Environment variables
//  2. .env file values (do not override existing env vars)
//  3. JSON file values
//  4. default:"..." struct tags
func Load(dst any, opts ...Option) error {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	// Load JSON file (base layer).
	var jsonValues map[string]string
	if o.jsonFile != "" {
		var err error
		jsonValues, err = loadJSONFile(o.jsonFile, o.required)
		if err != nil {
			return err
		}
	}

	// Load .env file into a map (does not modify process environment).
	var envFileValues map[string]string
	if o.envFile != "" {
		var err error
		envFileValues, err = loadEnvFile(o.envFile, o.required)
		if err != nil {
			return err
		}
	}

	// Resolve all values into the struct.
	if err := resolve(dst, o.prefix, envFileValues, jsonValues); err != nil {
		return err
	}

	// Validate using request package tag validators.
	if err := request.ValidateStruct(dst); err != nil {
		// Wrap the validation error with config context.
		var apiErr *errors.Error
		if stderrors.As(err, &apiErr) {
			return errors.Validation("config: validation failed", apiErr.Fields)
		}
		return err
	}

	return nil
}

// MustLoad calls Load and panics if an error occurs.
// It is intended for use in main() or init() functions.
func MustLoad(dst any, opts ...Option) {
	if err := Load(dst, opts...); err != nil {
		panic(fmt.Sprintf("config: %v", err))
	}
}
