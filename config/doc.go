// Package config loads application configuration from environment variables,
// .env files, and JSON config files into typed Go structs.
//
// It uses struct tags to map fields to environment variables and supports
// default values, type coercion, and validation via the request package's
// existing tag validators.
//
// # Quick Start
//
//	type AppConfig struct {
//	    Host     string `env:"HOST"      default:"localhost"  validate:"required"`
//	    Port     int    `env:"PORT"      default:"8080"       validate:"required,min=1,max=65535"`
//	    Debug    bool   `env:"DEBUG"     default:"false"`
//	    DBUrl    string `env:"DB_URL"    validate:"required,url"`
//	    LogLevel string `env:"LOG_LEVEL" default:"info"       validate:"oneof=debug info warn error"`
//	}
//
//	var cfg AppConfig
//	if err := config.Load(&cfg); err != nil {
//	    log.Fatal(err)
//	}
//
// # Struct Tags
//
//   - env:"VAR_NAME"       — maps the field to the named environment variable
//   - default:"value"      — fallback if env var and config file both miss
//   - validate:"..."       — reuses request package validators (required, min, max, url, etc.)
//
// # Source Priority (high to low)
//
//  1. Environment variables (always win)
//  2. .env file (loaded into process env, doesn't override existing)
//  3. JSON config file (base config layer)
//  4. default:"..." tags (fallback)
//
// # Options
//
//	config.Load(&cfg,
//	    config.WithPrefix("APP"),           // APP_PORT instead of PORT
//	    config.WithEnvFile(".env"),          // Load .env file
//	    config.WithJSONFile("config.json"), // Load JSON config as base
//	    config.WithRequired(),              // Error if files don't exist
//	)
//
// # Supported Types
//
//   - string, bool, int/int8/16/32/64, uint variants, float32/64
//   - time.Duration (parses "5s", "1m30s")
//   - []string, []int (comma-separated: "a,b,c")
//   - Nested structs (flattened env: DB_HOST for field DB.Host)
//
// # Nested Structs
//
//	type Config struct {
//	    DB struct {
//	        Host string `env:"HOST" default:"localhost"`
//	        Port int    `env:"PORT" default:"5432"`
//	    }
//	}
//
//	// With prefix "APP": reads APP_DB_HOST, APP_DB_PORT
//	config.Load(&cfg, config.WithPrefix("APP"))
//
// # MustLoad
//
// For use in main() or init(), MustLoad panics on error:
//
//	var cfg AppConfig
//	config.MustLoad(&cfg, config.WithEnvFile(".env"))
package config
