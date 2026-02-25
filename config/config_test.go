package config

import (
	stderrors "errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/KARTIKrocks/apikit/errors"
)

// --- helpers ---

func setEnv(t *testing.T, key, value string) {
	t.Helper()
	t.Setenv(key, value)
}

func tmpFile(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

// --- Load basic ---

func TestLoad_BasicStringAndInt(t *testing.T) {
	type cfg struct {
		Host string `env:"TEST_HOST"`
		Port int    `env:"TEST_PORT"`
	}

	setEnv(t, "TEST_HOST", "example.com")
	setEnv(t, "TEST_PORT", "9090")

	var c cfg
	if err := Load(&c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Host != "example.com" {
		t.Errorf("Host = %q, want %q", c.Host, "example.com")
	}
	if c.Port != 9090 {
		t.Errorf("Port = %d, want %d", c.Port, 9090)
	}
}

func TestLoad_DefaultValues(t *testing.T) {
	type cfg struct {
		Host string `env:"TEST_DEF_HOST" default:"localhost"`
		Port int    `env:"TEST_DEF_PORT" default:"8080"`
	}

	var c cfg
	if err := Load(&c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Host != "localhost" {
		t.Errorf("Host = %q, want %q", c.Host, "localhost")
	}
	if c.Port != 8080 {
		t.Errorf("Port = %d, want %d", c.Port, 8080)
	}
}

func TestLoad_EnvOverridesDefault(t *testing.T) {
	type cfg struct {
		Port int `env:"TEST_OVR_PORT" default:"8080"`
	}

	setEnv(t, "TEST_OVR_PORT", "3000")

	var c cfg
	if err := Load(&c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Port != 3000 {
		t.Errorf("Port = %d, want %d", c.Port, 3000)
	}
}

// --- Type parsing ---

func TestLoad_BoolField(t *testing.T) {
	type cfg struct {
		Debug bool `env:"TEST_DEBUG"`
	}

	setEnv(t, "TEST_DEBUG", "true")

	var c cfg
	if err := Load(&c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !c.Debug {
		t.Error("Debug = false, want true")
	}
}

func TestLoad_FloatField(t *testing.T) {
	type cfg struct {
		Rate float64 `env:"TEST_RATE"`
	}

	setEnv(t, "TEST_RATE", "3.14")

	var c cfg
	if err := Load(&c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Rate != 3.14 {
		t.Errorf("Rate = %f, want %f", c.Rate, 3.14)
	}
}

func TestLoad_UintField(t *testing.T) {
	type cfg struct {
		Count uint `env:"TEST_COUNT"`
	}

	setEnv(t, "TEST_COUNT", "42")

	var c cfg
	if err := Load(&c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Count != 42 {
		t.Errorf("Count = %d, want %d", c.Count, 42)
	}
}

func TestLoad_Duration(t *testing.T) {
	type cfg struct {
		Timeout time.Duration `env:"TEST_TIMEOUT"`
	}

	setEnv(t, "TEST_TIMEOUT", "5s")

	var c cfg
	if err := Load(&c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Timeout != 5*time.Second {
		t.Errorf("Timeout = %v, want %v", c.Timeout, 5*time.Second)
	}
}

func TestLoad_SliceString(t *testing.T) {
	type cfg struct {
		Tags []string `env:"TEST_TAGS"`
	}

	setEnv(t, "TEST_TAGS", "a,b,c")

	var c cfg
	if err := Load(&c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.Tags) != 3 || c.Tags[0] != "a" || c.Tags[1] != "b" || c.Tags[2] != "c" {
		t.Errorf("Tags = %v, want [a b c]", c.Tags)
	}
}

func TestLoad_SliceInt(t *testing.T) {
	type cfg struct {
		Ports []int `env:"TEST_PORTS"`
	}

	setEnv(t, "TEST_PORTS", "80,443,8080")

	var c cfg
	if err := Load(&c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.Ports) != 3 || c.Ports[0] != 80 || c.Ports[1] != 443 || c.Ports[2] != 8080 {
		t.Errorf("Ports = %v, want [80 443 8080]", c.Ports)
	}
}

// --- Prefix ---

func TestLoad_WithPrefix(t *testing.T) {
	type cfg struct {
		Host string `env:"HOST"`
		Port int    `env:"PORT"`
	}

	setEnv(t, "MYAPP_HOST", "prefixed.com")
	setEnv(t, "MYAPP_PORT", "5000")

	var c cfg
	if err := Load(&c, WithPrefix("MYAPP")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Host != "prefixed.com" {
		t.Errorf("Host = %q, want %q", c.Host, "prefixed.com")
	}
	if c.Port != 5000 {
		t.Errorf("Port = %d, want %d", c.Port, 5000)
	}
}

// --- Nested structs ---

func TestLoad_NestedStruct(t *testing.T) {
	type cfg struct {
		DB struct {
			Host string `env:"HOST" default:"localhost"`
			Port int    `env:"PORT" default:"5432"`
		}
	}

	setEnv(t, "DB_HOST", "dbserver")

	var c cfg
	if err := Load(&c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.DB.Host != "dbserver" {
		t.Errorf("DB.Host = %q, want %q", c.DB.Host, "dbserver")
	}
	if c.DB.Port != 5432 {
		t.Errorf("DB.Port = %d, want %d", c.DB.Port, 5432)
	}
}

func TestLoad_NestedStructWithPrefix(t *testing.T) {
	type cfg struct {
		DB struct {
			Host string `env:"HOST"`
		}
	}

	setEnv(t, "APP_DB_HOST", "nested-prefixed.com")

	var c cfg
	if err := Load(&c, WithPrefix("APP")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.DB.Host != "nested-prefixed.com" {
		t.Errorf("DB.Host = %q, want %q", c.DB.Host, "nested-prefixed.com")
	}
}

// --- .env file ---

func TestLoad_EnvFile(t *testing.T) {
	envContent := `# Comment
TEST_EF_HOST=envfile.com
TEST_EF_PORT=7070
TEST_EF_SECRET="my secret value"
`
	envPath := tmpFile(t, ".env", envContent)

	type cfg struct {
		Host   string `env:"TEST_EF_HOST"`
		Port   int    `env:"TEST_EF_PORT"`
		Secret string `env:"TEST_EF_SECRET"`
	}

	var c cfg
	if err := Load(&c, WithEnvFile(envPath)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Host != "envfile.com" {
		t.Errorf("Host = %q, want %q", c.Host, "envfile.com")
	}
	if c.Port != 7070 {
		t.Errorf("Port = %d, want %d", c.Port, 7070)
	}
	if c.Secret != "my secret value" {
		t.Errorf("Secret = %q, want %q", c.Secret, "my secret value")
	}
}

func TestLoad_EnvFileDoesNotOverrideExisting(t *testing.T) {
	envContent := `TEST_NOOVR_HOST=fromfile`
	envPath := tmpFile(t, ".env", envContent)

	setEnv(t, "TEST_NOOVR_HOST", "fromenv")

	type cfg struct {
		Host string `env:"TEST_NOOVR_HOST"`
	}

	var c cfg
	if err := Load(&c, WithEnvFile(envPath)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Host != "fromenv" {
		t.Errorf("Host = %q, want %q (env should override .env file)", c.Host, "fromenv")
	}
}

func TestLoad_EnvFileMissing_NotRequired(t *testing.T) {
	type cfg struct {
		Host string `env:"HOST" default:"fallback"`
	}

	var c cfg
	err := Load(&c, WithEnvFile("/nonexistent/.env"))
	if err != nil {
		t.Fatalf("expected no error for missing optional env file, got: %v", err)
	}
	if c.Host != "fallback" {
		t.Errorf("Host = %q, want %q", c.Host, "fallback")
	}
}

func TestLoad_EnvFileMissing_Required(t *testing.T) {
	type cfg struct {
		Host string `env:"HOST"`
	}

	var c cfg
	err := Load(&c, WithEnvFile("/nonexistent/.env"), WithRequired())
	if err == nil {
		t.Fatal("expected error for missing required env file")
	}
	if !strings.Contains(err.Error(), "file not found") {
		t.Errorf("error = %q, want it to contain 'file not found'", err.Error())
	}
}

// --- JSON file ---

func TestLoad_JSONFile(t *testing.T) {
	jsonContent := `{
		"host": "jsonhost.com",
		"port": 6060,
		"debug": true
	}`
	jsonPath := tmpFile(t, "config.json", jsonContent)

	type cfg struct {
		Host  string `env:"HOST"`
		Port  int    `env:"PORT"`
		Debug bool   `env:"DEBUG"`
	}

	var c cfg
	if err := Load(&c, WithJSONFile(jsonPath)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Host != "jsonhost.com" {
		t.Errorf("Host = %q, want %q", c.Host, "jsonhost.com")
	}
	if c.Port != 6060 {
		t.Errorf("Port = %d, want %d", c.Port, 6060)
	}
	if !c.Debug {
		t.Error("Debug = false, want true")
	}
}

func TestLoad_JSONFileNestedObject(t *testing.T) {
	jsonContent := `{
		"db": {
			"host": "dbhost",
			"port": 5432
		}
	}`
	jsonPath := tmpFile(t, "config.json", jsonContent)

	type cfg struct {
		DBHost string `env:"DB_HOST"`
		DBPort int    `env:"DB_PORT"`
	}

	var c cfg
	if err := Load(&c, WithJSONFile(jsonPath)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.DBHost != "dbhost" {
		t.Errorf("DBHost = %q, want %q", c.DBHost, "dbhost")
	}
	if c.DBPort != 5432 {
		t.Errorf("DBPort = %d, want %d", c.DBPort, 5432)
	}
}

func TestLoad_JSONFileArray(t *testing.T) {
	jsonContent := `{"tags": ["a", "b", "c"]}`
	jsonPath := tmpFile(t, "config.json", jsonContent)

	type cfg struct {
		Tags []string `env:"TAGS"`
	}

	var c cfg
	if err := Load(&c, WithJSONFile(jsonPath)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.Tags) != 3 || c.Tags[0] != "a" {
		t.Errorf("Tags = %v, want [a b c]", c.Tags)
	}
}

func TestLoad_JSONFileMissing_NotRequired(t *testing.T) {
	type cfg struct {
		Host string `env:"HOST" default:"fallback"`
	}

	var c cfg
	err := Load(&c, WithJSONFile("/nonexistent/config.json"))
	if err != nil {
		t.Fatalf("expected no error for missing optional JSON file, got: %v", err)
	}
}

func TestLoad_JSONFileMissing_Required(t *testing.T) {
	type cfg struct {
		Host string `env:"HOST"`
	}

	var c cfg
	err := Load(&c, WithJSONFile("/nonexistent/config.json"), WithRequired())
	if err == nil {
		t.Fatal("expected error for missing required JSON file")
	}
}

// --- Priority order ---

func TestLoad_PriorityOrder(t *testing.T) {
	// Env > .env file > JSON > default
	jsonContent := `{"host": "fromjson", "port": 1111, "debug": true, "level": "warn"}`
	jsonPath := tmpFile(t, "config.json", jsonContent)

	envContent := "TEST_PRIO_PORT=2222\nTEST_PRIO_DEBUG=false\n"
	envPath := tmpFile(t, ".env", envContent)

	// Real env var (highest priority).
	setEnv(t, "TEST_PRIO_PORT", "3333")

	type cfg struct {
		Host  string `env:"TEST_PRIO_HOST" default:"fromdefault"` // default wins (no env, no json match with prefix)
		Port  int    `env:"TEST_PRIO_PORT"`                       // real env wins over .env file
		Debug bool   `env:"TEST_PRIO_DEBUG"`                      // .env file wins over json
		Level string `env:"TEST_PRIO_LEVEL" default:"info"`       // default wins
	}

	var c cfg
	if err := Load(&c, WithJSONFile(jsonPath), WithEnvFile(envPath)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Host != "fromdefault" {
		t.Errorf("Host = %q, want %q", c.Host, "fromdefault")
	}
	if c.Port != 3333 {
		t.Errorf("Port = %d, want %d (real env should win)", c.Port, 3333)
	}
	if c.Debug != false {
		t.Errorf("Debug = %v, want false (.env should win over json)", c.Debug)
	}
	if c.Level != "info" {
		t.Errorf("Level = %q, want %q", c.Level, "info")
	}
}

// --- Validation ---

func TestLoad_ValidationError(t *testing.T) {
	type cfg struct {
		Host string `env:"TEST_VAL_HOST" validate:"required"`
	}

	var c cfg
	err := Load(&c)
	if err == nil {
		t.Fatal("expected validation error for missing required field")
	}

	var apiErr *errors.Error
	if !stderrors.As(err, &apiErr) {
		t.Fatalf("expected *errors.Error, got %T: %v", err, err)
	}
	if apiErr.Code != errors.CodeValidation {
		t.Errorf("code = %q, want %q", apiErr.Code, errors.CodeValidation)
	}
}

// --- Parse errors ---

func TestLoad_ParseErrorInt(t *testing.T) {
	type cfg struct {
		Port int `env:"TEST_PARSE_PORT"`
	}

	setEnv(t, "TEST_PARSE_PORT", "abc")

	var c cfg
	err := Load(&c)
	if err == nil {
		t.Fatal("expected parse error")
	}
	if !strings.Contains(err.Error(), "cannot parse") {
		t.Errorf("error = %q, want it to contain 'cannot parse'", err.Error())
	}
}

func TestLoad_ParseErrorBool(t *testing.T) {
	type cfg struct {
		Debug bool `env:"TEST_PARSE_BOOL"`
	}

	setEnv(t, "TEST_PARSE_BOOL", "notabool")

	var c cfg
	err := Load(&c)
	if err == nil {
		t.Fatal("expected parse error")
	}
	if !strings.Contains(err.Error(), "cannot parse") {
		t.Errorf("error = %q, want it to contain 'cannot parse'", err.Error())
	}
}

func TestLoad_ParseErrorDuration(t *testing.T) {
	type cfg struct {
		Timeout time.Duration `env:"TEST_PARSE_DUR"`
	}

	setEnv(t, "TEST_PARSE_DUR", "notaduration")

	var c cfg
	err := Load(&c)
	if err == nil {
		t.Fatal("expected parse error")
	}
	if !strings.Contains(err.Error(), "cannot parse") {
		t.Errorf("error = %q, want it to contain 'cannot parse'", err.Error())
	}
}

// --- Invalid dst ---

func TestLoad_NonPointerDst(t *testing.T) {
	type cfg struct {
		Host string `env:"HOST"`
	}

	var c cfg
	err := Load(c) // not a pointer
	if err == nil {
		t.Fatal("expected error for non-pointer dst")
	}
}

func TestLoad_NilPointerDst(t *testing.T) {
	type cfg struct {
		Host string `env:"HOST"`
	}

	err := Load((*cfg)(nil))
	if err == nil {
		t.Fatal("expected error for nil pointer dst")
	}
}

// --- MustLoad ---

func TestMustLoad_Panics(t *testing.T) {
	type cfg struct {
		Host string `env:"TEST_MUST_HOST" validate:"required"`
	}

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic from MustLoad")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("expected string panic, got %T", r)
		}
		if !strings.Contains(msg, "config:") {
			t.Errorf("panic = %q, want it to contain 'config:'", msg)
		}
	}()

	var c cfg
	MustLoad(&c)
}

func TestMustLoad_Success(t *testing.T) {
	type cfg struct {
		Host string `env:"TEST_MUSTSUC_HOST" default:"ok"`
	}

	var c cfg
	MustLoad(&c) // should not panic
	if c.Host != "ok" {
		t.Errorf("Host = %q, want %q", c.Host, "ok")
	}
}

// --- .env file parsing edge cases ---

func TestEnvFile_InlineComment(t *testing.T) {
	envContent := "TEST_IC_KEY=value # this is a comment\n"
	envPath := tmpFile(t, ".env", envContent)

	type cfg struct {
		Key string `env:"TEST_IC_KEY"`
	}

	var c cfg
	if err := Load(&c, WithEnvFile(envPath)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Key != "value" {
		t.Errorf("Key = %q, want %q", c.Key, "value")
	}
}

func TestEnvFile_SingleQuotes(t *testing.T) {
	envContent := "TEST_SQ_KEY='single quoted'\n"
	envPath := tmpFile(t, ".env", envContent)

	type cfg struct {
		Key string `env:"TEST_SQ_KEY"`
	}

	var c cfg
	if err := Load(&c, WithEnvFile(envPath)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Key != "single quoted" {
		t.Errorf("Key = %q, want %q", c.Key, "single quoted")
	}
}

func TestEnvFile_EmptyLinesAndComments(t *testing.T) {
	envContent := `
# Full line comment

TEST_ELAC_KEY=present

# Another comment
`
	envPath := tmpFile(t, ".env", envContent)

	type cfg struct {
		Key string `env:"TEST_ELAC_KEY"`
	}

	var c cfg
	if err := Load(&c, WithEnvFile(envPath)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Key != "present" {
		t.Errorf("Key = %q, want %q", c.Key, "present")
	}
}

// --- Int variants ---

func TestLoad_IntVariants(t *testing.T) {
	type cfg struct {
		I8  int8  `env:"TEST_IV_I8"`
		I16 int16 `env:"TEST_IV_I16"`
		I32 int32 `env:"TEST_IV_I32"`
		I64 int64 `env:"TEST_IV_I64"`
	}

	setEnv(t, "TEST_IV_I8", "127")
	setEnv(t, "TEST_IV_I16", "32000")
	setEnv(t, "TEST_IV_I32", "2000000")
	setEnv(t, "TEST_IV_I64", "9000000000")

	var c cfg
	if err := Load(&c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.I8 != 127 {
		t.Errorf("I8 = %d, want 127", c.I8)
	}
	if c.I16 != 32000 {
		t.Errorf("I16 = %d, want 32000", c.I16)
	}
	if c.I32 != 2000000 {
		t.Errorf("I32 = %d, want 2000000", c.I32)
	}
	if c.I64 != 9000000000 {
		t.Errorf("I64 = %d, want 9000000000", c.I64)
	}
}

func TestLoad_Float32(t *testing.T) {
	type cfg struct {
		Rate float32 `env:"TEST_F32_RATE"`
	}

	setEnv(t, "TEST_F32_RATE", "1.5")

	var c cfg
	if err := Load(&c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Rate != 1.5 {
		t.Errorf("Rate = %f, want 1.5", c.Rate)
	}
}

// --- JSON with env override ---

func TestLoad_EnvOverridesJSON(t *testing.T) {
	jsonContent := `{"host": "fromjson"}`
	jsonPath := tmpFile(t, "config.json", jsonContent)

	setEnv(t, "HOST", "fromenv")

	type cfg struct {
		Host string `env:"HOST"`
	}

	var c cfg
	if err := Load(&c, WithJSONFile(jsonPath)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Host != "fromenv" {
		t.Errorf("Host = %q, want %q (env should override json)", c.Host, "fromenv")
	}
}

// --- Invalid JSON ---

func TestLoad_InvalidJSON(t *testing.T) {
	jsonPath := tmpFile(t, "config.json", `{invalid json}`)

	type cfg struct {
		Host string `env:"HOST"`
	}

	var c cfg
	err := Load(&c, WithJSONFile(jsonPath))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "invalid JSON") {
		t.Errorf("error = %q, want it to contain 'invalid JSON'", err.Error())
	}
}

// --- No env tag field is skipped ---

func TestLoad_FieldWithoutEnvTagIsSkipped(t *testing.T) {
	type cfg struct {
		Internal string // No env tag.
		Host     string `env:"TEST_SKIP_HOST" default:"ok"`
	}

	var c cfg
	if err := Load(&c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Internal != "" {
		t.Errorf("Internal = %q, want empty", c.Internal)
	}
	if c.Host != "ok" {
		t.Errorf("Host = %q, want %q", c.Host, "ok")
	}
}

// --- .env file does not pollute process environment ---

func TestLoad_EnvFileDoesNotPolluteOsEnviron(t *testing.T) {
	envContent := "TEST_NOPOLLUTE_KEY=secretvalue\n"
	envPath := tmpFile(t, ".env", envContent)

	type cfg struct {
		Key string `env:"TEST_NOPOLLUTE_KEY"`
	}

	var c cfg
	if err := Load(&c, WithEnvFile(envPath)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Key != "secretvalue" {
		t.Errorf("Key = %q, want %q", c.Key, "secretvalue")
	}

	// The .env value must NOT appear in the real process environment.
	if val, ok := os.LookupEnv("TEST_NOPOLLUTE_KEY"); ok {
		t.Errorf("os.LookupEnv(TEST_NOPOLLUTE_KEY) = %q, want not set", val)
	}
}

// --- JSON + nested struct + prefix ---

func TestLoad_JSONNestedStructWithPrefix(t *testing.T) {
	jsonContent := `{
		"db": {
			"host": "jsondbhost",
			"port": 5432
		}
	}`
	jsonPath := tmpFile(t, "config.json", jsonContent)

	type cfg struct {
		DB struct {
			Host string `env:"HOST"`
			Port int    `env:"PORT"`
		}
	}

	var c cfg
	if err := Load(&c, WithJSONFile(jsonPath), WithPrefix("APP")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.DB.Host != "jsondbhost" {
		t.Errorf("DB.Host = %q, want %q", c.DB.Host, "jsondbhost")
	}
	if c.DB.Port != 5432 {
		t.Errorf("DB.Port = %d, want %d", c.DB.Port, 5432)
	}
}
